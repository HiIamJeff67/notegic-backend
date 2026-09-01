package routines

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/core/exceptions"
)

type RoutineRecordServiceInterface interface {
	SearchPrivateRoutineRecords(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineRecordInput) (*cgqlmodels.SearchRoutineRecordConnection, *cexceptions.Exception)
}

type RoutineRecordService struct {
	db *gorm.DB
}

func NewRoutineRecordService(
	db *gorm.DB,
) RoutineRecordServiceInterface {
	return &RoutineRecordService{
		db: db,
	}
}

func (s *RoutineRecordService) SearchPrivateRoutineRecords(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRoutineRecordInput,
) (*cgqlmodels.SearchRoutineRecordConnection, *cexceptions.Exception) {
	startTime := time.Now()
	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	query := db.Model(&sschemas.RoutineRecord{}).
		Select(`"RoutineRecordTable".*`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = "RoutineRecordTable".routine_id AND routine.deleted_at IS NULL`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions)

	if len(gqlInput.RoutineIds) > 0 {
		query = query.Where(`"RoutineRecordTable".routine_id IN ?`, gqlInput.RoutineIds)
	}

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			`"RoutineRecordTable".status::text ILIKE ?
				OR "RoutineRecordTable".snapshot::text ILIKE ?`,
			"%"+gqlInput.Query+"%",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchRoutineRecordCursorFields](*gqlInput.After)
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToDecode().WithOrigin(err)
		}

		query = query.Where(`"RoutineRecordTable".id > ?`, searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		cending := cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchRoutineRecordSortByStatus:
			query = query.Order(`"RoutineRecordTable".status ` + cending).
				Order(`"RoutineRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineRecordSortByScheduledAt:
			query = query.Order(`"RoutineRecordTable".scheduled_at ` + cending).
				Order(`"RoutineRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineRecordSortByActualStartedAt:
			query = query.Order(`"RoutineRecordTable".actual_started_at ` + cending).
				Order(`"RoutineRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineRecordSortByActualEndedAt:
			query = query.Order(`"RoutineRecordTable".actual_ended_at ` + cending).
				Order(`"RoutineRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineRecordSortByLastUpdate:
			query = query.Order(`"RoutineRecordTable".updated_at ` + cending).
				Order(`"RoutineRecordTable".created_at ` + cending)
		case cgqlmodels.SearchRoutineRecordSortByCreatedAt:
			query = query.Order(`"RoutineRecordTable".created_at ` + cending)
		default:
			query = query.Order(`"RoutineRecordTable".created_at ` + cending)
		}
	}

	limit := sconstants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, sconstants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var routineRecords []sschemas.RoutineRecord
	if err := query.Find(&routineRecords).Error; err != nil {
		return nil, apiexceptions.NewRoutineException().NotFound().WithOrigin(err)
	}

	hasNextPage := len(routineRecords) > limit
	searchEdges := make([]*cgqlmodels.SearchRoutineRecordEdge, len(routineRecords))

	for index, routineRecord := range routineRecords {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchRoutineRecordCursorFields]{
			Fields: cgqlmodels.SearchRoutineRecordCursorFields{ID: routineRecord.Id},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, apiexceptions.NewSearchException().FailedToUnmarshalSearchCursor()
		}

		searchEdges[index] = &cgqlmodels.SearchRoutineRecordEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                routineRecord.ToPrivateRoutineRecord(),
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

	return &cgqlmodels.SearchRoutineRecordConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
