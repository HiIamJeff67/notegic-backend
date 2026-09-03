package routines

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-tasks"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/core/exceptions"
	parsers "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/routines/parsers"
)

type RoutineTaskServiceInterface interface {
	GetMyRoutineTaskById(ctx context.Context, reqDto *capi.GetMyRoutineTaskByIdRequestDto) (*capi.GetMyRoutineTaskByIdResponseDto, *cexceptions.Exception)
	GetAllMyRoutineTasksByRoutineIds(ctx context.Context, reqDto *capi.GetAllMyRoutineTasksByRoutineIdsRequestDto) (*capi.GetAllMyRoutineTasksByRoutineIdsResponseDto, *cexceptions.Exception)
	GetAllMyRoutineTasks(ctx context.Context, reqDto *capi.GetAllMyRoutineTasksRequestDto) (*capi.GetAllMyRoutineTasksResponseDto, *cexceptions.Exception)
	CreateRoutineTaskByRoutineId(ctx context.Context, reqDto *capi.CreateRoutineTaskByRoutineIdRequestDto) (*capi.CreateRoutineTaskByRoutineIdResponseDto, *cexceptions.Exception)
	UpdateMyRoutineTaskById(ctx context.Context, reqDto *capi.UpdateMyRoutineTaskByIdRequestDto) (*capi.UpdateMyRoutineTaskByIdResponseDto, *cexceptions.Exception)
	HardDeleteMyRoutineTaskById(ctx context.Context, reqDto *capi.HardDeleteMyRoutineTaskByIdRequestDto) (*capi.HardDeleteMyRoutineTaskByIdResponseDto, *cexceptions.Exception)
	HardDeleteMyRoutineTasksByIds(ctx context.Context, reqDto *capi.HardDeleteMyRoutineTasksByIdsRequestDto) (*capi.HardDeleteMyRoutineTasksByIdsResponseDto, *cexceptions.Exception)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyRoutineTaskPurposeCount(ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskPurposeCountRequestDto) (*capi.VisualizeMyRoutineTaskPurposeCountResponseDto, *cexceptions.Exception)

	/* ============================== GraphQL Methods ============================== */
	SearchPrivateRoutineTasks(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineTaskInput) (*cgqlmodels.SearchRoutineTaskConnection, *cexceptions.Exception)
}

type RoutineTaskService struct {
	validator             *validator.Validate
	db                    *gorm.DB
	routineTaskScope      sscopes.RoutineTaskScopeInterface
	routineTaskRepository srepositories.RoutineTaskRepositoryInterface
	payloadParser         parsers.RoutineTaskPayloadParserInterface
	routineTaskException  apiexceptions.RoutineTaskException
	routineException      apiexceptions.RoutineException
	searchException       apiexceptions.SearchException
}

func NewRoutineTaskService(
	validator *validator.Validate,
	db *gorm.DB,
	routineTaskScope sscopes.RoutineTaskScopeInterface,
	routineTaskRepository srepositories.RoutineTaskRepositoryInterface,
	payloadParser parsers.RoutineTaskPayloadParserInterface,
	routineTaskException apiexceptions.RoutineTaskException,
	routineException apiexceptions.RoutineException,
	searchException apiexceptions.SearchException,
) RoutineTaskServiceInterface {
	return &RoutineTaskService{
		validator:             validator,
		db:                    db,
		routineTaskScope:      routineTaskScope,
		routineTaskRepository: routineTaskRepository,
		payloadParser:         payloadParser,
		routineTaskException:  routineTaskException,
		routineException:      routineException,
		searchException:       searchException,
	}
}

