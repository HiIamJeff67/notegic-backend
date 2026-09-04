package routines

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-records"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	stimes "github.com/HiIamJeff67/notegic-backend/shared/lib/times"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/core/exceptions"
)

type RoutineTaskRecordServiceInterface interface {
	GetMyRoutineTaskRecordsByRoutineTaskId(ctx context.Context, requestDto *capi.GetMyRoutineTaskRecordsByRoutineTaskIdRequestDto) (*capi.GetMyRoutineTaskRecordsByRoutineTaskIdResponseDto, *cexceptions.Exception)

	/* ============================== Visualization Methods ============================== */
	VisualizeMyRoutineTaskRecordStatusCount(ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordStatusCountRequestDto) (*capi.VisualizeMyRoutineTaskRecordStatusCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineTaskRecordPurposeCount(ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordPurposeCountRequestDto) (*capi.VisualizeMyRoutineTaskRecordPurposeCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineTaskRecordScheduledAtCount(ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto) (*capi.VisualizeMyRoutineTaskRecordScheduledAtCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineTaskRecordActualStartedAtCount(ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto) (*capi.VisualizeMyRoutineTaskRecordActualStartedAtCountResponseDto, *cexceptions.Exception)
	VisualizeMyRoutineTaskRecordActualEndedAtCount(ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto) (*capi.VisualizeMyRoutineTaskRecordActualEndedAtCountResponseDto, *cexceptions.Exception)

	/* ============================== GraphQL Methods ============================== */
	SearchPrivateRoutineTaskRecords(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineTaskRecordInput) (*cgqlmodels.SearchRoutineTaskRecordConnection, *cexceptions.Exception)
}

type RoutineTaskRecordService struct {
	validator                   *validator.Validate
	db                          *gorm.DB
	routineTaskRecordRepository srepositories.RoutineTaskRecordRepositoryInterface
	routineTaskException        apiexceptions.RoutineTaskException
	routineException            apiexceptions.RoutineException
	searchException             apiexceptions.SearchException
}

func NewRoutineTaskRecordService(
	validator *validator.Validate,
	db *gorm.DB,
	routineTaskRecordRepository srepositories.RoutineTaskRecordRepositoryInterface,
	routineTaskException apiexceptions.RoutineTaskException,
	routineException apiexceptions.RoutineException,
	searchException apiexceptions.SearchException,
) RoutineTaskRecordServiceInterface {
	return &RoutineTaskRecordService{
		validator:                   validator,
		db:                          db,
		routineTaskRecordRepository: routineTaskRecordRepository,
		routineTaskException:        routineTaskException,
		routineException:            routineException,
		searchException:             searchException,
	}
}

func (s *RoutineTaskRecordService) visualizeMyRoutineTaskRecordTimeCount(
	ctx context.Context,
	userId uuid.UUID,
	permission cenums.AccessControlPermission,
	routineTaskIds []uuid.UUID,
	timeHourUnit int,
	queryRangeStartedAt time.Time,
	queryRangeEndedAt time.Time,
	columnName string,
	fieldName string,
) ([]capi.RoutineTaskRecordCountDatum, *cexceptions.Exception) {
	db := s.db.WithContext(ctx)

	var buckets []struct {
		BucketStart            time.Time `gorm:"column:bucket_start;"`
		RoutineTaskRecordCount int64     `gorm:"column:routine_task_record_count;"`
	}

	recordJoin := `LEFT JOIN "RoutineTaskRecordTable" routine_task_record ON `
	if columnName == "scheduled_at" {
		recordJoin += `EXISTS (
			SELECT 1 FROM "RoutineRecordTable" routine_record
			WHERE routine_record.id = routine_task_record.routine_record_id
				AND routine_record.scheduled_at >= buckets.bucket_start
				AND routine_record.scheduled_at < buckets.bucket_start + ?::integer * interval '1 hour'
		)`
	} else {
		recordJoin += `routine_task_record.` + columnName + ` >= buckets.bucket_start
			AND routine_task_record.` + columnName + ` < buckets.bucket_start + ?::integer * interval '1 hour'`
	}
	recordJoinArgs := []any{timeHourUnit}
	if len(routineTaskIds) > 0 {
		recordJoin += ` AND routine_task_record.routine_task_id IN ?`
		recordJoinArgs = append(recordJoinArgs, routineTaskIds)
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
			COUNT(uts.station_id) AS routine_task_record_count
		`).
		Joins(recordJoin, recordJoinArgs...).
		Joins(`LEFT JOIN "RoutineTaskTable" routine_task ON routine_task.id = routine_task_record.routine_task_id`).
		Joins(`LEFT JOIN "RoutineTable" routine ON routine.id = routine_task.routine_id AND routine.deleted_at IS NULL`).
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
	if result.Error != nil {
		return nil, s.routineTaskException.NotFound().WithOrigin(result.Error)
	}

	data := make([]capi.RoutineTaskRecordCountDatum, len(buckets))
	for index, bucket := range buckets {
		bucketEnd := bucket.BucketStart.Add(time.Duration(timeHourUnit) * time.Hour)
		x := bucket.BucketStart.Format(time.DateOnly)
		if timeHourUnit < 24 {
			x = bucket.BucketStart.Format("2006-01-02 15:04")
		}

		metadata := map[string]any{
			"bucketStart":    bucket.BucketStart,
			"bucketEnd":      bucketEnd,
			"timeHourUnit":   timeHourUnit,
			"field":          fieldName,
			"routineTaskIds": routineTaskIds,
		}
		meta, err := json.Marshal(metadata)
		if err != nil {
			return nil, s.routineException.FailedToMarshalData(metadata)
		}

		data[index] = capi.RoutineTaskRecordCountDatum{
			Id:    bucket.BucketStart.Format(time.RFC3339),
			X:     x,
			Value: bucket.RoutineTaskRecordCount,
			Meta:  meta,
		}
	}

	return data, nil
}

func (s *RoutineTaskRecordService) GetMyRoutineTaskRecordsByRoutineTaskId(
	ctx context.Context, requestDto *capi.GetMyRoutineTaskRecordsByRoutineTaskIdRequestDto,
) (*capi.GetMyRoutineTaskRecordsByRoutineTaskIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	routineTaskRecords, exception := s.routineTaskRecordRepository.GetAllByRoutineTaskId(
		requestDto.Param.RoutineTaskId,
		actorUserId,
		requestDto.Param.Limit,
		nil,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := make(capi.GetMyRoutineTaskRecordsByRoutineTaskIdResponseDto, len(routineTaskRecords))
	for index, routineTaskRecord := range routineTaskRecords {
		var errorCode *string
		if routineTaskRecord.ErrorCode != nil {
			errorCodeValue := routineTaskRecord.ErrorCode.String()
			errorCode = &errorCodeValue
		}
		responseDto[index] = capi.RoutineTaskRecordResponseDto{
			Id:              routineTaskRecord.Id,
			RoutineRecordId: routineTaskRecord.RoutineRecordId,
			RoutineTaskId:   routineTaskRecord.RoutineTaskId,
			Purpose:         routineTaskRecord.Purpose.String(),
			Status:          routineTaskRecord.Status.String(),
			ErrorCode:       errorCode,
			ErrorReason:     routineTaskRecord.ErrorReason,
			CostUnit:        routineTaskRecord.CostUnit,
			Attempts:        routineTaskRecord.Attempts,
			PayloadSnapshot: routineTaskRecord.PayloadSnapshot,
			ResultSnapshot:  routineTaskRecord.ResultSnapshot,
			ActualStartedAt: routineTaskRecord.ActualStartedAt,
			ActualEndedAt:   routineTaskRecord.ActualEndedAt,
			UpdatedAt:       routineTaskRecord.UpdatedAt,
			CreatedAt:       routineTaskRecord.CreatedAt,
		}
	}

	return &responseDto, nil
}

/* ============================== Visualization Methods ============================== */

func (s *RoutineTaskRecordService) VisualizeMyRoutineTaskRecordStatusCount(
	ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordStatusCountRequestDto,
) (*capi.VisualizeMyRoutineTaskRecordStatusCountResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	permission, err := cenums.ConvertStringToAccessControlPermission(requestDto.Param.Permission)
	if err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	var rows []struct {
		Status                 cenums.RoutineTaskRecordStatus `gorm:"column:status;"`
		RoutineTaskRecordCount int64                          `gorm:"column:routine_task_record_count;"`
	}

	query := db.Model(&sschemas.RoutineTaskRecord{}).
		Select(`"RoutineTaskRecordTable".status AS status, COUNT(*) AS routine_task_record_count`).
		Joins(`INNER JOIN "RoutineTaskTable" routine_task ON routine_task.id = "RoutineTaskRecordTable".routine_task_id`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = routine_task.routine_id AND routine.deleted_at IS NULL`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Where("uts.user_id = ? AND uts.permission = ?", actorUserId, *permission)
	if len(requestDto.Param.RoutineTaskIds) > 0 {
		query = query.Where(`"RoutineTaskRecordTable".routine_task_id IN ?`, requestDto.Param.RoutineTaskIds)
	}

	result := query.Group(`"RoutineTaskRecordTable".status`).Scan(&rows)
	if result.Error != nil {
		return nil, s.routineTaskException.NotFound().WithOrigin(result.Error)
	}

	counts := make(map[cenums.RoutineTaskRecordStatus]int64, len(rows))
	for _, row := range rows {
		counts[row.Status] = row.RoutineTaskRecordCount
	}

	data := make([]capi.RoutineTaskRecordCountDatum, len(cenums.AllRoutineTaskRecordStatuses))
	for index, status := range cenums.AllRoutineTaskRecordStatuses {
		metadata := map[string]any{"status": status.String(), "routineTaskIds": requestDto.Param.RoutineTaskIds}
		meta, err := json.Marshal(metadata)
		if err != nil {
			return nil, s.routineException.FailedToMarshalData(metadata)
		}

		data[index] = capi.RoutineTaskRecordCountDatum{
			Id:    status.String() + "-routine-task-record-count",
			X:     status.String() + " Routine Task Record Count",
			Value: counts[status],
			Meta:  meta,
		}
	}

	return &capi.VisualizeMyRoutineTaskRecordStatusCountResponseDto{
		Data: data,
	}, nil
}

func (s *RoutineTaskRecordService) VisualizeMyRoutineTaskRecordPurposeCount(
	ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordPurposeCountRequestDto,
) (*capi.VisualizeMyRoutineTaskRecordPurposeCountResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	permission, err := cenums.ConvertStringToAccessControlPermission(requestDto.Param.Permission)
	if err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	var rows []struct {
		Purpose                cenums.RoutineTaskPurpose `gorm:"column:purpose;"`
		RoutineTaskRecordCount int64                     `gorm:"column:routine_task_record_count;"`
	}

	query := db.Model(&sschemas.RoutineTaskRecord{}).
		Select(`"RoutineTaskRecordTable".purpose AS purpose, COUNT(*) AS routine_task_record_count`).
		Joins(`INNER JOIN "RoutineTaskTable" routine_task ON routine_task.id = "RoutineTaskRecordTable".routine_task_id`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = routine_task.routine_id AND routine.deleted_at IS NULL`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Where("uts.user_id = ? AND uts.permission = ?", actorUserId, *permission)
	if len(requestDto.Param.RoutineTaskIds) > 0 {
		query = query.Where(`"RoutineTaskRecordTable".routine_task_id IN ?`, requestDto.Param.RoutineTaskIds)
	}

	result := query.Group(`"RoutineTaskRecordTable".purpose`).Scan(&rows)
	if result.Error != nil {
		return nil, s.routineTaskException.NotFound().WithOrigin(result.Error)
	}

	counts := make(map[cenums.RoutineTaskPurpose]int64, len(rows))
	for _, row := range rows {
		counts[row.Purpose] = row.RoutineTaskRecordCount
	}

	data := make([]capi.RoutineTaskRecordCountDatum, len(cenums.AllRoutineTaskPurposes))
	for index, purpose := range cenums.AllRoutineTaskPurposes {
		metadata := map[string]any{"purpose": purpose.String(), "routineTaskIds": requestDto.Param.RoutineTaskIds}
		meta, err := json.Marshal(metadata)
		if err != nil {
			return nil, s.routineException.FailedToMarshalData(metadata)
		}

		data[index] = capi.RoutineTaskRecordCountDatum{
			Id:    purpose.String() + "-routine-task-record-count",
			X:     purpose.String() + " Routine Task Record Count",
			Value: counts[purpose],
			Meta:  meta,
		}
	}

	return &capi.VisualizeMyRoutineTaskRecordPurposeCountResponseDto{
		Data: data,
	}, nil
}

func (s *RoutineTaskRecordService) VisualizeMyRoutineTaskRecordScheduledAtCount(
	ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordScheduledAtCountRequestDto,
) (*capi.VisualizeMyRoutineTaskRecordScheduledAtCountResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	if !requestDto.Param.QueryRangeStartedAt.Before(requestDto.Param.QueryRangeEndedAt) {
		return nil, s.routineTaskException.InvalidDto("queryRangeStartedAt should be earlier then queryRangeEndedAt")
	}
	if !stimes.IsTimeWithin(requestDto.Param.QueryRangeStartedAt, requestDto.Param.QueryRangeEndedAt, 360*24*time.Hour) {
		return nil, s.routineTaskException.InvalidDto("queryRangeStartedAt and queryRangeEndedAt should be within 360 days")
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	permission, err := cenums.ConvertStringToAccessControlPermission(requestDto.Param.Permission)
	if err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	data, exception := s.visualizeMyRoutineTaskRecordTimeCount(
		ctx,
		actorUserId,
		*permission,
		requestDto.Param.RoutineTaskIds,
		requestDto.Param.TimeHourUnit,
		requestDto.Param.QueryRangeStartedAt,
		requestDto.Param.QueryRangeEndedAt,
		"scheduled_at",
		"scheduledAt",
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.VisualizeMyRoutineTaskRecordScheduledAtCountResponseDto{
		Data: data,
	}, nil
}

func (s *RoutineTaskRecordService) VisualizeMyRoutineTaskRecordActualStartedAtCount(
	ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordActualStartedAtCountRequestDto,
) (*capi.VisualizeMyRoutineTaskRecordActualStartedAtCountResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	if !requestDto.Param.QueryRangeStartedAt.Before(requestDto.Param.QueryRangeEndedAt) {
		return nil, s.routineTaskException.InvalidDto("queryRangeStartedAt should be earlier then queryRangeEndedAt")
	}
	if !stimes.IsTimeWithin(requestDto.Param.QueryRangeStartedAt, requestDto.Param.QueryRangeEndedAt, 360*24*time.Hour) {
		return nil, s.routineTaskException.InvalidDto("queryRangeStartedAt and queryRangeEndedAt should be within 360 days")
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	permission, err := cenums.ConvertStringToAccessControlPermission(requestDto.Param.Permission)
	if err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	data, exception := s.visualizeMyRoutineTaskRecordTimeCount(
		ctx,
		actorUserId,
		*permission,
		requestDto.Param.RoutineTaskIds,
		requestDto.Param.TimeHourUnit,
		requestDto.Param.QueryRangeStartedAt,
		requestDto.Param.QueryRangeEndedAt,
		"actual_started_at",
		"actualStartedAt",
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.VisualizeMyRoutineTaskRecordActualStartedAtCountResponseDto{
		Data: data,
	}, nil
}

func (s *RoutineTaskRecordService) VisualizeMyRoutineTaskRecordActualEndedAtCount(
	ctx context.Context, requestDto *capi.VisualizeMyRoutineTaskRecordActualEndedAtCountRequestDto,
) (*capi.VisualizeMyRoutineTaskRecordActualEndedAtCountResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	if !requestDto.Param.QueryRangeStartedAt.Before(requestDto.Param.QueryRangeEndedAt) {
		return nil, s.routineTaskException.InvalidDto("queryRangeStartedAt should be earlier then queryRangeEndedAt")
	}
	if !stimes.IsTimeWithin(requestDto.Param.QueryRangeStartedAt, requestDto.Param.QueryRangeEndedAt, 360*24*time.Hour) {
		return nil, s.routineTaskException.InvalidDto("queryRangeStartedAt and queryRangeEndedAt should be within 360 days")
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	permission, err := cenums.ConvertStringToAccessControlPermission(requestDto.Param.Permission)
	if err != nil {
		return nil, s.routineTaskException.InvalidDto().WithOrigin(err)
	}
	data, exception := s.visualizeMyRoutineTaskRecordTimeCount(
		ctx,
		actorUserId,
		*permission,
		requestDto.Param.RoutineTaskIds,
		requestDto.Param.TimeHourUnit,
		requestDto.Param.QueryRangeStartedAt,
		requestDto.Param.QueryRangeEndedAt,
		"actual_ended_at",
		"actualEndedAt",
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.VisualizeMyRoutineTaskRecordActualEndedAtCountResponseDto{
		Data: data,
	}, nil
}

/* ============================== GraphQL Methods ============================== */

func (s *RoutineTaskRecordService) SearchPrivateRoutineTaskRecords(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineTaskRecordInput,
) (*cgqlmodels.SearchRoutineTaskRecordConnection, *cexceptions.Exception) {
	type PrivateRoutineTaskRecord struct {
		sschemas.RoutineTaskRecord
		Permission cenums.AccessControlPermission `gorm:"column:permission"`
	}

	startTime := time.Now()
	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	query := db.Model(&sschemas.RoutineTaskRecord{}).
		Select(`
			"RoutineTaskRecordTable".id,
			"RoutineTaskRecordTable".routine_record_id,
			"RoutineTaskRecordTable".routine_task_id,
			"RoutineTaskRecordTable".purpose,
			"RoutineTaskRecordTable".status,
			"RoutineTaskRecordTable".error_code,
			"RoutineTaskRecordTable".error_reason,
			"RoutineTaskRecordTable".cost_unit,
			"RoutineTaskRecordTable".attempts,
			"RoutineTaskRecordTable".payload_snapshot,
			"RoutineTaskRecordTable".result_snapshot,
			"RoutineTaskRecordTable".actual_started_at,
			"RoutineTaskRecordTable".actual_ended_at,
			"RoutineTaskRecordTable".updated_at,
			"RoutineTaskRecordTable".created_at,
			uts.permission AS permission`).
		Joins(`INNER JOIN "RoutineRecordTable" routine_record ON routine_record.id = "RoutineTaskRecordTable".routine_record_id`).
		Joins(`INNER JOIN "RoutineTaskTable" routine_task ON routine_task.id = "RoutineTaskRecordTable".routine_task_id`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = routine_task.routine_id AND routine.deleted_at IS NULL`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions)

	if len(gqlInput.RoutineTaskIds) > 0 {
		query = query.Where(`"RoutineTaskRecordTable".routine_task_id IN ?`, gqlInput.RoutineTaskIds)
	}

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			`"RoutineTaskRecordTable".purpose::text ILIKE ?
				OR "RoutineTaskRecordTable".status::text ILIKE ?
				OR "RoutineTaskRecordTable".error_code::text ILIKE ?
				OR "RoutineTaskRecordTable".error_reason ILIKE ?`,
			"%"+gqlInput.Query+"%",
			"%"+gqlInput.Query+"%",
			"%"+gqlInput.Query+"%",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchRoutineTaskRecordCursorFields](*gqlInput.After)
		if err != nil {
			return nil, s.searchException.FailedToDecode().WithOrigin(err)
		}

		query = query.Where(`"RoutineTaskRecordTable".id > ?`, searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		cending := cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchRoutineTaskRecordSortByPurpose:
			query = query.Order(`"RoutineTaskRecordTable".purpose ` + cending).
				Order(`"RoutineTaskRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineTaskRecordSortByStatus:
			query = query.Order(`"RoutineTaskRecordTable".status ` + cending).
				Order(`"RoutineTaskRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineTaskRecordSortByCostUnit:
			query = query.Order(`"RoutineTaskRecordTable".cost_unit ` + cending).
				Order(`"RoutineTaskRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineTaskRecordSortByAttempts:
			query = query.Order(`"RoutineTaskRecordTable".attempts ` + cending).
				Order(`"RoutineTaskRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineTaskRecordSortByScheduledAt:
			query = query.Order(`routine_record.scheduled_at ` + cending).
				Order(`"RoutineTaskRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineTaskRecordSortByActualStartedAt:
			query = query.Order(`"RoutineTaskRecordTable".actual_started_at ` + cending).
				Order(`"RoutineTaskRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineTaskRecordSortByActualEndedAt:
			query = query.Order(`"RoutineTaskRecordTable".actual_ended_at ` + cending).
				Order(`"RoutineTaskRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineTaskRecordSortByLastUpdate:
			query = query.Order(`"RoutineTaskRecordTable".updated_at ` + cending).
				Order(`"RoutineTaskRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineTaskRecordSortByCreatedAt:
			query = query.Order(`"RoutineTaskRecordTable".created_at ` + cending)
		default:
			query = query.Order(`"RoutineTaskRecordTable".created_at ` + cending)
		}
	}

	limit := sconstants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, sconstants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var routineTaskRecords []PrivateRoutineTaskRecord
	if err := query.Find(&routineTaskRecords).Error; err != nil {
		return nil, s.routineTaskException.NotFound().WithOrigin(err)
	}

	hasNextPage := len(routineTaskRecords) > limit
	searchEdges := make([]*cgqlmodels.SearchRoutineTaskRecordEdge, len(routineTaskRecords))

	for index, routineTaskRecord := range routineTaskRecords {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchRoutineTaskRecordCursorFields]{
			Fields: cgqlmodels.SearchRoutineTaskRecordCursorFields{
				ID: routineTaskRecord.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, s.searchException.FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, s.searchException.FailedToUnmarshalSearchCursor()
		}

		searchEdges[index] = &cgqlmodels.SearchRoutineTaskRecordEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                routineTaskRecord.RoutineTaskRecord.ToPrivateRoutineTaskRecord(),
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

	return &cgqlmodels.SearchRoutineTaskRecordConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
