package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"
)

type BlockPatternResolverInterface interface {
	Resolve(ctx context.Context, db *gorm.DB, actorUserId uuid.UUID, pattern croutinetasktypes.RoutineTaskPattern, allowedPermissions []cenums.AccessControlPermission) (map[string]string, *cexceptions.Exception)
	ResolveMany(ctx context.Context, db *gorm.DB, actorUserIds []uuid.UUID, patterns []croutinetasktypes.RoutineTaskPattern, allowedPermissions []cenums.AccessControlPermission) ([]map[string]string, []bool, *cexceptions.Exception)
}

type BlockPatternResolver struct {
	db              *gorm.DB
	blockRepository srepositories.BlockRepositoryInterface
}

func NewBlockPatternResolver(db *gorm.DB) BlockPatternResolverInterface {
	return BlockPatternResolver{
		db:              db,
		blockRepository: srepositories.NewBlockRepository(db, sscopes.NewBlockScope()),
	}
}

func (r BlockPatternResolver) Resolve(
	ctx context.Context,
	db *gorm.DB,
	actorUserId uuid.UUID,
	pattern croutinetasktypes.RoutineTaskPattern,
	allowedPermissions []cenums.AccessControlPermission,
) (map[string]string, *cexceptions.Exception) {
	values, successes, exception := r.ResolveMany(
		ctx,
		db,
		[]uuid.UUID{actorUserId},
		[]croutinetasktypes.RoutineTaskPattern{pattern},
		allowedPermissions,
	)
	if exception != nil {
		return nil, exception
	}
	if len(successes) == 0 || !successes[0] {
		return nil, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		)
	}
	return values[0], nil
}

func (r BlockPatternResolver) ResolveMany(
	ctx context.Context,
	db *gorm.DB,
	actorUserIds []uuid.UUID,
	patterns []croutinetasktypes.RoutineTaskPattern,
	allowedPermissions []cenums.AccessControlPermission,
) ([]map[string]string, []bool, *cexceptions.Exception) {
	values := make([]map[string]string, len(patterns))
	taskSuccesses := make([]bool, len(patterns))
	for index := range patterns {
		values[index] = map[string]string{}
		taskSuccesses[index] = true
	}
	if len(actorUserIds) != len(patterns) {
		return nil, nil, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).
			WithOrigin(fmt.Errorf("actorUserIds and patterns length mismatch"))
	}

	checkInputs := make([]sinputs.BulkCheckBlockPermissionInput, 0)
	keysByUserAndBlockId := map[[2]uuid.UUID][]struct {
		taskIndex int
		key       string
	}{}

	for patternIndex, pattern := range patterns {
		for key, binding := range pattern {
			if binding.Source != PatternSourceBlockText {
				continue
			}
			if binding.BlockId == nil || *binding.BlockId == uuid.Nil {
				taskSuccesses[patternIndex] = false
				continue
			}
			mapKey := [2]uuid.UUID{actorUserIds[patternIndex], *binding.BlockId}
			if _, exists := keysByUserAndBlockId[mapKey]; !exists {
				checkInputs = append(checkInputs, sinputs.BulkCheckBlockPermissionInput{
					UserId: actorUserIds[patternIndex],
					Id:     *binding.BlockId,
				})
			}
			keysByUserAndBlockId[mapKey] = append(keysByUserAndBlockId[mapKey], struct {
				taskIndex int
				key       string
			}{taskIndex: patternIndex, key: key})
		}
	}
	if len(checkInputs) == 0 {
		return values, taskSuccesses, nil
	}
	if db == nil || r.blockRepository == nil {
		return nil, nil, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).
			WithOrigin(fmt.Errorf("block pattern source is not available"))
	}

	permissionSuccesses, blocks, exception := r.blockRepository.BulkCheckPermissionsAndGetManyByIds(
		checkInputs,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(db.WithContext(ctx)),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return nil, nil, exception
	}

	blocksById := make(map[uuid.UUID]sschemas.Block, len(blocks))
	for _, block := range blocks {
		blocksById[block.Id] = block
	}
	for index, success := range permissionSuccesses {
		if !success {
			for _, request := range keysByUserAndBlockId[[2]uuid.UUID{checkInputs[index].UserId, checkInputs[index].Id}] {
				taskSuccesses[request.taskIndex] = false
			}
			continue
		}
		block := blocksById[checkInputs[index].Id]
		var content any
		if err := json.Unmarshal(block.Content, &content); err != nil {
			return nil, nil, cexceptions.New(
				"InvalidRoutineTaskPayload",
				"RoutineTask",
				"Resolve",
				"Routine task payload is invalid",
				http.StatusBadRequest,
			).WithOrigin(err)
		}
		parts := make([]string, 0)
		var walk func(any)
		walk = func(current any) {
			switch typed := current.(type) {
			case []any:
				for _, item := range typed {
					walk(item)
				}
			case map[string]any:
				if text, ok := typed["text"].(string); ok {
					parts = append(parts, text)
				}
				for _, value := range typed {
					walk(value)
				}
			}
		}
		walk(content)
		text := strings.Join(parts, "")
		for _, request := range keysByUserAndBlockId[[2]uuid.UUID{checkInputs[index].UserId, block.Id}] {
			values[request.taskIndex][request.key] = text
		}
	}

	return values, taskSuccesses, nil
}