func (s *RoutineTaskService) GetMyRoutineTaskById(
	ctx context.Context, reqDto *capi.GetMyRoutineTaskByIdRequestDto,
) (*capi.GetMyRoutineTaskByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	if reqDto.Param.IsDeleted != nil && *reqDto.Param.IsDeleted {
		return nil, s.routineTaskException.NotFound()
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	routineTask, exception := s.routineTaskRepository.GetOneById(
		reqDto.Param.RoutineTaskId,
		actorUserId,
		[]sschemas.RoutineTaskRelation{sschemas.RoutineTaskRelation_PreviousTasks},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.GetMyRoutineTaskByIdResponseDto{
		Id:                     routineTask.Id,
		RoutineId:              routineTask.RoutineId,
		Title:                  routineTask.Title,
		Purpose:                routineTask.Purpose,
		Payload:                routineTask.Payload,
		CostUnit:               routineTask.CostUnit,
		Priority:               routineTask.Priority,
		MaxAttempts:            routineTask.MaxAttempts,
		PreviousRoutineTaskIds: routineTask.ToPrivateRoutineTask().PreviousRoutineTaskIds,
		UpdatedAt:              routineTask.UpdatedAt,
		CreatedAt:              routineTask.CreatedAt,
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
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
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
		[]sschemas.RoutineTaskRelation{sschemas.RoutineTaskRelation_PreviousTasks},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	resDto := make(capi.GetAllMyRoutineTasksByRoutineIdsResponseDto, len(routineTasks))
	for index, routineTask := range routineTasks {
		resDto[index] = capi.RoutineTaskResponseDto{
			Id:                     routineTask.Id,
			RoutineId:              routineTask.RoutineId,
			Title:                  routineTask.Title,
			Purpose:                routineTask.Purpose,
			CostUnit:               routineTask.CostUnit,
			Priority:               routineTask.Priority,
			MaxAttempts:            routineTask.MaxAttempts,
			PreviousRoutineTaskIds: routineTask.ToPrivateRoutineTask().PreviousRoutineTaskIds,
			UpdatedAt:              routineTask.UpdatedAt,
			CreatedAt:              routineTask.CreatedAt,
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
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
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
		[]sschemas.RoutineTaskRelation{sschemas.RoutineTaskRelation_PreviousTasks},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	resDto := make(capi.GetAllMyRoutineTasksResponseDto, len(routineTasks))
	for index, routineTask := range routineTasks {
		resDto[index] = capi.GetMyRoutineTaskByIdResponseDto{
			Id:                     routineTask.Id,
			RoutineId:              routineTask.RoutineId,
			Title:                  routineTask.Title,
			Purpose:                routineTask.Purpose,
			Payload:                routineTask.Payload,
			CostUnit:               routineTask.CostUnit,
			Priority:               routineTask.Priority,
			MaxAttempts:            routineTask.MaxAttempts,
			PreviousRoutineTaskIds: routineTask.ToPrivateRoutineTask().PreviousRoutineTaskIds,
			UpdatedAt:              routineTask.UpdatedAt,
			CreatedAt:              routineTask.CreatedAt,
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
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	if exception := s.payloadParser.ValidateRoutineTaskPayload(
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
			ActorUserId: actorUserId,
			Title:       reqDto.Body.Title,
			Purpose:     reqDto.Body.Purpose,
			Payload:     reqDto.Body.Payload,
			Priority:    reqDto.Body.Priority,
			MaxAttempts: reqDto.Body.MaxAttempts,
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
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
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
		if exception := s.payloadParser.ValidateRoutineTaskPayload(finalPurpose, *finalPayload); exception != nil {
			return nil, exception
		}
	}

	updatedRoutineTask, exception := s.routineTaskRepository.UpdateOneById(
		reqDto.Body.RoutineTaskId,
		actorUserId,
		sinputs.PartialUpdateRoutineTaskInput{
			Values: sinputs.UpdateRoutineTaskInput{
				RoutineId:   reqDto.Body.Values.RoutineId,
				Title:       reqDto.Body.Values.Title,
				Purpose:     reqDto.Body.Values.Purpose,
				Payload:     reqDto.Body.Values.Payload,
				Priority:    reqDto.Body.Values.Priority,
				MaxAttempts: reqDto.Body.Values.MaxAttempts,
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

func (s *RoutineTaskService) HardDeleteMyRoutineTaskById(
	ctx context.Context, reqDto *capi.HardDeleteMyRoutineTaskByIdRequestDto,
) (*capi.HardDeleteMyRoutineTaskByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
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
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
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

/* ============================== Visualization Methods ============================== */

func (s *RoutineTaskService) VisualizeMyRoutineTaskPurposeCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutineTaskPurposeCountRequestDto,
) (*capi.VisualizeMyRoutineTaskPurposeCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
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
		return nil, s.routineTaskException.NotFound().WithOrigin(err)
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
			return nil, s.routineException.FailedToMarshalData(metadata)
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

/* ============================== GraphQL Methods ============================== */

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
			return nil, s.searchException.FailedToDecode().WithOrigin(err)
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
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByPurpose:
			query = query.Order("purpose " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByPriority:
			query = query.Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByMaxAttempts:
			query = query.Order("max_attempts " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("priority " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineTaskSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("priority " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("updated_at " + cending).
				Order("priority " + cending).
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
		return nil, s.routineTaskException.NotFound().WithOrigin(err)
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
			return nil, s.searchException.FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, s.searchException.FailedToUnmarshalSearchCursor()
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
