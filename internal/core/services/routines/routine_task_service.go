package routines

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tasks"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	stimes "github.com/HiIamJeff67/notegic-backend/shared/lib/times"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type RoutineTaskServiceInterface interface {
	GetMyRoutineTaskById(ctx context.Context, reqDto *capi.GetMyRoutineTaskByIdRequestDto) (*capi.GetMyRoutineTaskByIdResponseDto, *cexceptions.Exception)
	GetAllMyRoutineTasksByRoutineIds(ctx context.Context, reqDto *capi.GetAllMyRoutineTasksByRoutineIdsRequestDto) (*capi.GetAllMyRoutineTasksByRoutineIdsResponseDto, *cexceptions.Exception)
	GetAllMyRoutineTasks(ctx context.Context, reqDto *capi.GetAllMyRoutineTasksRequestDto) (*capi.GetAllMyRoutineTasksResponseDto, *cexceptions.Exception)
	CreateRoutineTaskByRoutineId(ctx context.Context, reqDto *capi.CreateRoutineTaskByRoutineIdRequestDto) (*capi.CreateRoutineTaskByRoutineIdResponseDto, *cexceptions.Exception)
	UpdateMyRoutineTaskById(ctx context.Context, reqDto *capi.UpdateMyRoutineTaskByIdRequestDto) (*capi.UpdateMyRoutineTaskByIdResponseDto, *cexceptions.Exception)
	PauseMyRoutineTaskById(ctx context.Context, reqDto *capi.PauseMyRoutineTaskByIdRequestDto) (*capi.PauseMyRoutineTaskByIdResponseDto, *cexceptions.Exception)
	ResumeMyRoutineTaskById(ctx context.Context, reqDto *capi.ResumeMyRoutineTaskByIdRequestDto) (*capi.ResumeMyRoutineTaskByIdResponseDto, *cexceptions.Exception)
	HardDeleteMyRoutineTaskById(ctx context.Context, reqDto *capi.HardDeleteMyRoutineTaskByIdRequestDto) (*capi.HardDeleteMyRoutineTaskByIdResponseDto, *cexceptions.Exception)
	HardDeleteMyRoutineTasksByIds(ctx context.Context, reqDto *capi.HardDeleteMyRoutineTasksByIdsRequestDto) (*capi.HardDeleteMyRoutineTasksByIdsResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineTaskStatusCount(ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskStatusCountRequestDto) (*capi.VisualizeMyRoutineTaskStatusCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineTaskPurposeCount(ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskPurposeCountRequestDto) (*capi.VisualizeMyRoutineTaskPurposeCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineTaskScheduledAtCount(ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskScheduledAtCountRequestDto) (*capi.VisualizeMyRoutineTaskScheduledAtCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineTaskActualStartedAtCount(ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskActualStartedAtCountRequestDto) (*capi.VisualizeMyRoutineTaskActualStartedAtCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineTaskActualEndedAtCount(ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskActualEndedAtCountRequestDto) (*capi.VisualizeMyRoutineTaskActualEndedAtCountResponseDto, *cexceptions.Exception)

	SearchPrivateRoutineTasks(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineTaskInput) (*cgqlmodels.SearchRoutineTaskConnection, *cexceptions.Exception)

	MarkCompletedRoutineTasks(ctx context.Context, eventId uuid.UUID, request *cdurablejob.MarkCompletedRoutineTasksRequestDto) *cexceptions.Exception
	MarkFailedRoutineTasks(ctx context.Context, eventId uuid.UUID, request *cdurablejob.MarkFailedRoutineTasksRequestDto) *cexceptions.Exception
}

type RoutineTaskService struct {
	validator                   *validator.Validate
	db                          *gorm.DB
	routineTaskScope            sscopes.RoutineTaskScopeInterface
	routineTaskRepository       srepositories.RoutineTaskRepositoryInterface
	routineTaskRecordRepository srepositories.RoutineTaskRecordRepositoryInterface
	userQuotaRepository         srepositories.UserQuotaRepositoryInterface
	routineTaskExecutionService RoutineTaskExecutionServiceInterface
}

func NewRoutineTaskService(
	validator *validator.Validate,
	db *gorm.DB,
	routineTaskScope sscopes.RoutineTaskScopeInterface,
	routineTaskRepository srepositories.RoutineTaskRepositoryInterface,
	routineTaskRecordRepository srepositories.RoutineTaskRecordRepositoryInterface,
	userQuotaRepository srepositories.UserQuotaRepositoryInterface,
	routineTaskExecutionServices ...RoutineTaskExecutionServiceInterface,
) RoutineTaskServiceInterface {
	if db == nil {
		db = data.DB
	}
	if routineTaskScope == nil {
		routineTaskScope = sscopes.NewRoutineTaskScope()
	}
	if routineTaskRecordRepository == nil {
		routineTaskRecordRepository = srepositories.NewRoutineTaskRecordRepository(
			db,
			sscopes.NewRoutineTaskRecordScope(),
		)
	}
	if userQuotaRepository == nil {
		userQuotaRepository = srepositories.NewUserQuotaRepository(db)
	}
	var routineTaskExecutionService RoutineTaskExecutionServiceInterface
	if len(routineTaskExecutionServices) > 0 {
		routineTaskExecutionService = routineTaskExecutionServices[0]
	}
	if routineTaskExecutionService == nil {
		routineTaskExecutionService = NewRoutineTaskExecutionService(validator, db, nil)
	}

	return &RoutineTaskService{
		validator:                   validator,
		db:                          db,
		routineTaskScope:            routineTaskScope,
		routineTaskRepository:       routineTaskRepository,
		routineTaskRecordRepository: routineTaskRecordRepository,
		userQuotaRepository:         userQuotaRepository,
		routineTaskExecutionService: routineTaskExecutionService,
	}
}

/* ============================== Auxiliary Functions ============================== */

func (s *RoutineTaskService) visualizeMyRoutineTaskTimeCount(
	ctx context.Context,
	userId uuid.UUID,
	permission cenums.AccessControlPermission,
	timeHourUnit int,
	queryRangeStartedAt time.Time,
	queryRangeEndedAt time.Time,
	columnName string,
	fieldName string,
) ([]capi.RoutineTaskCountDatum, *cexceptions.Exception) {
	db := s.db.WithContext(ctx)

	var buckets []struct {
		BucketStart      time.Time `gorm:"column:bucket_start;"`
		RoutineTaskCount int64     `gorm:"column:routine_task_count;"`
	}

	result := db.
		Table(
			`generate_series(
				date_trunc('hour', ?::timestamptz),
				date_trunc('hour', ?::timestamptz - interval '1 microsecond'),
				?::integer * interval '1 hour'
			) AS buckets(bucket_start)`,
			queryRangeStartedAt,
			queryRangeEndedAt,
			timeHourUnit,
		).
		Select(`
			buckets.bucket_start AS bucket_start,
			COUNT(uts.station_id) AS routine_task_count
		`).
		Joins(
			`LEFT JOIN "RoutineTaskTable" routine_task
				ON routine_task.`+columnName+` >= buckets.bucket_start
				AND routine_task.`+columnName+` < buckets.bucket_start + ?::integer * interval '1 hour'`,
			timeHourUnit,
		).
		Joins(
			`LEFT JOIN "RoutineTable" routine
				ON routine.id = routine_task.routine_id
				AND routine.deleted_at IS NULL`,
		).
		Joins(
			`LEFT JOIN "UsersToStationsTable" uts
				ON uts.station_id = routine.station_id
				AND uts.user_id = ?
				AND uts.permission = ?`,
			userId,
			permission,
		).
		Group("buckets.bucket_start").
		Order("buckets.bucket_start ASC").
		Scan(&buckets)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(err)
	}

	data := make([]capi.RoutineTaskCountDatum, len(buckets))
	for index, bucket := range buckets {
		bucketEnd := bucket.BucketStart.Add(time.Duration(timeHourUnit) * time.Hour)
		x := bucket.BucketStart.Format(time.DateOnly)
		if timeHourUnit < 24 {
			x = bucket.BucketStart.Format("2006-01-02 15:04")
		}

		metadata := map[string]any{
			"bucketStart":  bucket.BucketStart,
			"bucketEnd":    bucketEnd,
			"timeHourUnit": timeHourUnit,
			"field":        fieldName,
		}
		meta, err := json.Marshal(metadata)
		if err != nil {
			return nil, apiexceptions.NewRoutineException().FailedToMarshalData(metadata).WithOrigin(err)
		}

		data[index] = capi.RoutineTaskCountDatum{
			Id:    bucket.BucketStart.Format(time.RFC3339),
			X:     x,
			Value: bucket.RoutineTaskCount,
			Meta:  meta,
		}
	}

	return data, nil
}

/* ============================== Service Methods for RoutineTask ============================== */

/* ============================== Main Methods ============================== */

func (s *RoutineTaskService) GetMyRoutineTaskById(
	ctx context.Context, reqDto *capi.GetMyRoutineTaskByIdRequestDto,
) (*capi.GetMyRoutineTaskByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}
	if reqDto.Param.IsDeleted != nil && *reqDto.Param.IsDeleted {
		return nil, apiexceptions.NewRoutineTaskException().NotFound()
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	routineTask, exception := s.routineTaskRepository.GetOneById(
		reqDto.Param.RoutineTaskId,
		actorUserId,
		nil,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.GetMyRoutineTaskByIdResponseDto{
		Id:              routineTask.Id,
		RoutineId:       routineTask.RoutineId,
		Title:           routineTask.Title,
		Purpose:         routineTask.Purpose,
		Payload:         routineTask.Payload,
		CostUnit:        routineTask.CostUnit,
		Priority:        routineTask.Priority,
		Status:          routineTask.Status,
		Attempts:        routineTask.Attempts,
		MaxAttempts:     routineTask.MaxAttempts,
		Period:          routineTask.Period,
		NextScheduledAt: routineTask.NextScheduledAt,
		ScheduledAt:     routineTask.ScheduledAt,
		ActualStartedAt: routineTask.ActualStartedAt,
		ActualEndedAt:   routineTask.ActualEndedAt,
		UpdatedAt:       routineTask.UpdatedAt,
		CreatedAt:       routineTask.CreatedAt,
	}, nil
}

func (s *RoutineTaskService) GetAllMyRoutineTasksByRoutineIds(
	ctx context.Context, reqDto *capi.GetAllMyRoutineTasksByRoutineIdsRequestDto,
) (*capi.GetAllMyRoutineTasksByRoutineIdsResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}
	if reqDto.Param.AreDeleted != nil && *reqDto.Param.AreDeleted {
		resDto := capi.GetAllMyRoutineTasksByRoutineIdsResponseDto{}
		return &resDto, nil
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	routineTasks, exception := s.routineTaskRepository.GetAllByRoutineIds(
		reqDto.Param.RoutineIds,
		actorUserId,
		nil,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	resDto := make(capi.GetAllMyRoutineTasksByRoutineIdsResponseDto, len(routineTasks))
	for index, routineTask := range routineTasks {
		resDto[index] = capi.RoutineTaskResponseDto{
			Id:              routineTask.Id,
			RoutineId:       routineTask.RoutineId,
			Title:           routineTask.Title,
			Purpose:         routineTask.Purpose,
			CostUnit:        routineTask.CostUnit,
			Priority:        routineTask.Priority,
			Status:          routineTask.Status,
			Attempts:        routineTask.Attempts,
			MaxAttempts:     routineTask.MaxAttempts,
			Period:          routineTask.Period,
			NextScheduledAt: routineTask.NextScheduledAt,
			ScheduledAt:     routineTask.ScheduledAt,
			ActualStartedAt: routineTask.ActualStartedAt,
			ActualEndedAt:   routineTask.ActualEndedAt,
			UpdatedAt:       routineTask.UpdatedAt,
			CreatedAt:       routineTask.CreatedAt,
		}
	}

	return &resDto, nil
}

func (s *RoutineTaskService) GetAllMyRoutineTasks(
	ctx context.Context, reqDto *capi.GetAllMyRoutineTasksRequestDto,
) (*capi.GetAllMyRoutineTasksResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}
	if reqDto.Param.AreDeleted != nil && *reqDto.Param.AreDeleted {
		resDto := capi.GetAllMyRoutineTasksResponseDto{}
		return &resDto, nil
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	routineTasks, exception := s.routineTaskRepository.GetAllByUserId(
		actorUserId,
		nil,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	resDto := make(capi.GetAllMyRoutineTasksResponseDto, len(routineTasks))
	for index, routineTask := range routineTasks {
		resDto[index] = capi.GetMyRoutineTaskByIdResponseDto{
			Id:              routineTask.Id,
			RoutineId:       routineTask.RoutineId,
			Title:           routineTask.Title,
			Purpose:         routineTask.Purpose,
			Payload:         routineTask.Payload,
			CostUnit:        routineTask.CostUnit,
			Priority:        routineTask.Priority,
			Status:          routineTask.Status,
			Attempts:        routineTask.Attempts,
			MaxAttempts:     routineTask.MaxAttempts,
			Period:          routineTask.Period,
			NextScheduledAt: routineTask.NextScheduledAt,
			ScheduledAt:     routineTask.ScheduledAt,
			ActualStartedAt: routineTask.ActualStartedAt,
			ActualEndedAt:   routineTask.ActualEndedAt,
			UpdatedAt:       routineTask.UpdatedAt,
			CreatedAt:       routineTask.CreatedAt,
		}
	}

	return &resDto, nil
}

func (s *RoutineTaskService) CreateRoutineTaskByRoutineId(
	ctx context.Context, reqDto *capi.CreateRoutineTaskByRoutineIdRequestDto,
) (*capi.CreateRoutineTaskByRoutineIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}
	if exception := s.routineTaskExecutionService.ValidateRoutineTaskPayload(
		reqDto.Body.Purpose,
		reqDto.Body.Payload,
	); exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	newRoutineTaskId, exception := s.routineTaskRepository.CreateOneByRoutineId(
		reqDto.Body.RoutineId,
		actorUserId,
		sinputs.CreateRoutineTaskInput{
			ActorUserId:     actorUserId,
			Title:           reqDto.Body.Title,
			Purpose:         reqDto.Body.Purpose,
			Payload:         reqDto.Body.Payload,
			Priority:        reqDto.Body.Priority,
			MaxAttempts:     reqDto.Body.MaxAttempts,
			Period:          reqDto.Body.Period,
			NextScheduledAt: reqDto.Body.NextScheduledAt,
		},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.CreateRoutineTaskByRoutineIdResponseDto{
		Id:        *newRoutineTaskId,
		CreatedAt: time.Now(),
	}, nil
}

func (s *RoutineTaskService) UpdateMyRoutineTaskById(
	ctx context.Context, reqDto *capi.UpdateMyRoutineTaskByIdRequestDto,
) (*capi.UpdateMyRoutineTaskByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	if reqDto.Body.Values.Purpose != nil || reqDto.Body.Values.Payload != nil {
		var finalPurpose cenums.RoutineTaskPurpose
		finalPayload := reqDto.Body.Values.Payload
		if reqDto.Body.Values.Purpose == nil || finalPayload == nil {
			existingRoutineTask, exception := s.routineTaskRepository.GetOneById(
				reqDto.Body.RoutineTaskId,
				actorUserId,
				nil,
				srepositories.WithDB(db),
				srepositories.WithAllowedPermissions(allowedPermissions),
			)
			if exception != nil {
				return nil, exception
			}
			if reqDto.Body.Values.Purpose == nil {
				finalPurpose = existingRoutineTask.Purpose
			} else {
				finalPurpose = cenums.RoutineTaskPurpose(*reqDto.Body.Values.Purpose)
			}
			if finalPayload == nil {
				finalPayload = &existingRoutineTask.Payload
			}
		} else {
			finalPurpose = *reqDto.Body.Values.Purpose
		}
		if exception := s.routineTaskExecutionService.ValidateRoutineTaskPayload(finalPurpose, *finalPayload); exception != nil {
			return nil, exception
		}
	}

	updatedRoutineTask, exception := s.routineTaskRepository.UpdateOneById(
		reqDto.Body.RoutineTaskId,
		actorUserId,
		sinputs.PartialUpdateRoutineTaskInput{
			Values: sinputs.UpdateRoutineTaskInput{
				RoutineId:       reqDto.Body.Values.RoutineId,
				Title:           reqDto.Body.Values.Title,
				Purpose:         reqDto.Body.Values.Purpose,
				Payload:         reqDto.Body.Values.Payload,
				Priority:        reqDto.Body.Values.Priority,
				MaxAttempts:     reqDto.Body.Values.MaxAttempts,
				Period:          reqDto.Body.Values.Period,
				NextScheduledAt: reqDto.Body.Values.NextScheduledAt,
			},
			SetNull: reqDto.Body.SetNull,
		},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.UpdateMyRoutineTaskByIdResponseDto{
		UpdatedAt: updatedRoutineTask.UpdatedAt,
	}, nil
}

func (s *RoutineTaskService) PauseMyRoutineTaskById(
	ctx context.Context, reqDto *capi.PauseMyRoutineTaskByIdRequestDto,
) (*capi.PauseMyRoutineTaskByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	routineTask, exception := s.routineTaskRepository.CheckPermissionAndGetOneById(
		reqDto.Body.RoutineTaskId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if routineTask.Status != cenums.RoutineTaskStatus_Idle {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().InvalidInput("only idle routine tasks can be paused")
	}

	now := time.Now()
	result := tx.Model(&sschemas.RoutineTask{}).
		Where("id = ? AND status = ?", reqDto.Body.RoutineTaskId, cenums.RoutineTaskStatus_Idle).
		Updates(map[string]any{
			"status":     cenums.RoutineTaskStatus_Pause,
			"updated_at": now,
		})
	if result.Error != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().FailedToUpdate().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().NoChanges()
	}

	if err := tx.Commit().Error; err != nil {
		return nil, apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.PauseMyRoutineTaskByIdResponseDto{UpdatedAt: now}, nil
}

func (s *RoutineTaskService) ResumeMyRoutineTaskById(
	ctx context.Context, reqDto *capi.ResumeMyRoutineTaskByIdRequestDto,
) (*capi.ResumeMyRoutineTaskByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	routineTask, exception := s.routineTaskRepository.CheckPermissionAndGetOneById(
		reqDto.Body.RoutineTaskId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if routineTask.Status != cenums.RoutineTaskStatus_Pause {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().InvalidInput("only paused routine tasks can be resumed")
	}

	now := time.Now()
	result := tx.Model(&sschemas.RoutineTask{}).
		Where("id = ? AND status = ?", reqDto.Body.RoutineTaskId, cenums.RoutineTaskStatus_Pause).
		Updates(map[string]any{
			"status":     cenums.RoutineTaskStatus_Idle,
			"updated_at": now,
		})
	if result.Error != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().FailedToUpdate().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().NoChanges()
	}

	if err := tx.Commit().Error; err != nil {
		return nil, apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.ResumeMyRoutineTaskByIdResponseDto{UpdatedAt: now}, nil
}

func (s *RoutineTaskService) HardDeleteMyRoutineTaskById(
	ctx context.Context, reqDto *capi.HardDeleteMyRoutineTaskByIdRequestDto,
) (*capi.HardDeleteMyRoutineTaskByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.routineTaskRepository.HardDeleteOneById(
		reqDto.Body.RoutineTaskId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.HardDeleteMyRoutineTaskByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *RoutineTaskService) HardDeleteMyRoutineTasksByIds(
	ctx context.Context, reqDto *capi.HardDeleteMyRoutineTasksByIdsRequestDto,
) (*capi.HardDeleteMyRoutineTasksByIdsResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.routineTaskRepository.HardDeleteManyByIds(
		reqDto.Body.RoutineTaskIds,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.HardDeleteMyRoutineTasksByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

/* ============================== Service Methods for Charts ============================== */

func (s *RoutineTaskService) VisualizeMyRoutineTaskStatusCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskStatusCountRequestDto,
) (*capi.VisualizeMyRoutineTaskStatusCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	var rows []struct {
		Status           cenums.RoutineTaskStatus `gorm:"column:status;"`
		RoutineTaskCount int64                    `gorm:"column:routine_task_count;"`
	}
	result := db.Model(&sschemas.RoutineTask{}).
		Select(`"RoutineTaskTable".status AS status, COUNT(*) AS routine_task_count`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = "RoutineTaskTable".routine_id AND routine.deleted_at IS NULL`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Where("uts.user_id = ? AND uts.permission = ?", actorUserId, reqDto.Param.Permission).
		Group(`"RoutineTaskTable".status`).
		Scan(&rows)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(err)
	}

	counts := make(map[cenums.RoutineTaskStatus]int64, len(rows))
	for _, row := range rows {
		counts[row.Status] = row.RoutineTaskCount
	}

	data := make([]capi.RoutineTaskCountDatum, len(cenums.AllRoutineTaskStatuses))
	for index, status := range cenums.AllRoutineTaskStatuses {
		metadata := map[string]string{"status": status.String()}
		meta, err := json.Marshal(metadata)
		if err != nil {
			return nil, apiexceptions.NewRoutineException().FailedToMarshalData(metadata)
		}

		data[index] = capi.RoutineTaskCountDatum{
			Id:    status.String() + "-routine-task-count",
			X:     status.String() + " Routine Task Count",
			Value: counts[status],
			Meta:  meta,
		}
	}

	return &capi.VisualizeMyRoutineTaskStatusCountResponseDto{
		Data: data,
	}, nil
}

func (s *RoutineTaskService) VisualizeMyRoutineTaskPurposeCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskPurposeCountRequestDto,
) (*capi.VisualizeMyRoutineTaskPurposeCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	var rows []struct {
		Purpose          cenums.RoutineTaskPurpose `gorm:"column:purpose;"`
		RoutineTaskCount int64                     `gorm:"column:routine_task_count;"`
	}
	result := db.Model(&sschemas.RoutineTask{}).
		Select(`"RoutineTaskTable".purpose AS purpose, COUNT(*) AS routine_task_count`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = "RoutineTaskTable".routine_id AND routine.deleted_at IS NULL`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Where("uts.user_id = ? AND uts.permission = ?", actorUserId, reqDto.Param.Permission).
		Group(`"RoutineTaskTable".purpose`).
		Scan(&rows)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(err)
	}

	counts := make(map[cenums.RoutineTaskPurpose]int64, len(rows))
	for _, row := range rows {
		counts[row.Purpose] = row.RoutineTaskCount
	}

	data := make([]capi.RoutineTaskCountDatum, len(cenums.AllRoutineTaskPurposes))
	for index, purpose := range cenums.AllRoutineTaskPurposes {
		metadata := map[string]string{"purpose": purpose.String()}
		meta, err := json.Marshal(metadata)
		if err != nil {
			return nil, apiexceptions.NewRoutineException().FailedToMarshalData(metadata)
		}

		data[index] = capi.RoutineTaskCountDatum{
			Id:    purpose.String() + "-routine-task-count",
			X:     purpose.String() + " Routine Task Count",
			Value: counts[purpose],
			Meta:  meta,
		}
	}

	return &capi.VisualizeMyRoutineTaskPurposeCountResponseDto{
		Data: data,
	}, nil
}

func (s *RoutineTaskService) VisualizeMyRoutineTaskScheduledAtCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskScheduledAtCountRequestDto,
) (*capi.VisualizeMyRoutineTaskScheduledAtCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}
	if !reqDto.Param.QueryRangeStartedAt.Before(reqDto.Param.QueryRangeEndedAt) {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto("queryRangeStartedAt should be earlier then queryRangeEndedAt")
	}
	if !stimes.IsTimeWithin(reqDto.Param.QueryRangeStartedAt, reqDto.Param.QueryRangeEndedAt, 360*24*time.Hour) {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto("queryRangeStartedAt and queryRangeEndedAt should be within 360 days")
	}

	data, exception := s.visualizeMyRoutineTaskTimeCount(
		ctx,
		actorUserId,
		cenums.AccessControlPermission(reqDto.Param.Permission),
		reqDto.Param.TimeHourUnit,
		reqDto.Param.QueryRangeStartedAt,
		reqDto.Param.QueryRangeEndedAt,
		"scheduled_at",
		"scheduledAt",
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.VisualizeMyRoutineTaskScheduledAtCountResponseDto{
		Data: data,
	}, nil
}

func (s *RoutineTaskService) VisualizeMyRoutineTaskActualStartedAtCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskActualStartedAtCountRequestDto,
) (*capi.VisualizeMyRoutineTaskActualStartedAtCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}
	if !reqDto.Param.QueryRangeStartedAt.Before(reqDto.Param.QueryRangeEndedAt) {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto("queryRangeStartedAt should be earlier then queryRangeEndedAt")
	}
	if !stimes.IsTimeWithin(reqDto.Param.QueryRangeStartedAt, reqDto.Param.QueryRangeEndedAt, 360*24*time.Hour) {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto("queryRangeStartedAt and queryRangeEndedAt should be within 360 days")
	}

	data, exception := s.visualizeMyRoutineTaskTimeCount(
		ctx,
		actorUserId,
		cenums.AccessControlPermission(reqDto.Param.Permission),
		reqDto.Param.TimeHourUnit,
		reqDto.Param.QueryRangeStartedAt,
		reqDto.Param.QueryRangeEndedAt,
		"actual_started_at",
		"actualStartedAt",
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.VisualizeMyRoutineTaskActualStartedAtCountResponseDto{
		Data: data,
	}, nil
}

func (s *RoutineTaskService) VisualizeMyRoutineTaskActualEndedAtCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskActualEndedAtCountRequestDto,
) (*capi.VisualizeMyRoutineTaskActualEndedAtCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto().WithOrigin(err)
	}
	if !reqDto.Param.QueryRangeStartedAt.Before(reqDto.Param.QueryRangeEndedAt) {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto("queryRangeStartedAt should be earlier then queryRangeEndedAt")
	}
	if !stimes.IsTimeWithin(reqDto.Param.QueryRangeStartedAt, reqDto.Param.QueryRangeEndedAt, 360*24*time.Hour) {
		return nil, apiexceptions.NewRoutineTaskException().InvalidDto("queryRangeStartedAt and queryRangeEndedAt should be within 360 days")
	}

	data, exception := s.visualizeMyRoutineTaskTimeCount(
		ctx,
		actorUserId,
		cenums.AccessControlPermission(reqDto.Param.Permission),
		reqDto.Param.TimeHourUnit,
		reqDto.Param.QueryRangeStartedAt,
		reqDto.Param.QueryRangeEndedAt,
		"actual_ended_at",
		"actualEndedAt",
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.VisualizeMyRoutineTaskActualEndedAtCountResponseDto{
		Data: data,
	}, nil
}

/* ============================== Service Methods for GraphQL RoutineTask ============================== */

func (s *RoutineTaskService) SearchPrivateRoutineTasks(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineTaskInput,
) (*cgqlmodels.SearchRoutineTaskConnection, *cexceptions.Exception) {
	type PrivateRoutineTask struct {
		sschemas.RoutineTask
		Permission cenums.AccessControlPermission `gorm:"column:permission"`
	}

	startTime := time.Now()
	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	query := db.Model(&sschemas.RoutineTask{}).
		Select(`"RoutineTaskTable".*, uts.permission AS permission`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = "RoutineTaskTable".routine_id AND routine.deleted_at IS NULL`).
		Joins(`LEFT JOIN "UsersToStationsTable" uts ON routine.station_id = uts.station_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions)

	if len(gqlInput.RoutineIds) > 0 {
		query = query.Where(
			`"RoutineTaskTable".routine_id IN ?`,
			gqlInput.RoutineIds,
		)
	}

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"title ILIKE ? OR purpose::text ILIKE ? OR payload::text ILIKE ?",
			"%"+gqlInput.Query+"%",
			"%"+gqlInput.Query+"%",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchRoutineTaskCursorFields](*gqlInput.After)
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToDecode().WithOrigin(err)
		}

		query = query.Where("id > ?", searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		var cending string = cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchRoutineTaskSortByTitle:
			query = query.Order("title " + cending).
				Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByPurpose:
			query = query.Order("purpose " + cending).
				Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByPriority:
			query = query.Order("priority " + cending).
				Order("scheduled_at " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByStatus:
			query = query.Order("status " + cending).
				Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByAttempts:
			query = query.Order("attempts " + cending).
				Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByMaxAttempts:
			query = query.Order("max_attempts " + cending).
				Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByScheduledAt:
			query = query.Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByActualStartedAt:
			query = query.Order("actual_started_at " + cending).
				Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByActualEndedAt:
			query = query.Order("actual_ended_at " + cending).
				Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("scheduled_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		}
	}

	limit := sconstants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, sconstants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var routineTasks []PrivateRoutineTask
	if err := query.Scopes(s.routineTaskScope.IncludePreloads(
		nil,
	)).Find(&routineTasks).Error; err != nil {
		return nil, apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(err)
	}

	hasNextPage := len(routineTasks) > limit
	searchEdges := make([]*cgqlmodels.SearchRoutineTaskEdge, len(routineTasks))

	for index, routineTask := range routineTasks {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchRoutineTaskCursorFields]{
			Fields: cgqlmodels.SearchRoutineTaskCursorFields{
				ID: routineTask.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, apiexceptions.NewSearchException().FailedToUnmarshalSearchCursor()
		}

		searchEdges[index] = &cgqlmodels.SearchRoutineTaskEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                routineTask.RoutineTask.ToPrivateRoutineTask(),
		}
	}

	searchPageInfo := &cgqlmodels.SearchPageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0,
	}

	if len(searchEdges) > 0 {
		searchPageInfo.StartEncodedSearchCursor = &searchEdges[0].EncodedSearchCursor
		searchPageInfo.EndEncodedSearchCursor = &searchEdges[len(searchEdges)-1].EncodedSearchCursor
	}

	searchTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	if hasNextPage {
		searchEdges = searchEdges[:limit]
	}

	return &cgqlmodels.SearchRoutineTaskConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}

/* ============================== System Methods for DurableJob RoutineTask ============================== */

func (s *RoutineTaskService) MarkCompletedRoutineTasks(
	ctx context.Context,
	eventId uuid.UUID,
	request *cdurablejob.MarkCompletedRoutineTasksRequestDto,
) *cexceptions.Exception {
	if eventId == uuid.Nil || request == nil || request.WorkerId == uuid.Nil || len(request.Tasks) == 0 {
		return cexceptions.New(
			"InvalidDto",
			"RoutineTask",
			"MarkCompletedRoutineTasks",
			"The routine task completion response is invalid",
			http.StatusBadRequest,
		)
	}
	if err := s.validator.Struct(request); err != nil {
		return cexceptions.New(
			"InvalidDto",
			"RoutineTask",
			"MarkCompletedRoutineTasks",
			"The routine task completion response is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
	}
	isNewInboxEvent, inboxException := srepositories.NewInboxEventRepository().CreateOne(
		sinputs.CreateInboxEventInput{EventId: eventId},
		srepositories.RepositoryOptionFields{DB: tx, IsTransactionStarted: true},
	)
	if inboxException != nil {
		tx.Rollback()
		return cexceptions.New("FailedToRecordInboxEvent", "RoutineTask", "MarkCompletedRoutineTasks", "Failed to record the Kafka result event", http.StatusInternalServerError, true).WithOrigin(inboxException)
	}
	if !isNewInboxEvent {
		if err := tx.Commit().Error; err != nil {
			return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
		}
		return nil
	}

	now := time.Now().UTC()
	taskIds := make([]uuid.UUID, len(request.Tasks))
	recordIds := make([]uuid.UUID, len(request.Tasks))
	for index, task := range request.Tasks {
		taskIds[index] = task.RoutineTaskId
		recordIds[index] = task.RoutineTaskRecordId
	}
	result := tx.Model(&sschemas.RoutineTask{}).
		Where("id IN ? AND status = ?", taskIds, cenums.RoutineTaskStatus_Running).
		Updates(map[string]any{
			"status":          cenums.RoutineTaskStatus_Idle,
			"attempts":        0,
			"actual_ended_at": now,
			"updated_at":      now,
		})
	if result.Error != nil || result.RowsAffected != int64(len(taskIds)) {
		var finalizedRecordCount int64
		finalizedResult := tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Success).
			Count(&finalizedRecordCount)
		if result.Error == nil && finalizedResult.Error == nil && finalizedRecordCount == int64(len(recordIds)) {
			if err := tx.Commit().Error; err != nil {
				return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
			}
			return nil
		}
		tx.Rollback()
		if result.Error != nil {
			return apiexceptions.NewRoutineTaskException().FailedToUpdate().WithOrigin(result.Error)
		}
		return cexceptions.New("ResultStateMismatch", "RoutineTask", "MarkCompletedRoutineTasks", "Routine task completion count does not match the claimed batch", http.StatusConflict, true)
	}
	result = tx.Model(&sschemas.RoutineTaskRecord{}).
		Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Running).
		Updates(map[string]any{
			"status":          cenums.RoutineTaskRecordStatus_Success,
			"actual_ended_at": now,
			"error_code":      nil,
			"error_reason":    nil,
			"updated_at":      now,
		})
	if result.Error != nil || result.RowsAffected != int64(len(recordIds)) {
		var finalizedRecordCount int64
		finalizedResult := tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Success).
			Count(&finalizedRecordCount)
		if result.Error == nil && finalizedResult.Error == nil && finalizedRecordCount == int64(len(recordIds)) {
			if err := tx.Commit().Error; err != nil {
				return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
			}
			return nil
		}
		tx.Rollback()
		if result.Error != nil {
			return apiexceptions.NewRoutineTaskException().FailedToUpdate().WithOrigin(result.Error)
		}
		return cexceptions.New("ResultStateMismatch", "RoutineTaskRecord", "MarkCompletedRoutineTasks", "Routine task record completion count does not match the claimed batch", http.StatusConflict, true)
	}
	if err := tx.Commit().Error; err != nil {
		return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
	}
	return nil
}

func (s *RoutineTaskService) MarkFailedRoutineTasks(
	ctx context.Context,
	eventId uuid.UUID,
	request *cdurablejob.MarkFailedRoutineTasksRequestDto,
) *cexceptions.Exception {
	if eventId == uuid.Nil || request == nil || request.WorkerId == uuid.Nil || len(request.Tasks) == 0 {
		return cexceptions.New("InvalidDto", "RoutineTask", "MarkFailedRoutineTasks", "The routine task failure response is invalid", http.StatusBadRequest)
	}
	if err := s.validator.Struct(request); err != nil {
		return cexceptions.New("InvalidDto", "RoutineTask", "MarkFailedRoutineTasks", "The routine task failure response is invalid", http.StatusBadRequest).WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
	}
	isNewInboxEvent, inboxException := srepositories.NewInboxEventRepository().CreateOne(
		sinputs.CreateInboxEventInput{EventId: eventId},
		srepositories.RepositoryOptionFields{DB: tx, IsTransactionStarted: true},
	)
	if inboxException != nil {
		tx.Rollback()
		return cexceptions.New("FailedToRecordInboxEvent", "RoutineTask", "MarkFailedRoutineTasks", "Failed to record the Kafka result event", http.StatusInternalServerError, true).WithOrigin(inboxException)
	}
	if !isNewInboxEvent {
		if err := tx.Commit().Error; err != nil {
			return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
		}
		return nil
	}

	now := time.Now().UTC()
	taskIds := make([]uuid.UUID, 0, len(request.Tasks))
	recordIds := make([]uuid.UUID, 0, len(request.Tasks))
	failureInputs := make([]sinputs.UpdateRoutineTaskRecordFailureInput, 0, len(request.Tasks))
	for _, task := range request.Tasks {
		taskIds = append(taskIds, task.RoutineTaskId)
		recordIds = append(recordIds, task.RoutineTaskRecordId)
		failureInputs = append(failureInputs, sinputs.UpdateRoutineTaskRecordFailureInput{
			Id:          task.RoutineTaskRecordId,
			ErrorCode:   cenums.RoutineTaskRecordErrorCode(task.ErrorCode),
			ErrorReason: task.ErrorReason,
		})
	}
	result := tx.Model(&sschemas.RoutineTask{}).
		Where("id IN ? AND status = ?", taskIds, cenums.RoutineTaskStatus_Running).
		Updates(map[string]any{
			"status":          cenums.RoutineTaskStatus_Idle,
			"actual_ended_at": now,
			"updated_at":      now,
		})
	if result.Error != nil || result.RowsAffected != int64(len(taskIds)) {
		var finalizedRecordCount int64
		finalizedResult := tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Failed).
			Count(&finalizedRecordCount)
		if result.Error == nil && finalizedResult.Error == nil && finalizedRecordCount == int64(len(request.Tasks)) {
			if err := tx.Commit().Error; err != nil {
				return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
			}
			return nil
		}
		tx.Rollback()
		if result.Error != nil {
			return apiexceptions.NewRoutineTaskException().FailedToUpdate().WithOrigin(result.Error)
		}
		return cexceptions.New("ResultStateMismatch", "RoutineTask", "MarkFailedRoutineTasks", "Routine task failure count does not match the claimed batch", http.StatusConflict, true)
	}
	updatedRecordCount, exception := s.routineTaskRecordRepository.UpdateManyAsFailed(
		failureInputs,
		now,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	if updatedRecordCount != int64(len(request.Tasks)) {
		var finalizedRecordCount int64
		finalizedResult := tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Failed).
			Count(&finalizedRecordCount)
		if finalizedResult.Error == nil && finalizedRecordCount == int64(len(request.Tasks)) {
			if err := tx.Commit().Error; err != nil {
				return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
			}
			return nil
		}
		tx.Rollback()
		return cexceptions.New("ResultStateMismatch", "RoutineTaskRecord", "MarkFailedRoutineTasks", "Routine task record failure count does not match the claimed batch", http.StatusConflict, true)
	}
	if err := tx.Commit().Error; err != nil {
		return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
	}
	return nil
}
