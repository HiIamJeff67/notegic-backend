package handlers

import (
	"context"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	matchers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/resolvers"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type RoutineHandlerInterface interface {
	HandleCreateRoutine(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateRoutine(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type RoutineHandler struct {
	db                   *gorm.DB
	validator            *validator.Validate
	patternResolver      resolvers.RoutineTaskPatternResolverInterface
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface
	routineRepository    repositories.RoutineRepositoryInterface
}

func NewRoutineHandler(
	db *gorm.DB,
	validatorInstance *validator.Validate,
	patternResolver resolvers.RoutineTaskPatternResolverInterface,
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface,
) RoutineHandlerInterface {
	if validatorInstance == nil {
		validatorInstance = validator.New()
	}
	if patternResolver == nil {
		patternResolver = resolvers.NewRoutineTaskPatternResolver(db)
	}
	if templateBlockMatcher == nil {
		templateBlockMatcher = matchers.NewRoutineTaskTemplateMatcher()
	}
	return &RoutineHandler{
		db:                   db,
		validator:            validatorInstance,
		patternResolver:      patternResolver,
		templateBlockMatcher: templateBlockMatcher,
		routineRepository:    repositories.NewRoutineRepository(scopes.NewRoutineScope()),
	}
}

func (s *RoutineHandler) HandleCreateRoutine(
	ctx context.Context,
	db *gorm.DB,
	tasks []schemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	candidateTaskIndexes := make([]int, 0, len(tasks))
	candidateTasks := make([]schemas.RoutineTask, 0, len(tasks))
	candidateActorUserIds := make([]uuid.UUID, 0, len(tasks))
	candidatePayloads := make([]croutinetasktypes.CreateRoutineRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.CreateRoutineRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		candidateTaskIndexes = append(candidateTaskIndexes, taskIndex)
		candidateTasks = append(candidateTasks, task)
		candidateActorUserIds = append(candidateActorUserIds, actorUserId)
		candidatePayloads = append(candidatePayloads, *payload)
		candidatePatterns = append(candidatePatterns, payload.Pattern)
	}
	if len(candidateTasks) == 0 {
		return successes, nil
	}

	patternValuesByCandidate, patternSuccesses, exception := s.patternResolver.ResolveMany(
		ctx,
		db,
		candidateTasks,
		candidateActorUserIds,
		candidatePatterns,
		allowedPermissions,
	)
	if exception != nil {
		return successes, exception
	}

	bulkInputs := make([]inputs.BulkCreateRoutineInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		title := s.templateBlockMatcher.MatchString(payload.Title, patternValues)
		description := s.templateBlockMatcher.MatchString(payload.Description, patternValues)
		bulkInputs = append(bulkInputs, inputs.BulkCreateRoutineInput{
			UserId:           candidateActorUserIds[candidateIndex],
			Id:               payload.Id,
			StationId:        payload.StationId,
			Title:            title,
			Description:      description,
			Status:           (*enums.RoutineStatus)(payload.Status),
			IsPinned:         payload.IsPinned,
			ScheduledStartAt: payload.ScheduledStartAt,
			ScheduledEndAt:   payload.ScheduledEndAt,
			Period:           (*enums.RoutinePeriod)(payload.Period),
			Timezone:         payload.Timezone,
		})
		taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
	}

	if len(bulkInputs) == 0 {
		return successes, nil
	}
	bulkSuccesses, exception := s.routineRepository.BulkCreateMany(
		bulkInputs,
		options.WithTransactionDB(db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}

	for index, success := range bulkSuccesses {
		successes[taskIndexes[index]] = success
	}

	return successes, nil
}

func (s *RoutineHandler) HandleUpdateRoutine(
	ctx context.Context,
	db *gorm.DB,
	tasks []schemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	candidateTaskIndexes := make([]int, 0, len(tasks))
	candidateTasks := make([]schemas.RoutineTask, 0, len(tasks))
	candidateActorUserIds := make([]uuid.UUID, 0, len(tasks))
	candidatePayloads := make([]croutinetasktypes.UpdateRoutineRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.UpdateRoutineRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		candidateTaskIndexes = append(candidateTaskIndexes, taskIndex)
		candidateTasks = append(candidateTasks, task)
		candidateActorUserIds = append(candidateActorUserIds, actorUserId)
		candidatePayloads = append(candidatePayloads, *payload)
		candidatePatterns = append(candidatePatterns, payload.Pattern)
	}
	if len(candidateTasks) == 0 {
		return successes, nil
	}

	patternValuesByCandidate, patternSuccesses, exception := s.patternResolver.ResolveMany(
		ctx,
		db,
		candidateTasks,
		candidateActorUserIds,
		candidatePatterns,
		allowedPermissions,
	)
	if exception != nil {
		return successes, exception
	}

	bulkInputs := make([]inputs.BulkUpdateRoutineInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		title := payload.Title
		if title != nil {
			matchedTitle := s.templateBlockMatcher.MatchString(*title, patternValues)
			title = &matchedTitle
		}
		description := payload.Description
		if description != nil {
			matchedDescription := s.templateBlockMatcher.MatchString(*description, patternValues)
			description = &matchedDescription
		}
		bulkInputs = append(bulkInputs, inputs.BulkUpdateRoutineInput{
			UserId: candidateActorUserIds[candidateIndex],
			Id:     payload.RoutineId,
			PartialUpdateInput: inputs.PartialUpdateRoutineInput{
				Values: inputs.UpdateRoutineInput{
					Title:            title,
					Description:      description,
					Status:           (*enums.RoutineStatus)(payload.Status),
					IsPinned:         payload.IsPinned,
					ScheduledStartAt: payload.ScheduledStartAt,
					ScheduledEndAt:   payload.ScheduledEndAt,
					Period:           (*enums.RoutinePeriod)(payload.Period),
					Timezone:         payload.Timezone,
				},
			},
		})
		taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
	}

	if len(bulkInputs) == 0 {
		return successes, nil
	}
	bulkSuccesses, exception := s.routineRepository.BulkUpdateMany(
		bulkInputs,
		options.WithTransactionDB(db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}

	for index, success := range bulkSuccesses {
		successes[taskIndexes[index]] = success
	}

	return successes, nil
}
