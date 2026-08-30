package resolvers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

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

type BlockPackPatternResolverInterface interface {
	Resolve(ctx context.Context, db *gorm.DB, actorUserId uuid.UUID, pattern croutinetasktypes.RoutineTaskPattern, allowedPermissions []cenums.AccessControlPermission) (map[string]string, *cexceptions.Exception)
	ResolveMany(ctx context.Context, db *gorm.DB, actorUserIds []uuid.UUID, patterns []croutinetasktypes.RoutineTaskPattern, allowedPermissions []cenums.AccessControlPermission) ([]map[string]string, []bool, *cexceptions.Exception)
}

type BlockPackPatternResolver struct {
	db                  *gorm.DB
	blockPackRepository srepositories.BlockPackRepositoryInterface
}

func NewBlockPackPatternResolver(db *gorm.DB) BlockPackPatternResolverInterface {
	return BlockPackPatternResolver{
		db:                  db,
		blockPackRepository: srepositories.NewBlockPackRepository(db, sscopes.NewBlockPackScope()),
	}
}

func (r BlockPackPatternResolver) Resolve(
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

func (r BlockPackPatternResolver) ResolveMany(
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

	checkInputs := make([]sinputs.BulkCheckBlockPackPermissionInput, 0)
	keysByUserAndBlockPackId := map[[2]uuid.UUID][]struct {
		taskIndex int
		key       string
	}{}

	for patternIndex, pattern := range patterns {
		for key, binding := range pattern {
			if binding.Source != PatternSourceBlockCheckboxCount {
				continue
			}
			if binding.BlockPackId == nil || *binding.BlockPackId == uuid.Nil {
				taskSuccesses[patternIndex] = false
				continue
			}
			mapKey := [2]uuid.UUID{actorUserIds[patternIndex], *binding.BlockPackId}
			if _, exists := keysByUserAndBlockPackId[mapKey]; !exists {
				checkInputs = append(checkInputs, sinputs.BulkCheckBlockPackPermissionInput{
					UserId: actorUserIds[patternIndex],
					Id:     *binding.BlockPackId,
				})
			}
			keysByUserAndBlockPackId[mapKey] = append(keysByUserAndBlockPackId[mapKey], struct {
				taskIndex int
				key       string
			}{taskIndex: patternIndex, key: key})
		}
	}
	if len(checkInputs) == 0 {
		return values, taskSuccesses, nil
	}
	if db == nil || r.blockPackRepository == nil {
		return nil, nil, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).
			WithOrigin(fmt.Errorf("block pack pattern source is not available"))
	}

	permissionSuccesses, _, exception := r.blockPackRepository.BulkCheckPermissionsAndGetManyByIds(
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

	validBlockPackIds := make([]uuid.UUID, 0, len(checkInputs))
	validBlockPackIdSet := map[uuid.UUID]bool{}
	for index, success := range permissionSuccesses {
		if !success {
			for _, request := range keysByUserAndBlockPackId[[2]uuid.UUID{checkInputs[index].UserId, checkInputs[index].Id}] {
				taskSuccesses[request.taskIndex] = false
			}
			continue
		}
		blockPackId := checkInputs[index].Id
		validBlockPackIds = append(validBlockPackIds, blockPackId)
		validBlockPackIdSet[blockPackId] = true
	}
	if len(validBlockPackIds) == 0 {
		return values, taskSuccesses, nil
	}

	var rows []struct {
		BlockPackId uuid.UUID `gorm:"column:block_pack_id"`
		Checked     bool      `gorm:"column:checked"`
	}
	if err := db.WithContext(ctx).
		Model(&sschemas.Block{}).
		Select(`block_pack_id, COALESCE((props->>'checked')::boolean, false) AS checked`).
		Where("block_pack_id IN ? AND type = ? AND deleted_at IS NULL", validBlockPackIds, cenums.BlockType_CheckListItem).
		Find(&rows).Error; err != nil {
		return nil, nil, cexceptions.New(
			"QueryFailed",
			"Block",
			"ResolvePattern",
			"Failed to retrieve block pattern values",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	totalByBlockPackId := map[uuid.UUID]int{}
	checkedByBlockPackId := map[uuid.UUID]int{}
	uncheckedByBlockPackId := map[uuid.UUID]int{}
	for _, row := range rows {
		totalByBlockPackId[row.BlockPackId]++
		if row.Checked {
			checkedByBlockPackId[row.BlockPackId]++
		} else {
			uncheckedByBlockPackId[row.BlockPackId]++
		}
	}

	for patternIndex, pattern := range patterns {
		for key, binding := range pattern {
			if binding.Source != PatternSourceBlockCheckboxCount || binding.BlockPackId == nil {
				continue
			}
			blockPackId := *binding.BlockPackId
			if !validBlockPackIdSet[blockPackId] {
				continue
			}

			count := totalByBlockPackId[blockPackId]
			if binding.Checked != nil {
				if *binding.Checked {
					count = checkedByBlockPackId[blockPackId]
				} else {
					count = uncheckedByBlockPackId[blockPackId]
				}
			}
			values[patternIndex][key] = strconv.Itoa(count)
		}
	}

	return values, taskSuccesses, nil
}
