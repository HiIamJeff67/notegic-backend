package routines

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routines"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	stimes "github.com/HiIamJeff67/notegic-backend/shared/lib/times"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type RoutineServiceInterface interface {
	GetMyRoutineById(ctx context.Context, reqDto *capi.GetMyRoutineByIdRequestDto) (*capi.GetMyRoutineByIdResponseDto, *cexceptions.Exception)
	GetMyRoutinesByStationId(ctx context.Context, reqDto *capi.GetMyRoutinesByStationIdRequestDto) (*capi.GetMyRoutinesByStationIdResponseDto, *cexceptions.Exception)
	GetAllMyRoutinesByTimeRange(ctx context.Context, reqDto *capi.GetAllMyRoutinesByTimeRangeRequestDto) (*capi.GetAllMyRoutinesByTimeRangeResponseDto, *cexceptions.Exception)
	CreateRoutineByStationId(ctx context.Context, reqDto *capi.CreateRoutineByStationIdRequestDto) (*capi.CreateRoutineByStationIdResponseDto, *cexceptions.Exception)
	CreateRoutinesByStationIds(ctx context.Context, reqDto *capi.CreateRoutinesByStationIdsRequestDto) (*capi.CreateRoutinesByStationIdsResponseDto, *cexceptions.Exception)
	UpdateMyRoutineById(ctx context.Context, reqDto *capi.UpdateMyRoutineByIdRequestDto) (*capi.UpdateMyRoutineByIdResponseDto, *cexceptions.Exception)
	UpdateMyRoutinesByIds(ctx context.Context, reqDto *capi.UpdateMyRoutinesByIdsRequestDto) (*capi.UpdateMyRoutinesByIdsResponseDto, *cexceptions.Exception)
	LinkRoutineTagById(ctx context.Context, reqDto *capi.LinkRoutineTagByIdRequestDto) (*capi.LinkRoutineTagByIdResponseDto, *cexceptions.Exception)
	LinkRoutineTagsByIds(ctx context.Context, reqDto *capi.LinkRoutineTagsByIdsRequestDto) (*capi.LinkRoutineTagsByIdsResponseDto, *cexceptions.Exception)
	LinkRoutineItemById(ctx context.Context, reqDto *capi.LinkRoutineItemByIdRequestDto) (*capi.LinkRoutineItemByIdResponseDto, *cexceptions.Exception)
	LinkRoutineItemsByIds(ctx context.Context, reqDto *capi.LinkRoutineItemsByIdsRequestDto) (*capi.LinkRoutineItemsByIdsResponseDto, *cexceptions.Exception)
	RestoreMyRoutineById(ctx context.Context, reqDto *capi.RestoreMyRoutineByIdRequestDto) (*capi.RestoreMyRoutineByIdResponseDto, *cexceptions.Exception)
	RestoreMyRoutinesByIds(ctx context.Context, reqDto *capi.RestoreMyRoutinesByIdsRequestDto) (*capi.RestoreMyRoutinesByIdsResponseDto, *cexceptions.Exception)
	DeleteMyRoutineById(ctx context.Context, reqDto *capi.DeleteMyRoutineByIdRequestDto) (*capi.DeleteMyRoutineByIdResponseDto, *cexceptions.Exception)
	DeleteMyRoutinesByIds(ctx context.Context, reqDto *capi.DeleteMyRoutinesByIdsRequestDto) (*capi.DeleteMyRoutinesByIdsResponseDto, *cexceptions.Exception)
	HardDeleteMyRoutineById(ctx context.Context, reqDto *capi.HardDeleteMyRoutineByIdRequestDto) (*capi.HardDeleteMyRoutineByIdResponseDto, *cexceptions.Exception)
	HardDeleteMyRoutinesByIds(ctx context.Context, reqDto *capi.HardDeleteMyRoutinesByIdsRequestDto) (*capi.HardDeleteMyRoutinesByIdsResponseDto, *cexceptions.Exception)

	VisualizeMyRoutineStatusCount(ctx context.Context, reqDto *capi.VisualizeMyRoutineStatusCountRequestDto) (*capi.VisualizeMyRoutineStatusCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutinePeriodCount(ctx context.Context, reqDto *capi.VisualizeMyRoutinePeriodCountRequestDto) (*capi.VisualizeMyRoutinePeriodCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineScheduledStartAtCount(ctx context.Context, reqDto *capi.VisualizeMyRoutineScheduledStartAtCountRequestDto) (*capi.VisualizeMyRoutineScheduledStartAtCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineScheduledEndAtCount(ctx context.Context, reqDto *capi.VisualizeMyRoutineScheduledEndAtCountRequestDto) (*capi.VisualizeMyRoutineScheduledEndAtCountResponseDto, *cexceptions.Exception)

	SearchPrivateRoutines(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineInput) (*cgqlmodels.SearchRoutineConnection, *cexceptions.Exception)
}

type RoutineService struct {
	validator             *validator.Validate
	db                    *gorm.DB
	routineScope          sscopes.RoutineScopeInterface
	stationRepository     srepositories.StationRepositoryInterface
	routineRepository     srepositories.RoutineRepositoryInterface
	routineTagRepository  srepositories.RoutineTagRepositoryInterface
	routineTaskRepository srepositories.RoutineTaskRepositoryInterface
	itemRepository        srepositories.ItemRepositoryInterface
}

func NewRoutineService(
	validator *validator.Validate,
	db *gorm.DB,
	routineScope sscopes.RoutineScopeInterface,
	stationRepository srepositories.StationRepositoryInterface,
	routineRepository srepositories.RoutineRepositoryInterface,
	routineTagRepository srepositories.RoutineTagRepositoryInterface,
	routineTaskRepository srepositories.RoutineTaskRepositoryInterface,
	itemRepository srepositories.ItemRepositoryInterface,
) RoutineServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &RoutineService{
		validator:             validator,
		db:                    db,
		routineScope:          routineScope,
		stationRepository:     stationRepository,
		routineRepository:     routineRepository,
		routineTagRepository:  routineTagRepository,
		routineTaskRepository: routineTaskRepository,
		itemRepository:        itemRepository,
	}
}

/* ============================== Auxiliary Functions ============================== */

func (s *RoutineService) filterReadableRoutineItems(
	ctx context.Context,
	userId uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
	routines []sschemas.Routine,
) (map[stypes.Pair[uuid.UUID, cenums.ItemType]]struct{}, *cexceptions.Exception) {
	itemIdentitySet := make(map[stypes.Pair[uuid.UUID, cenums.ItemType]]struct{})
	for _, routine := range routines {
		for _, routineToItem := range routine.RoutinesToItems {
			itemIdentitySet[stypes.Pair[uuid.UUID, cenums.ItemType]{
				First:  routineToItem.ItemId,
				Second: routineToItem.ItemType,
			}] = struct{}{}
		}
	}

	itemIdentities := make([]stypes.Pair[uuid.UUID, cenums.ItemType], 0, len(itemIdentitySet))
	for itemIdentity := range itemIdentitySet {
		itemIdentities = append(itemIdentities, itemIdentity)
	}
	permittedItemIdentities, exception := s.itemRepository.GetPermittedIdentities(
		itemIdentities,
		userId,
		allowedPermissions,
		srepositories.WithDB(s.db.WithContext(ctx)),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return nil, exception
	}

	permittedItemIdentitySet := make(map[stypes.Pair[uuid.UUID, cenums.ItemType]]struct{}, len(permittedItemIdentities))
	for _, itemIdentity := range permittedItemIdentities {
		permittedItemIdentitySet[itemIdentity] = struct{}{}
	}

	return permittedItemIdentitySet, nil
}

func (s *RoutineService) visualizeMyRoutineTimeCount(
	ctx context.Context,
	userId uuid.UUID,
	permission cenums.AccessControlPermission,
	timeHourUnit int,
	queryRangeStartedAt time.Time,
	queryRangeEndedAt time.Time,
	columnName string,
	fieldName string,
) ([]capi.RoutineCountDatum, *cexceptions.Exception) {
	db := s.db.WithContext(ctx)
	var buckets []struct {
		BucketStart  time.Time `gorm:"column:bucket_start;"`
		RoutineCount int64     `gorm:"column:routine_count;"`
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
			COUNT(uts.station_id) AS routine_count
		`).
		Joins(
			`LEFT JOIN "RoutineTable" routine
				ON routine.`+columnName+` >= buckets.bucket_start
				AND routine.`+columnName+` < buckets.bucket_start + ?::integer * interval '1 hour'
				AND routine.deleted_at IS NULL`,
			timeHourUnit,
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
		return nil, apiexceptions.NewRoutineException().NotFound().WithOrigin(err)
	}

	data := make([]capi.RoutineCountDatum, len(buckets))
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
			return nil, apiexceptions.NewRoutineException().FailedToMarshalData(metadata)
		}

		data[index] = capi.RoutineCountDatum{
			Id:    bucket.BucketStart.Format(time.RFC3339),
			X:     x,
			Value: bucket.RoutineCount,
			Meta:  meta,
		}
	}

	return data, nil
}

/* ============================== Service Methods for Routine ============================== */

/* ============================== Main Methods ============================== */

func (s *RoutineService) GetMyRoutineById(
	ctx context.Context, reqDto *capi.GetMyRoutineByIdRequestDto,
) (*capi.GetMyRoutineByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	onlyDeleted := stypes.Ternary_Neutral
	if reqDto.Param.IsDeleted != nil {
		if *reqDto.Param.IsDeleted {
			onlyDeleted = stypes.Ternary_Positive
		} else {
			onlyDeleted = stypes.Ternary_Negative
		}
	}

	routine, exception := s.routineRepository.GetOneById(
		reqDto.Param.RoutineId,
		actorUserId,
		[]sschemas.RoutineRelation{
			sschemas.RoutineRelation_RoutinesToTags,
			sschemas.RoutineRelation_RoutineTasks,
			sschemas.RoutineRelation_RoutinesToItems,
		},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(onlyDeleted),
	)
	if exception != nil {
		return nil, exception
	}
	permittedItemIdentitySet, exception := s.filterReadableRoutineItems(
		ctx,
		actorUserId,
		allowedPermissions,
		[]sschemas.Routine{*routine},
	)
	if exception != nil {
		return nil, exception
	}

	tagIds := make([]uuid.UUID, len(routine.RoutinesToTags))
	for index, routineToTag := range routine.RoutinesToTags {
		tagIds[index] = routineToTag.TagId
	}
	taskIds := make([]uuid.UUID, len(routine.RoutineTasks))
	for index, routineTask := range routine.RoutineTasks {
		taskIds[index] = routineTask.Id
	}
	itemIds := make([]uuid.UUID, 0, len(routine.RoutinesToItems))
	for _, routineToItem := range routine.RoutinesToItems {
		if _, exists := permittedItemIdentitySet[stypes.Pair[uuid.UUID, cenums.ItemType]{
			First:  routineToItem.ItemId,
			Second: routineToItem.ItemType,
		}]; exists {
			itemIds = append(itemIds, routineToItem.ItemId)
		}
	}

	return &capi.GetMyRoutineByIdResponseDto{
		Id:               routine.Id,
		StationId:        routine.StationId,
		Title:            routine.Title,
		Description:      routine.Description,
		Status:           routine.Status,
		IsPinned:         routine.IsPinned,
		ScheduledStartAt: routine.ScheduledStartAt,
		ScheduledEndAt:   routine.ScheduledEndAt,
		Period:           routine.Period,
		Timezone:         routine.Timezone,
		DeletedAt:        routine.DeletedAt,
		UpdatedAt:        routine.UpdatedAt,
		CreatedAt:        routine.CreatedAt,
		TagIds:           tagIds,
		TaskIds:          taskIds,
		ItemIds:          itemIds,
	}, nil
}

func (s *RoutineService) GetMyRoutinesByStationId(
	ctx context.Context, reqDto *capi.GetMyRoutinesByStationIdRequestDto,
) (*capi.GetMyRoutinesByStationIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	onlyDeleted := stypes.Ternary_Neutral
	if reqDto.Param.AreDeleted != nil {
		if *reqDto.Param.AreDeleted {
			onlyDeleted = stypes.Ternary_Positive
		} else {
			onlyDeleted = stypes.Ternary_Negative
		}
	}

	var routines []sschemas.Routine
	query := db.Model(&sschemas.Routine{}).
		Select(`"RoutineTable".*`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = "RoutineTable".station_id`).
		Joins(`INNER JOIN "StationTable" station ON station.id = "RoutineTable".station_id AND station.deleted_at IS NULL`).
		Where(`"RoutineTable".station_id = ?`, reqDto.Param.StationId).
		Where("uts.user_id = ? AND uts.permission IN ?", actorUserId, allowedPermissions).
		Scopes(s.routineScope.IncludePreloads(
			[]sschemas.RoutineRelation{
				sschemas.RoutineRelation_RoutinesToTags,
				sschemas.RoutineRelation_RoutineTasks,
				sschemas.RoutineRelation_RoutinesToItems,
			},
			&actorUserId,
		))

	query = query.Scopes(s.routineScope.FilterOnlyDeleted(onlyDeleted))

	result := query.Order(`"RoutineTable".scheduled_start_at ASC`).
		Order(`"RoutineTable".scheduled_end_at ASC`).
		Order(`"RoutineTable".id ASC`).
		Find(&routines)
	if result.Error != nil {
		return nil, apiexceptions.NewRoutineException().NotFound().WithOrigin(result.Error)
	}
	permittedItemIdentitySet, exception := s.filterReadableRoutineItems(
		ctx,
		actorUserId,
		allowedPermissions,
		routines,
	)
	if exception != nil {
		return nil, exception
	}

	resDto := make(capi.GetMyRoutinesByStationIdResponseDto, len(routines))
	for index, routine := range routines {
		tagIds := make([]uuid.UUID, len(routine.RoutinesToTags))
		for index, routineToTag := range routine.RoutinesToTags {
			tagIds[index] = routineToTag.TagId
		}
		taskIds := make([]uuid.UUID, len(routine.RoutineTasks))
		for index, routineTask := range routine.RoutineTasks {
			taskIds[index] = routineTask.Id
		}
		itemIds := make([]uuid.UUID, 0, len(routine.RoutinesToItems))
		for _, routineToItem := range routine.RoutinesToItems {
			if _, exists := permittedItemIdentitySet[stypes.Pair[uuid.UUID, cenums.ItemType]{
				First:  routineToItem.ItemId,
				Second: routineToItem.ItemType,
			}]; exists {
				itemIds = append(itemIds, routineToItem.ItemId)
			}
		}
		resDto[index] = capi.RoutineResponseDto{
			Id:               routine.Id,
			StationId:        routine.StationId,
			Title:            routine.Title,
			Status:           routine.Status,
			IsPinned:         routine.IsPinned,
			ScheduledStartAt: routine.ScheduledStartAt,
			ScheduledEndAt:   routine.ScheduledEndAt,
			Period:           routine.Period,
			Timezone:         routine.Timezone,
			DeletedAt:        routine.DeletedAt,
			UpdatedAt:        routine.UpdatedAt,
			CreatedAt:        routine.CreatedAt,
			TagIds:           tagIds,
			TaskIds:          taskIds,
			ItemIds:          itemIds,
		}
	}

	return &resDto, nil
}

func (s *RoutineService) GetAllMyRoutinesByTimeRange(
	ctx context.Context, reqDto *capi.GetAllMyRoutinesByTimeRangeRequestDto,
) (*capi.GetAllMyRoutinesByTimeRangeResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}
	if !reqDto.Param.From.Before(reqDto.Param.To) { // make sure from is before to
		return nil, apiexceptions.NewRoutineException().InvalidInput().WithOrigin(fmt.Errorf("from must be before to"))
	}
	if !stimes.IsTimeWithin(reqDto.Param.From, reqDto.Param.To, 360*24*time.Hour) { // make sure the time range is within 360 days which is approximate 1 year
		return nil, apiexceptions.NewRoutineException().QueriedTimeRangeTooLarge(reqDto.Param.From, reqDto.Param.To)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	onlyDeleted := stypes.Ternary_Neutral
	if reqDto.Param.AreDeleted != nil {
		if *reqDto.Param.AreDeleted {
			onlyDeleted = stypes.Ternary_Positive
		} else {
			onlyDeleted = stypes.Ternary_Negative
		}
	}

	routines, exception := s.routineRepository.GetAllByTimeRange(
		reqDto.Param.From,
		reqDto.Param.To,
		reqDto.Param.StationIds,
		actorUserId,
		[]sschemas.RoutineRelation{
			sschemas.RoutineRelation_RoutinesToTags,
			sschemas.RoutineRelation_RoutineTasks,
			sschemas.RoutineRelation_RoutinesToItems,
		},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(onlyDeleted),
	)
	if exception != nil {
		return nil, exception
	}
	permittedItemIdentitySet, exception := s.filterReadableRoutineItems(
		ctx,
		actorUserId,
		allowedPermissions,
		routines,
	)
	if exception != nil {
		return nil, exception
	}

	resDto := make(capi.GetAllMyRoutinesByTimeRangeResponseDto, len(routines))
	for index, routine := range routines {
		tagIds := make([]uuid.UUID, len(routine.RoutinesToTags))
		for index, routineToTag := range routine.RoutinesToTags {
			tagIds[index] = routineToTag.TagId
		}
		taskIds := make([]uuid.UUID, len(routine.RoutineTasks))
		for index, routineTask := range routine.RoutineTasks {
			taskIds[index] = routineTask.Id
		}
		itemIds := make([]uuid.UUID, 0, len(routine.RoutinesToItems))
		for _, routineToItem := range routine.RoutinesToItems {
			if _, exists := permittedItemIdentitySet[stypes.Pair[uuid.UUID, cenums.ItemType]{
				First:  routineToItem.ItemId,
				Second: routineToItem.ItemType,
			}]; exists {
				itemIds = append(itemIds, routineToItem.ItemId)
			}
		}
		resDto[index] = capi.RoutineResponseDto{
			Id:               routine.Id,
			StationId:        routine.StationId,
			Title:            routine.Title,
			Status:           routine.Status,
			IsPinned:         routine.IsPinned,
			ScheduledStartAt: routine.ScheduledStartAt,
			ScheduledEndAt:   routine.ScheduledEndAt,
			Period:           routine.Period,
			Timezone:         routine.Timezone,
			DeletedAt:        routine.DeletedAt,
			UpdatedAt:        routine.UpdatedAt,
			CreatedAt:        routine.CreatedAt,
			TagIds:           tagIds,
			TaskIds:          taskIds,
			ItemIds:          itemIds,
		}
	}

	return &resDto, nil
}

func (s *RoutineService) CreateRoutineByStationId(
	ctx context.Context, reqDto *capi.CreateRoutineByStationIdRequestDto,
) (*capi.CreateRoutineByStationIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	newRoutineId, exception := s.routineRepository.CreateOneByStationId(
		reqDto.Body.StationId,
		actorUserId,
		sinputs.CreateRoutineInput{
			Id:               reqDto.Body.Id,
			Title:            reqDto.Body.Title,
			Description:      reqDto.Body.Description,
			Status:           reqDto.Body.Status,
			IsPinned:         reqDto.Body.IsPinned,
			ScheduledStartAt: reqDto.Body.ScheduledStartAt,
			ScheduledEndAt:   reqDto.Body.ScheduledEndAt,
			Period:           reqDto.Body.Period,
			Timezone:         reqDto.Body.Timezone,
		},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.CreateRoutineByStationIdResponseDto{
		Id:        *newRoutineId,
		CreatedAt: time.Now(),
	}, nil
}

func (s *RoutineService) CreateRoutinesByStationIds(
	ctx context.Context, reqDto *capi.CreateRoutinesByStationIdsRequestDto,
) (*capi.CreateRoutinesByStationIdsResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]sinputs.CreateRoutineByStationIdInput, len(reqDto.Body.CreatedRoutines))
	for index, createdRoutine := range reqDto.Body.CreatedRoutines {
		input[index] = sinputs.CreateRoutineByStationIdInput{
			Id:               createdRoutine.Id,
			StationId:        createdRoutine.StationId,
			Title:            createdRoutine.Title,
			Description:      createdRoutine.Description,
			Status:           createdRoutine.Status,
			IsPinned:         createdRoutine.IsPinned,
			ScheduledStartAt: createdRoutine.ScheduledStartAt,
			ScheduledEndAt:   createdRoutine.ScheduledEndAt,
			Period:           createdRoutine.Period,
			Timezone:         createdRoutine.Timezone,
		}
	}
	newRoutineIds, exception := s.routineRepository.CreateManyByStationIds(
		actorUserId,
		input,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.CreateRoutinesByStationIdsResponseDto{
		Ids:       newRoutineIds,
		CreatedAt: time.Now(),
	}, nil
}

func (s *RoutineService) UpdateMyRoutineById(
	ctx context.Context, reqDto *capi.UpdateMyRoutineByIdRequestDto,
) (*capi.UpdateMyRoutineByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	updatedRoutine, exception := s.routineRepository.UpdateOneById(
		reqDto.Body.RoutineId,
		actorUserId,
		sinputs.PartialUpdateRoutineInput{
			Values: sinputs.UpdateRoutineInput{
				StationId:        reqDto.Body.Values.StationId,
				Title:            reqDto.Body.Values.Title,
				Description:      reqDto.Body.Values.Description,
				Status:           reqDto.Body.Values.Status,
				IsPinned:         reqDto.Body.Values.IsPinned,
				ScheduledStartAt: reqDto.Body.Values.ScheduledStartAt,
				ScheduledEndAt:   reqDto.Body.Values.ScheduledEndAt,
				Period:           reqDto.Body.Values.Period,
				Timezone:         reqDto.Body.Values.Timezone,
			},
			SetNull: reqDto.Body.SetNull,
		},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.UpdateMyRoutineByIdResponseDto{
		UpdatedAt: updatedRoutine.UpdatedAt,
	}, nil
}

func (s *RoutineService) UpdateMyRoutinesByIds(
	ctx context.Context, reqDto *capi.UpdateMyRoutinesByIdsRequestDto,
) (*capi.UpdateMyRoutinesByIdsResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]sinputs.UpdateRoutineByIdInput, len(reqDto.Body.UpdatedRoutines))
	for index, updatedRoutine := range reqDto.Body.UpdatedRoutines {
		input[index] = sinputs.UpdateRoutineByIdInput{
			Id: updatedRoutine.RoutineId,
			PartialUpdateInput: sinputs.PartialUpdateInput[sinputs.UpdateRoutineInput]{
				Values: sinputs.UpdateRoutineInput{
					StationId:        updatedRoutine.Values.StationId,
					Title:            updatedRoutine.Values.Title,
					Description:      updatedRoutine.Values.Description,
					Status:           updatedRoutine.Values.Status,
					IsPinned:         updatedRoutine.Values.IsPinned,
					ScheduledStartAt: updatedRoutine.Values.ScheduledStartAt,
					ScheduledEndAt:   updatedRoutine.Values.ScheduledEndAt,
					Period:           updatedRoutine.Values.Period,
					Timezone:         updatedRoutine.Values.Timezone,
				},
				SetNull: updatedRoutine.SetNull,
			},
		}
	}
	exception = s.routineRepository.UpdateManyByIds(
		actorUserId,
		input,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.UpdateMyRoutinesByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *RoutineService) LinkRoutineTagById(
	ctx context.Context, reqDto *capi.LinkRoutineTagByIdRequestDto,
) (*capi.LinkRoutineTagByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	routine, exception := s.routineRepository.CheckPermissionAndGetOneById(
		reqDto.Body.RoutineId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineException().NoPermission("get the routine")
	}

	if _, exception := s.routineTagRepository.GetOneById(
		reqDto.Body.RoutineTagId,
		actorUserId,
		nil,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}

	var newRoutinesToTags sschemas.RoutinesToTags
	newRoutinesToTags.RoutineId = reqDto.Body.RoutineId
	newRoutinesToTags.TagId = reqDto.Body.RoutineTagId
	newRoutinesToTags.UserId = actorUserId
	newRoutinesToTags.StationId = routine.StationId

	var result *gorm.DB
	if reqDto.Body.IsUnlink {
		result = tx.Model(&sschemas.RoutinesToTags{}).
			Where(
				"routine_id = ? AND tag_id = ? AND user_id = ?",
				newRoutinesToTags.RoutineId,
				newRoutinesToTags.TagId,
				newRoutinesToTags.UserId,
			).
			Delete(&sschemas.RoutinesToTags{})
	} else {
		result = tx.Model(&sschemas.RoutinesToTags{}).
			Create(&newRoutinesToTags)
	}
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineException().FailedToLinkRoutineTags().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineException().NoChanges()},
	}); exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.LinkRoutineTagByIdResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *RoutineService) LinkRoutineTagsByIds(
	ctx context.Context, reqDto *capi.LinkRoutineTagsByIdsRequestDto,
) (*capi.LinkRoutineTagsByIdsResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	isRoutineExist := make(map[uuid.UUID]bool)
	isRoutineTagExist := make(map[uuid.UUID]bool)
	var routineIds []uuid.UUID
	var routineTagIds []uuid.UUID
	for _, linkedRoutineAndTag := range reqDto.Body.LinkedRoutinesAndTags {
		if !isRoutineExist[linkedRoutineAndTag.RoutineId] {
			isRoutineExist[linkedRoutineAndTag.RoutineId] = true
			routineIds = append(routineIds, linkedRoutineAndTag.RoutineId)
		}
		if !isRoutineTagExist[linkedRoutineAndTag.RoutineTagId] {
			isRoutineTagExist[linkedRoutineAndTag.RoutineTagId] = true
			routineTagIds = append(routineTagIds, linkedRoutineAndTag.RoutineTagId)
		}
	}

	validRoutineStationIds := make(map[uuid.UUID]uuid.UUID)
	validRoutines, exception := s.routineRepository.CheckPermissionsAndGetManyByIds(
		routineIds,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	for _, validRoutine := range validRoutines {
		validRoutineStationIds[validRoutine.Id] = validRoutine.StationId
	}

	isRoutineTagValid := make(map[uuid.UUID]bool)
	validRoutineTags, exception := s.routineTagRepository.GetManyByIds(
		routineTagIds,
		actorUserId,
		nil,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	for _, validRoutineTag := range validRoutineTags {
		isRoutineTagValid[validRoutineTag.Id] = true
	}

	var newRoutinesToTags []sschemas.RoutinesToTags
	for _, linkedRoutineAndTag := range reqDto.Body.LinkedRoutinesAndTags {
		stationId, isRoutineValid := validRoutineStationIds[linkedRoutineAndTag.RoutineId]
		if !isRoutineValid ||
			!isRoutineTagValid[linkedRoutineAndTag.RoutineTagId] {
			continue
		}
		newRoutinesToTags = append(newRoutinesToTags, sschemas.RoutinesToTags{
			RoutineId: linkedRoutineAndTag.RoutineId,
			TagId:     linkedRoutineAndTag.RoutineTagId,
			UserId:    actorUserId,
			StationId: stationId,
		})
	}
	if len(newRoutinesToTags) == 0 {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineException().NoChanges()
	}

	values := make([][]any, len(newRoutinesToTags))
	for index, newRoutineToTag := range newRoutinesToTags {
		values[index] = []any{newRoutineToTag.RoutineId, newRoutineToTag.TagId, newRoutineToTag.UserId}
	}

	var result *gorm.DB
	if reqDto.Body.IsUnlink {
		result = tx.Model(&sschemas.RoutinesToTags{}).
			Where("(routine_id, tag_id, user_id) IN ?", values).
			Delete(&sschemas.RoutinesToTags{})
	} else {
		result = tx.Model(&sschemas.RoutinesToTags{}).
			Create(&newRoutinesToTags)
	}
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineException().FailedToLinkRoutineTags().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineException().NoChanges()},
	}); exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.LinkRoutineTagsByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *RoutineService) LinkRoutineItemById(
	ctx context.Context, reqDto *capi.LinkRoutineItemByIdRequestDto,
) (*capi.LinkRoutineItemByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	if !s.routineRepository.HasPermission(
		reqDto.Body.RoutineId,
		actorUserId,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	) {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineException().NoPermission("get the routine")
	}

	if !s.itemRepository.HasPermission(
		reqDto.Body.ItemId,
		cenums.ItemType(reqDto.Body.ItemType),
		actorUserId,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	) {
		tx.Rollback()
		return nil, apiexceptions.NewItemException().NoPermission("get the item")
	}

	var newRoutinesToItems sschemas.RoutinesToItems
	newRoutinesToItems.RoutineId = reqDto.Body.RoutineId
	newRoutinesToItems.ItemId = reqDto.Body.ItemId
	newRoutinesToItems.ItemType = cenums.ItemType(reqDto.Body.ItemType)

	var result *gorm.DB
	if reqDto.Body.IsUnlink {
		result = tx.Model(&sschemas.RoutinesToItems{}).
			Where(
				"routine_id = ? AND item_id = ? AND item_type = ?",
				newRoutinesToItems.RoutineId,
				newRoutinesToItems.ItemId,
				newRoutinesToItems.ItemType,
			).
			Delete(&sschemas.RoutinesToItems{})
	} else {
		result = tx.Model(&sschemas.RoutinesToItems{}).
			Create(&newRoutinesToItems)
	}
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineException().FailedToLinkItems().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineException().NoChanges()},
	}); exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.LinkRoutineItemByIdResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *RoutineService) LinkRoutineItemsByIds(
	ctx context.Context, reqDto *capi.LinkRoutineItemsByIdsRequestDto,
) (*capi.LinkRoutineItemsByIdsResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	isRoutineExist := make(map[uuid.UUID]bool)
	isItemExist := make(map[stypes.Pair[uuid.UUID, cenums.ItemType]]bool)
	var routineIds []uuid.UUID
	var itemIdentities []stypes.Pair[uuid.UUID, cenums.ItemType]
	for _, linkedRoutineAndItem := range reqDto.Body.LinkedRoutinesAndItems {
		if !isRoutineExist[linkedRoutineAndItem.RoutineId] {
			isRoutineExist[linkedRoutineAndItem.RoutineId] = true
			routineIds = append(routineIds, linkedRoutineAndItem.RoutineId)
		}
		itemIdentity := stypes.Pair[uuid.UUID, cenums.ItemType]{
			First:  linkedRoutineAndItem.ItemId,
			Second: cenums.ItemType(linkedRoutineAndItem.ItemType),
		}
		if !isItemExist[itemIdentity] {
			isItemExist[itemIdentity] = true
			itemIdentities = append(itemIdentities, itemIdentity)
		}
	}

	isRoutineValid := make(map[uuid.UUID]bool)
	validRoutines, exception := s.routineRepository.CheckPermissionsAndGetManyByIds(
		routineIds,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	for _, validRoutine := range validRoutines {
		isRoutineValid[validRoutine.Id] = true
	}

	isItemValid := make(map[stypes.Pair[uuid.UUID, cenums.ItemType]]bool)
	validItems, exception := s.itemRepository.CheckPermissionsAndGetManyByIds(
		itemIdentities,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	for _, validItem := range validItems {
		isItemValid[stypes.Pair[uuid.UUID, cenums.ItemType]{
			First:  validItem.Id,
			Second: validItem.Type,
		}] = true
	}

	var newRoutinesToItems []sschemas.RoutinesToItems
	for _, linkedRoutineAndItem := range reqDto.Body.LinkedRoutinesAndItems {
		itemIdentity := stypes.Pair[uuid.UUID, cenums.ItemType]{
			First:  linkedRoutineAndItem.ItemId,
			Second: cenums.ItemType(linkedRoutineAndItem.ItemType),
		}
		if !isRoutineValid[linkedRoutineAndItem.RoutineId] ||
			!isItemValid[itemIdentity] {
			continue
		}
		newRoutinesToItems = append(newRoutinesToItems, sschemas.RoutinesToItems{
			RoutineId: linkedRoutineAndItem.RoutineId,
			ItemId:    linkedRoutineAndItem.ItemId,
			ItemType:  cenums.ItemType(linkedRoutineAndItem.ItemType),
		})
	}
	if len(newRoutinesToItems) == 0 {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineException().NoChanges()
	}

	values := make([][]any, len(newRoutinesToItems))
	for index, newRoutineToItem := range newRoutinesToItems {
		values[index] = []any{newRoutineToItem.RoutineId, newRoutineToItem.ItemId, newRoutineToItem.ItemType}
	}

	var result *gorm.DB
	if reqDto.Body.IsUnlink {
		result = tx.Model(&sschemas.RoutinesToItems{}).
			Where("(routine_id, item_id, item_type) IN ?", values).
			Delete(&sschemas.RoutinesToItems{})
	} else {
		result = tx.Model(&sschemas.RoutinesToItems{}).
			Create(&newRoutinesToItems)
	}
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineException().FailedToLinkItems().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineException().NoChanges()},
	}); exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.LinkRoutineItemsByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *RoutineService) RestoreMyRoutineById(
	ctx context.Context, reqDto *capi.RestoreMyRoutineByIdRequestDto,
) (*capi.RestoreMyRoutineByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	restoredRoutine, exception := s.routineRepository.RestoreSoftDeletedOneById(
		reqDto.Body.RoutineId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.RestoreMyRoutineByIdResponseDto{
		Id:               restoredRoutine.Id,
		StationId:        restoredRoutine.StationId,
		Title:            restoredRoutine.Title,
		Description:      restoredRoutine.Description,
		Status:           restoredRoutine.Status,
		IsPinned:         restoredRoutine.IsPinned,
		ScheduledStartAt: restoredRoutine.ScheduledStartAt,
		ScheduledEndAt:   restoredRoutine.ScheduledEndAt,
		Period:           restoredRoutine.Period,
		Timezone:         restoredRoutine.Timezone,
		DeletedAt:        restoredRoutine.DeletedAt,
		UpdatedAt:        restoredRoutine.UpdatedAt,
		CreatedAt:        restoredRoutine.CreatedAt,
	}, nil
}

func (s *RoutineService) RestoreMyRoutinesByIds(
	ctx context.Context, reqDto *capi.RestoreMyRoutinesByIdsRequestDto,
) (*capi.RestoreMyRoutinesByIdsResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	restoredRoutines, exception := s.routineRepository.RestoreSoftDeletedManyByIds(
		reqDto.Body.RoutineIds,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	resDto := capi.RestoreMyRoutinesByIdsResponseDto{}
	for _, restoredRoutine := range restoredRoutines {
		resDto = append(resDto, capi.RestoreMyRoutineByIdResponseDto{
			Id:               restoredRoutine.Id,
			StationId:        restoredRoutine.StationId,
			Title:            restoredRoutine.Title,
			Description:      restoredRoutine.Description,
			Status:           restoredRoutine.Status,
			IsPinned:         restoredRoutine.IsPinned,
			ScheduledStartAt: restoredRoutine.ScheduledStartAt,
			ScheduledEndAt:   restoredRoutine.ScheduledEndAt,
			Period:           restoredRoutine.Period,
			Timezone:         restoredRoutine.Timezone,
			DeletedAt:        restoredRoutine.DeletedAt,
			UpdatedAt:        restoredRoutine.UpdatedAt,
			CreatedAt:        restoredRoutine.CreatedAt,
		})
	}

	return &resDto, nil
}

func (s *RoutineService) DeleteMyRoutineById(
	ctx context.Context, reqDto *capi.DeleteMyRoutineByIdRequestDto,
) (*capi.DeleteMyRoutineByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.routineRepository.SoftDeleteOneById(
		reqDto.Body.RoutineId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.DeleteMyRoutineByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *RoutineService) DeleteMyRoutinesByIds(
	ctx context.Context, reqDto *capi.DeleteMyRoutinesByIdsRequestDto,
) (*capi.DeleteMyRoutinesByIdsResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.routineRepository.SoftDeleteManyByIds(
		reqDto.Body.RoutineIds,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.DeleteMyRoutinesByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *RoutineService) HardDeleteMyRoutineById(
	ctx context.Context, reqDto *capi.HardDeleteMyRoutineByIdRequestDto,
) (*capi.HardDeleteMyRoutineByIdResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.routineRepository.HardDeleteOneById(
		reqDto.Body.RoutineId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.HardDeleteMyRoutineByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *RoutineService) HardDeleteMyRoutinesByIds(
	ctx context.Context, reqDto *capi.HardDeleteMyRoutinesByIdsRequestDto,
) (*capi.HardDeleteMyRoutinesByIdsResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.routineRepository.HardDeleteManyByIds(
		reqDto.Body.RoutineIds,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.HardDeleteMyRoutinesByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

/* ============================== Service Methods for Charts ============================== */

func (s *RoutineService) VisualizeMyRoutineStatusCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutineStatusCountRequestDto,
) (*capi.VisualizeMyRoutineStatusCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	var counts struct {
		ScheduledCount  int64 `gorm:"column:scheduled_count;"`
		InProgressCount int64 `gorm:"column:in_progress_count;"`
		CompletedCount  int64 `gorm:"column:completed_count;"`
		OverDueCount    int64 `gorm:"column:over_due_count;"`
	}
	result := db.Model(&sschemas.Routine{}).
		Select(`
			COUNT(*) FILTER (WHERE status = ?) as scheduled_count,
			COUNT(*) FILTER (WHERE status = ?) as in_progress_count,
			COUNT(*) FILTER (WHERE status = ?) as completed_count,
			COUNT(*) FILTER (WHERE status = ?) as over_due_count
		`,
			cenums.RoutineStatus_Scheduled,
			cenums.RoutineStatus_InProgress,
			cenums.RoutineStatus_Completed,
			cenums.RoutineStatus_OverDue,
		).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = "RoutineTable".station_id`).
		Where("uts.user_id = ? AND uts.permission = ?", actorUserId, reqDto.Param.Permission).
		Where(`"RoutineTable".deleted_at IS NULL`).
		Scan(&counts)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewRoutineException().NotFound().WithOrigin(err)
	}

	scheduledRoutineMetadata := map[string]string{"status": "scheduled"}
	scheduledRoutineMeta, err := json.Marshal(scheduledRoutineMetadata)
	if err != nil {
		return nil, apiexceptions.NewRoutineException().FailedToMarshalData(scheduledRoutineMetadata)
	}

	inProgressRoutineMetadata := map[string]string{"status": "inProgress"}
	inProgressRoutineMeta, err := json.Marshal(inProgressRoutineMetadata)
	if err != nil {
		return nil, apiexceptions.NewRoutineException().FailedToMarshalData(inProgressRoutineMetadata)
	}

	completedRoutineMetadata := map[string]string{"status": "completed"}
	completedRoutineMeta, err := json.Marshal(completedRoutineMetadata)
	if err != nil {
		return nil, apiexceptions.NewRoutineException().FailedToMarshalData(completedRoutineMetadata)
	}

	overDueRoutineMetadata := map[string]string{"status": "overDue"}
	overDueRoutineMeta, err := json.Marshal(overDueRoutineMetadata)
	if err != nil {
		return nil, apiexceptions.NewRoutineException().FailedToMarshalData(overDueRoutineMetadata)
	}

	return &capi.VisualizeMyRoutineStatusCountResponseDto{
		Data: []capi.RoutineCountDatum{
			capi.RoutineCountDatum{
				Id:    "scheduled-routine-count",
				X:     "Scheduled Routine Count",
				Value: counts.ScheduledCount,
				Meta:  scheduledRoutineMeta,
			},
			capi.RoutineCountDatum{
				Id:    "in-progress-routine-count",
				X:     "In Progress Routine Count",
				Value: counts.InProgressCount,
				Meta:  inProgressRoutineMeta,
			},
			capi.RoutineCountDatum{
				Id:    "completed-routine-count",
				X:     "Completed Routine Count",
				Value: counts.CompletedCount,
				Meta:  completedRoutineMeta,
			},
			capi.RoutineCountDatum{
				Id:    "over-due-routine-count",
				X:     "Over Due Routine Count",
				Value: counts.OverDueCount,
				Meta:  overDueRoutineMeta,
			},
		},
	}, nil
}

func (s *RoutineService) VisualizeMyRoutinePeriodCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutinePeriodCountRequestDto,
) (*capi.VisualizeMyRoutinePeriodCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	var counts struct {
		DailyCount   int64 `gorm:"column:daily_count;"`
		WeeklyCount  int64 `gorm:"column:weekly_count;"`
		MonthlyCount int64 `gorm:"column:monthly_count;"`
	}
	result := db.Model(&sschemas.Routine{}).
		Select(`
			COUNT(*) FILTER (WHERE period = ?) as daily_count,
			COUNT(*) FILTER (WHERE period = ?) as weekly_count,
			COUNT(*) FILTER (WHERE period = ?) as monthly_count
		`,
			cenums.RoutinePeriod_Daily,
			cenums.RoutinePeriod_Weekly,
			cenums.RoutinePeriod_Monthly,
		).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = "RoutineTable".station_id`).
		Where("uts.user_id = ? AND uts.permission = ?", actorUserId, reqDto.Param.Permission).
		Where(`"RoutineTable".deleted_at IS NULL`).
		Scan(&counts)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewRoutineException().NotFound().WithOrigin(err)
	}

	dailyRoutineMetadata := map[string]string{"period": "daily"}
	dailyRoutineMeta, err := json.Marshal(dailyRoutineMetadata)
	if err != nil {
		return nil, apiexceptions.NewRoutineException().FailedToMarshalData(dailyRoutineMetadata)
	}

	weeklyRoutineMetadata := map[string]string{"period": "daily"}
	weeklyRoutineMeta, err := json.Marshal(weeklyRoutineMetadata)
	if err != nil {
		return nil, apiexceptions.NewRoutineException().FailedToMarshalData(weeklyRoutineMetadata)
	}

	monthlyRoutineMetadata := map[string]string{"period": "daily"}
	monthlyRoutineMeta, err := json.Marshal(monthlyRoutineMetadata)
	if err != nil {
		return nil, apiexceptions.NewRoutineException().FailedToMarshalData(monthlyRoutineMetadata)
	}

	return &capi.VisualizeMyRoutinePeriodCountResponseDto{
		Data: []capi.RoutineCountDatum{
			capi.RoutineCountDatum{
				Id:    "daily-routine-count",
				X:     "Daily Routine Count",
				Value: counts.DailyCount,
				Meta:  dailyRoutineMeta,
			},
			capi.RoutineCountDatum{
				Id:    "weekly-routine-count",
				X:     "Weekly Routine Count",
				Value: counts.WeeklyCount,
				Meta:  weeklyRoutineMeta,
			},
			capi.RoutineCountDatum{
				Id:    "monthly-routine-count",
				X:     "Monthly Routine Count",
				Value: counts.MonthlyCount,
				Meta:  monthlyRoutineMeta,
			},
		},
	}, nil
}

func (s *RoutineService) VisualizeMyRoutineScheduledStartAtCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutineScheduledStartAtCountRequestDto,
) (*capi.VisualizeMyRoutineScheduledStartAtCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}
	if !reqDto.Param.QueryRangeStartedAt.Before(reqDto.Param.QueryRangeEndedAt) {
		return nil, apiexceptions.NewRoutineException().InvalidDto("queryRangeStartedAt should be earlier then queryRangeEndedAt")
	}
	if !stimes.IsTimeWithin(reqDto.Param.QueryRangeStartedAt, reqDto.Param.QueryRangeEndedAt, 360*24*time.Hour) {
		return nil, apiexceptions.NewRoutineException().QueriedTimeRangeTooLarge(reqDto.Param.QueryRangeStartedAt, reqDto.Param.QueryRangeEndedAt)
	}

	data, exception := s.visualizeMyRoutineTimeCount(
		ctx,
		actorUserId,
		cenums.AccessControlPermission(reqDto.Param.Permission),
		reqDto.Param.TimeHourUnit,
		reqDto.Param.QueryRangeStartedAt,
		reqDto.Param.QueryRangeEndedAt,
		"scheduled_start_at",
		"scheduledStartAt",
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.VisualizeMyRoutineScheduledStartAtCountResponseDto{
		Data: data,
	}, nil
}

func (s *RoutineService) VisualizeMyRoutineScheduledEndAtCount(
	ctx context.Context, reqDto *capi.VisualizeMyRoutineScheduledEndAtCountRequestDto,
) (*capi.VisualizeMyRoutineScheduledEndAtCountResponseDto, *cexceptions.Exception) {
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewRoutineException().InvalidDto().WithOrigin(err)
	}
	if !reqDto.Param.QueryRangeStartedAt.Before(reqDto.Param.QueryRangeEndedAt) {
		return nil, apiexceptions.NewRoutineException().InvalidDto("queryRangeStartedAt should be earlier then queryRangeEndedAt")
	}
	if !stimes.IsTimeWithin(reqDto.Param.QueryRangeStartedAt, reqDto.Param.QueryRangeEndedAt, 360*24*time.Hour) {
		return nil, apiexceptions.NewRoutineException().QueriedTimeRangeTooLarge(reqDto.Param.QueryRangeStartedAt, reqDto.Param.QueryRangeEndedAt)
	}

	data, exception := s.visualizeMyRoutineTimeCount(
		ctx,
		actorUserId,
		cenums.AccessControlPermission(reqDto.Param.Permission),
		reqDto.Param.TimeHourUnit,
		reqDto.Param.QueryRangeStartedAt,
		reqDto.Param.QueryRangeEndedAt,
		"scheduled_end_at",
		"scheduledEndAt",
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.VisualizeMyRoutineScheduledEndAtCountResponseDto{
		Data: data,
	}, nil
}

/* ============================== Service Methods for GraphQL Routine ============================== */

func (s *RoutineService) SearchPrivateRoutines(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineInput,
) (*cgqlmodels.SearchRoutineConnection, *cexceptions.Exception) {
	startTime := time.Now()
	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	onlyDeleted := stypes.Ternary_Negative
	if gqlInput.IsDeletedAt != nil && *gqlInput.IsDeletedAt {
		onlyDeleted = stypes.Ternary_Positive
	}

	query := db.Model(&sschemas.Routine{}).
		Select(`"RoutineTable".*, uts.permission AS permission`).
		Joins(`LEFT JOIN "UsersToStationsTable" uts ON "RoutineTable".station_id = uts.station_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions).
		Scopes(s.routineScope.FilterOnlyDeleted(onlyDeleted))

	if len(gqlInput.StationIds) > 0 {
		query = query.Where(
			`"RoutineTable".station_id IN ?`,
			gqlInput.StationIds,
		)
	}

	if len(gqlInput.TagIds) > 0 {
		subQuery := db.
			Session(&gorm.Session{NewDB: true}).
			Model(&sschemas.RoutinesToTags{}).
			Select("1").
			Where(`"RoutinesToTagsTable".routine_id = "RoutineTable".id`).
			Where(`"RoutinesToTagsTable".user_id = ?`, userId).
			Where(`"RoutinesToTagsTable".tag_id IN ?`, gqlInput.TagIds)

		query = query.Where("EXISTS (?)", subQuery)
	}

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"title ILIKE ?",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchRoutineCursorFields](*gqlInput.After)
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToDecode().WithOrigin(err)
		}

		query = query.Where(
			`"RoutineTable".id > ?`,
			searchCursor.Fields.ID,
		)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		var cending string = cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchRoutineSortByTitle:
			query = query.Order("title " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineSortByStatus:
			query = query.Order("status " + cending).
				Order("title " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineSortByScheduledStartAt:
			query = query.Order("scheduled_start_at " + cending).
				Order("title " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineSortByScheduledEndAt:
			query = query.Order("scheduled_end_at " + cending).
				Order("title " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineSortByPeriod:
			query = query.Order("period " + cending).
				Order("title " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("title " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRoutineSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("title " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("title " + cending).
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

	var routines []sschemas.Routine
	if err := query.Scopes(s.routineScope.IncludePreloads(
		[]sschemas.RoutineRelation{
			sschemas.RoutineRelation_RoutinesToTags,
			sschemas.RoutineRelation_RoutineTasks,
			sschemas.RoutineRelation_RoutinesToItems,
		},
		&userId,
	)).Find(&routines).Error; err != nil {
		return nil, apiexceptions.NewRoutineException().NotFound().WithOrigin(err)
	}
	permittedItemIdentitySet, exception := s.filterReadableRoutineItems(
		ctx,
		userId,
		allowedPermissions,
		routines,
	)
	if exception != nil {
		return nil, exception
	}
	for index := range routines {
		filteredRoutineToItems := make([]sschemas.RoutinesToItems, 0, len(routines[index].RoutinesToItems))
		for _, routineToItem := range routines[index].RoutinesToItems {
			if _, exists := permittedItemIdentitySet[stypes.Pair[uuid.UUID, cenums.ItemType]{
				First:  routineToItem.ItemId,
				Second: routineToItem.ItemType,
			}]; exists {
				filteredRoutineToItems = append(filteredRoutineToItems, routineToItem)
			}
		}

		routines[index].RoutinesToItems = filteredRoutineToItems
	}

	hasNextPage := len(routines) > limit
	searchEdges := make([]*cgqlmodels.SearchRoutineEdge, len(routines))

	for index, routine := range routines {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchRoutineCursorFields]{
			Fields: cgqlmodels.SearchRoutineCursorFields{
				ID: routine.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, apiexceptions.NewSearchException().FailedToUnmarshalSearchCursor()
		}

		searchEdges[index] = &cgqlmodels.SearchRoutineEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                routine.ToPrivateSearchableRoutine(),
		}
	}

	searchPageInfo := &cgqlmodels.SearchPageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: false,
	}

	if len(searchEdges) > 0 {
		searchPageInfo.StartEncodedSearchCursor = &searchEdges[0].EncodedSearchCursor
		searchPageInfo.EndEncodedSearchCursor = &searchEdges[len(searchEdges)-1].EncodedSearchCursor
	}

	searchTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	if hasNextPage {
		searchEdges = searchEdges[:limit]
	}

	return &cgqlmodels.SearchRoutineConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
