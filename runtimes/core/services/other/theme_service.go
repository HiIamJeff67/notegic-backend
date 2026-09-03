package other

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type ThemeServiceInterface interface {
	/* ============================== GraphQL Methods ============================== */
	GetPublicThemeByPublicId(ctx context.Context, publicId uuid.UUID) (*cgqlmodels.PublicTheme, *cexceptions.Exception)

	SearchPublicThemes(ctx context.Context, gqlInput cgqlmodels.SearchThemeInput) (*cgqlmodels.SearchThemeConnection, *cexceptions.Exception)
}

type ThemeService struct {
	db *gorm.DB
}

func NewThemeService(db *gorm.DB) ThemeServiceInterface {
	return &ThemeService{
		db: db,
	}
}

// get the theme which are created by the current user
func (s *ThemeService) GetMyThemeById() {}

/* ============================== GraphQL Methods ============================== */

func (s *ThemeService) GetPublicThemeByPublicId(
	ctx context.Context, publicId uuid.UUID,
) (*cgqlmodels.PublicTheme, *cexceptions.Exception) {
	db := s.db.WithContext(ctx)

	theme := sschemas.Theme{}
	result := db.
		Model(&sschemas.Theme{}).
		Where("public_id = ?", publicId).
		First(&theme)
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
			"NotFound",
			"Theme",
			"GetPublicThemeByPublicId",
			"Theme was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	return theme.ToPublicTheme(), nil
}

func (s *ThemeService) SearchPublicThemes(
	ctx context.Context,
	gqlInput cgqlmodels.SearchThemeInput,
) (*cgqlmodels.SearchThemeConnection, *cexceptions.Exception) {
	startTime := time.Now()
	db := s.db.WithContext(ctx)

	query := db.Model(&sschemas.Theme{})

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"name ILIKE ?",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchThemeCursorFields](*gqlInput.After)
		if err != nil {
			return nil, cexceptions.New(
				"CursorDecodeFailed",
				"Search",
				"SearchPublicThemes",
				"Failed to decode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}

		query = query.Where("public_id > ?", searchCursor.Fields.PublicID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		var cending string = cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchThemeSortByName:
			query = query.Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchThemeSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("name " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchThemeSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("name " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("name " + cending).
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

	var themes []sschemas.Theme
	if err := query.Find(&themes).Error; err != nil {
		return nil, cexceptions.New(
			"QueryFailed",
			"Theme",
			"SearchPublicThemes",
			"Failed to retrieve themes",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	hasNextPage := len(themes) > limit
	searchEdges := make([]*cgqlmodels.SearchThemeEdge, len(themes))

	for index, theme := range themes {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchThemeCursorFields]{
			Fields: cgqlmodels.SearchThemeCursorFields{
				PublicID: theme.PublicId,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, cexceptions.New(
				"CursorEncodeFailed",
				"Search",
				"SearchPublicThemes",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, cexceptions.New(
				"CursorEncodingFailed",
				"Search",
				"SearchPublicThemes",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			)
		}

		searchEdges[index] = &cgqlmodels.SearchThemeEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                theme.ToPublicTheme(),
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

	return &cgqlmodels.SearchThemeConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
