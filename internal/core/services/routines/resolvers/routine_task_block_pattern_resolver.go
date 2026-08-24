package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	routinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"

	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	coreenums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
)

type BlockPatternResolverInterface interface {
	Resolve(ctx context.Context, db *gorm.DB, actorUserId uuid.UUID, pattern routinetasktypes.RoutineTaskPattern, allowedPermissions []coreenums.AccessControlPermission) (map[string]string, *exceptions.Exception)
	ResolveMany(ctx context.Context, db *gorm.DB, actorUserIds []uuid.UUID, patterns []routinetasktypes.RoutineTaskPattern, allowedPermissions []coreenums.AccessControlPermission) ([]map[string]string, []bool, *exceptions.Exception)
}

type BlockPatternResolver struct {
	db              *gorm.DB
	blockRepository repositories.BlockRepositoryInterface
}

func NewBlockPatternResolver(db *gorm.DB) BlockPatternResolverInterface {
	return BlockPatternResolver{
		db:              db,
		blockRepository: repositories.NewBlockRepository(scopes.NewBlockScope()),
	}
}

func (r BlockPatternResolver) Resolve(
	ctx context.Context,
	db *gorm.DB,
	actorUserId uuid.UUID,
	pattern routinetasktypes.RoutineTaskPattern,
	allowedPermissions []coreenums.AccessControlPermission,
) (map[string]string, *exceptions.Exception) {
	values, successes, exception := r.ResolveMany(
		ctx,
		db,
		[]uuid.UUID{actorUserId},
		[]routinetasktypes.RoutineTaskPattern{pattern},
		allowedPermissions,
	)
	if exception != nil {
		return nil, exception
	}
	if len(successes) == 0 || !successes[0] {
		return nil, exceptions.New(
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
	patterns []routinetasktypes.RoutineTaskPattern,
	allowedPermissions []coreenums.AccessControlPermission,
) ([]map[string]string, []bool, *exceptions.Exception) {
	values := make([]map[string]string, len(patterns))
	taskSuccesses := make([]bool, len(patterns))
	for index := range patterns {
		values[index] = map[string]string{}
		taskSuccesses[index] = true
	}
	if len(actorUserIds) != len(patterns) {
		return nil, nil, exceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).
			WithOrigin(fmt.Errorf("actorUserIds and patterns length mismatch"))
	}

	checkInputs := make([]inputs.BulkCheckBlockPermissionInput, 0)
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
				checkInputs = append(checkInputs, inputs.BulkCheckBlockPermissionInput{
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
		return nil, nil, exceptions.New(
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
		options.WithTransactionDB(db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return nil, nil, exception
	}

	blocksById := make(map[uuid.UUID]schemas.Block, len(blocks))
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
			return nil, nil, exceptions.New(
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
