package resolvers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

const (
	PatternSourceScheduledAt        = "scheduledAt"
	PatternSourceRecordId           = "recordId"
	PatternSourceShortRecordId      = "shortRecordId"
	PatternSourceRoutineTaskId      = "routineTaskId"
	PatternSourceBlockText          = "blockText"
	PatternSourceBlockCheckboxCount = "blockCheckboxCount"
)

type RoutineTaskPatternResolverInterface interface {
	Resolve(ctx context.Context, db *gorm.DB, task sschemas.RoutineTask, actorUserId uuid.UUID, pattern croutinetasktypes.RoutineTaskPattern, allowedPermissions []cenums.AccessControlPermission) (map[string]string, *cexceptions.Exception)
	ResolveMany(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, actorUserIds []uuid.UUID, patterns []croutinetasktypes.RoutineTaskPattern, allowedPermissions []cenums.AccessControlPermission) ([]map[string]string, []bool, *cexceptions.Exception)
}

type RoutineTaskPatternResolver struct {
	blockPatternResolver     BlockPatternResolverInterface
	blockPackPatternResolver BlockPackPatternResolverInterface
}

func NewRoutineTaskPatternResolver(db *gorm.DB) RoutineTaskPatternResolverInterface {
	return RoutineTaskPatternResolver{
		blockPatternResolver:     NewBlockPatternResolver(db),
		blockPackPatternResolver: NewBlockPackPatternResolver(db),
	}
}

func (r RoutineTaskPatternResolver) Resolve(
	ctx context.Context,
	db *gorm.DB,
	task sschemas.RoutineTask,
	actorUserId uuid.UUID,
	pattern croutinetasktypes.RoutineTaskPattern,
	allowedPermissions []cenums.AccessControlPermission,
) (map[string]string, *cexceptions.Exception) {
	values, successes, exception := r.ResolveMany(
		ctx,
		db,
		[]sschemas.RoutineTask{task},
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

func (r RoutineTaskPatternResolver) ResolveMany(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	actorUserIds []uuid.UUID,
	patterns []croutinetasktypes.RoutineTaskPattern,
	allowedPermissions []cenums.AccessControlPermission,
) ([]map[string]string, []bool, *cexceptions.Exception) {
	values := make([]map[string]string, len(patterns))
	successes := make([]bool, len(patterns))
	for index := range patterns {
		values[index] = make(map[string]string, len(patterns[index]))
		successes[index] = true
	}
	if len(tasks) != len(patterns) || len(actorUserIds) != len(patterns) {
		return nil, nil, cexceptions.New(
			"InvalidRoutineTaskPayload",
			"RoutineTask",
			"Resolve",
			"Routine task payload is invalid",
			http.StatusBadRequest,
		).
			WithOrigin(fmt.Errorf("tasks, actorUserIds and patterns length mismatch"))
	}

	hasBlockPatternSource := false
	hasBlockPackPatternSource := false
	for _, pattern := range patterns {
		for _, binding := range pattern {
			switch binding.Source {
			case PatternSourceBlockText:
				hasBlockPatternSource = true
			case PatternSourceBlockCheckboxCount:
				hasBlockPackPatternSource = true
			}
		}
	}

	if hasBlockPatternSource {
		blockValues, blockSuccesses, exception := r.blockPatternResolver.ResolveMany(
			ctx,
			db,
			actorUserIds,
			patterns,
			allowedPermissions,
		)
		if exception != nil {
			return nil, nil, exception
		}
		for index, success := range blockSuccesses {
			if !success {
				successes[index] = false
			}
			for key, value := range blockValues[index] {
				values[index][key] = value
			}
		}
	}

	if hasBlockPackPatternSource {
		blockPackValues, blockPackSuccesses, exception := r.blockPackPatternResolver.ResolveMany(
			ctx,
			db,
			actorUserIds,
			patterns,
			allowedPermissions,
		)
		if exception != nil {
			return nil, nil, exception
		}
		for index, success := range blockPackSuccesses {
			if !success {
				successes[index] = false
			}
			for key, value := range blockPackValues[index] {
				values[index][key] = value
			}
		}
	}

	for patternIndex, pattern := range patterns {
		for key, binding := range pattern {
			switch binding.Source {
			case PatternSourceScheduledAt:
				scheduledAt := tasks[patternIndex].RecordScheduledAt
				if scheduledAt.IsZero() {
					scheduledAt = tasks[patternIndex].ScheduledAt
				}
				if binding.Timezone != nil && *binding.Timezone != "" {
					location, err := time.LoadLocation(*binding.Timezone)
					if err != nil {
						successes[patternIndex] = false
						continue
					}
					scheduledAt = scheduledAt.In(location)
				}
				format := time.RFC3339
				if binding.Format != nil && *binding.Format != "" {
					format = *binding.Format
				}
				values[patternIndex][key] = scheduledAt.Format(format)

			case PatternSourceRecordId:
				values[patternIndex][key] = tasks[patternIndex].RecordId.String()

			case PatternSourceShortRecordId:
				recordId := tasks[patternIndex].RecordId.String()
				if len(recordId) > 8 {
					recordId = recordId[:8]
				}
				values[patternIndex][key] = recordId

			case PatternSourceRoutineTaskId:
				values[patternIndex][key] = tasks[patternIndex].Id.String()

			case PatternSourceBlockText, PatternSourceBlockCheckboxCount:
				continue

			default:
				successes[patternIndex] = false
			}
		}
	}

	return values, successes, nil
}
