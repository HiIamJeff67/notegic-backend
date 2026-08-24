package other

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	searchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"

	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
)

type ThemeServiceInterface interface {
	GetPublicThemeByPublicId(ctx context.Context, publicId uuid.UUID) (*gqlmodels.PublicTheme, *exceptions.Exception)

	SearchPublicThemes(ctx context.Context, gqlInput gqlmodels.SearchThemeInput) (*gqlmodels.SearchThemeConnection, *exceptions.Exception)
}

type ThemeService struct {
	db *gorm.DB
}

func NewThemeService(db *gorm.DB) ThemeServiceInterface {
	return &ThemeService{
		db: db,
	}
}

/* ============================== Service Methods for Theme ============================== */

// get the theme which are created by the current user
func (s *ThemeService) GetMyThemeById() {}

/* ============================== Service Methods for Public Theme ============================== */

func (s *ThemeService) GetPublicThemeByPublicId(
	ctx context.Context, publicId uuid.UUID,
) (*gqlmodels.PublicTheme, *exceptions.Exception) {
	db := s.db.WithContext(ctx)

	theme := schemas.Theme{}
	result := db.
		Model(&schemas.Theme{}).
		Where("public_id = ?", publicId).
		First(&theme)
	if err := result.Error; err != nil {
		return nil, exceptions.New(
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
	gqlInput gqlmodels.SearchThemeInput,
) (*gqlmodels.SearchThemeConnection, *exceptions.Exception) {
	startTime := time.Now()
	db := s.db.WithContext(ctx)

	query := db.Model(&schemas.Theme{})

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"name ILIKE ?",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := searchcursor.Decode[gqlmodels.SearchThemeCursorFields](*gqlInput.After)
		if err != nil {
			return nil, exceptions.New(
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
		var cending string = gqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == gqlmodels.SearchSortOrderDesc {
			cending = gqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case gqlmodels.SearchThemeSortByName:
			query = query.Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case gqlmodels.SearchThemeSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("name " + cending).
				Order("created_at " + cending)
		case gqlmodels.SearchThemeSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("name " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("name " + cending).
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

	var themes []schemas.Theme
	if err := query.Find(&themes).Error; err != nil {
		return nil, exceptions.New(
			"QueryFailed",
			"Theme",
			"SearchPublicThemes",
			"Failed to retrieve themes",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	hasNextPage := len(themes) > limit
	searchEdges := make([]*gqlmodels.SearchThemeEdge, len(themes))

	for index, theme := range themes {
		searchCursor := searchcursor.SearchCursor[gqlmodels.SearchThemeCursorFields]{
			Fields: gqlmodels.SearchThemeCursorFields{
				PublicID: theme.PublicId,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, exceptions.New(
				"CursorEncodeFailed",
				"Search",
				"SearchPublicThemes",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, exceptions.New(
				"CursorEncodingFailed",
				"Search",
				"SearchPublicThemes",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			)
		}

		searchEdges[index] = &gqlmodels.SearchThemeEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                theme.ToPublicTheme(),
		}
	}

	searchPageInfo := &gqlmodels.SearchPageInfo{
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

	return &gqlmodels.SearchThemeConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
