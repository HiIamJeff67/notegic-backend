package shelves

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
)

type ItemServiceInterface interface {
	SearchPrivateItems(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchItemInput) (*cgqlmodels.SearchItemConnection, *cexceptions.Exception)
}

type ItemService struct {
	db        *gorm.DB
	itemScope sscopes.ItemScopeInterface
}

func NewItemService(
	db *gorm.DB,
	itemScope sscopes.ItemScopeInterface,
) ItemServiceInterface {
	return &ItemService{
		db:        db,
		itemScope: itemScope,
	}
}

/* ============================== Service Methods for GraphQL Item ============================== */

func (s *ItemService) SearchPrivateItems(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchItemInput,
) (*cgqlmodels.SearchItemConnection, *cexceptions.Exception) {
	type PrivateItem struct {
		sschemas.Item
		Permission cenums.AccessControlPermission `gorm:"column:permission"`
	}

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

	query := db.Model(&sschemas.Item{}).
		Select(`"ItemTable".*, uts.permission AS permission`).
		Joins(`INNER JOIN "UsersToShelvesTable" uts ON "ItemTable".root_shelf_id = uts.root_shelf_id`).
		Joins(`LEFT JOIN "MaterialTable" m ON "ItemTable".type = 'Material'::"ItemType" AND m.id = "ItemTable".id`).
		Joins(`LEFT JOIN "BlockPackTable" bp ON "ItemTable".type = 'BlockPack'::"ItemType" AND bp.id = "ItemTable".id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions).
		Scopes(s.itemScope.FilterOnlyDeleted(onlyDeleted))

	if gqlInput.ParentSubShelfID != nil {
		query = query.Where(
			`"ItemTable".parent_sub_shelf_id = ?`,
			*gqlInput.ParentSubShelfID,
		)
	}

	if gqlInput.RootShelfID != nil {
		query = query.Where(
			`"ItemTable".root_shelf_id = ?`,
			*gqlInput.RootShelfID,
		)
	}

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"COALESCE(m.name, bp.name) ILIKE ?",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchItemCursorFields](*gqlInput.After)
		if err != nil {
			return nil, cexceptions.New(
				"CursorDecodeFailed",
				"Search",
				"SearchPrivateItems",
				"Failed to decode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}

		query = query.Where("id > ?", searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		var cending string = cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchItemSortByType:
			query = query.Order("type " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchItemSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("type " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchItemSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("type " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("type " + cending).
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

	var items []PrivateItem
	if err := query.Preload(
		string(sschemas.ItemRelation_RoutinesToItems),
		func(preloadDB *gorm.DB) *gorm.DB {
			return preloadDB.
				Joins(`INNER JOIN "RoutineTable" ON "RoutineTable".id = "RoutinesToItemsTable".routine_id`).
				Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = "RoutineTable".station_id`).
				Where(`"RoutineTable".deleted_at IS NULL`).
				Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions)
		},
	).Find(&items).Error; err != nil {
		return nil, cexceptions.New(
			"QueryFailed",
			"Item",
			"SearchPrivateItems",
			"Failed to retrieve items",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	hasNextPage := len(items) > limit
	searchEdges := make([]*cgqlmodels.SearchItemEdge, len(items))

	for index, item := range items {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchItemCursorFields]{
			Fields: cgqlmodels.SearchItemCursorFields{
				ID: item.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, cexceptions.New(
				"CursorEncodeFailed",
				"Search",
				"SearchPrivateItems",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, cexceptions.New(
				"CursorEncodingFailed",
				"Search",
				"SearchPrivateItems",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			)
		}

		searchEdges[index] = &cgqlmodels.SearchItemEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                item.Item.ToPrivateItem(),
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

	return &cgqlmodels.SearchItemConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
