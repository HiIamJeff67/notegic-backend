package shelves

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	searchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"

	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
)

type ItemServiceInterface interface {
	SearchPrivateItems(ctx context.Context, userId uuid.UUID, gqlInput gqlmodels.SearchItemInput) (*gqlmodels.SearchItemConnection, *exceptions.Exception)
}

type ItemService struct {
	db        *gorm.DB
	itemScope scopes.ItemScopeInterface
}

func NewItemService(
	db *gorm.DB,
	itemScope scopes.ItemScopeInterface,
) ItemServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &ItemService{
		db:        db,
		itemScope: itemScope,
	}
}

/* ============================== Service Methods for GraphQL Item ============================== */

func (s *ItemService) SearchPrivateItems(
	ctx context.Context, userId uuid.UUID, gqlInput gqlmodels.SearchItemInput,
) (*gqlmodels.SearchItemConnection, *exceptions.Exception) {
	type PrivateItem struct {
		schemas.Item
		Permission enums.AccessControlPermission `gorm:"column:permission"`
	}

	startTime := time.Now()
	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	onlyDeleted := types.Ternary_Negative
	if gqlInput.IsDeletedAt != nil && *gqlInput.IsDeletedAt {
		onlyDeleted = types.Ternary_Positive
	}

	query := db.Model(&schemas.Item{}).
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
		searchCursor, err := searchcursor.Decode[gqlmodels.SearchItemCursorFields](*gqlInput.After)
		if err != nil {
			return nil, exceptions.New(
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
		var cending string = gqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == gqlmodels.SearchSortOrderDesc {
			cending = gqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case gqlmodels.SearchItemSortByType:
			query = query.Order("type " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case gqlmodels.SearchItemSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("type " + cending).
				Order("created_at " + cending)
		case gqlmodels.SearchItemSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("type " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("type " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		}
	}

	limit := constants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, constants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var items []PrivateItem
	if err := query.Preload(
		string(schemas.ItemRelation_RoutinesToItems),
		func(preloadDB *gorm.DB) *gorm.DB {
			return preloadDB.
				Joins(`INNER JOIN "RoutineTable" ON "RoutineTable".id = "RoutinesToItemsTable".routine_id`).
				Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = "RoutineTable".station_id`).
				Where(`"RoutineTable".deleted_at IS NULL`).
				Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions)
		},
	).Find(&items).Error; err != nil {
		return nil, exceptions.New(
			"QueryFailed",
			"Item",
			"SearchPrivateItems",
			"Failed to retrieve items",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	hasNextPage := len(items) > limit
	searchEdges := make([]*gqlmodels.SearchItemEdge, len(items))

	for index, item := range items {
		searchCursor := searchcursor.SearchCursor[gqlmodels.SearchItemCursorFields]{
			Fields: gqlmodels.SearchItemCursorFields{
				ID: item.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, exceptions.New(
				"CursorEncodeFailed",
				"Search",
				"SearchPrivateItems",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, exceptions.New(
				"CursorEncodingFailed",
				"Search",
				"SearchPrivateItems",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			)
		}

		searchEdges[index] = &gqlmodels.SearchItemEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                item.Item.ToPrivateItem(),
		}
	}

	searchPageInfo := &gqlmodels.SearchPageInfo{
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

	return &gqlmodels.SearchItemConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
