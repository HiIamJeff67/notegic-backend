package user

import (
	"context"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	searchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/users"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	userdata "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata"
	cacheinputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata/inputs"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type UserServiceInterface interface {
	GetUserData(ctx context.Context, requestDto *capi.GetUserDataRequestDto) (*capi.GetUserDataResponseDto, *cexceptions.Exception)
	GetMe(ctx context.Context, requestDto *capi.GetMeRequestDto) (*capi.GetMeResponseDto, *cexceptions.Exception)
	UpdateMe(ctx context.Context, requestDto *capi.UpdateMeRequestDto) (*capi.UpdateMeResponseDto, *cexceptions.Exception)

	// services for graphql users
	GetPublicUserByPublicId(ctx context.Context, publicId uuid.UUID) (*cgqlmodels.PublicUser, *cexceptions.Exception)
	GetPublicAuthorByThemePublicIds(ctx context.Context, publicIds []uuid.UUID) ([]*cgqlmodels.PublicUser, *cexceptions.Exception)
	SearchPublicUsers(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchUserInput) (*cgqlmodels.SearchUserConnection, *cexceptions.Exception)
}

type UserService struct {
	validator           *validator.Validate
	db                  *gorm.DB
	userRepository      repositories.UserRepositoryInterface
	userDataCacheClient *userdata.UserDataCacheClient
}

func NewUserService(
	validator *validator.Validate,
	db *gorm.DB,
	userRepository repositories.UserRepositoryInterface,
	userDataCacheClient *userdata.UserDataCacheClient,
) UserServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &UserService{
		validator:           validator,
		db:                  db,
		userRepository:      userRepository,
		userDataCacheClient: userDataCacheClient,
	}
}

/* ============================== Service Methods for Users ============================== */

func (s *UserService) GetUserData(
	ctx context.Context, requestDto *capi.GetUserDataRequestDto,
) (*capi.GetUserDataResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"User",
			"GetUserData",
			"User data request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	var user schemas.User
	if result := s.db.WithContext(ctx).Select("name").Where("id = ?", actorUserId).First(&user); result.Error != nil {
		return nil, cexceptions.New("NotFound", "User", "ResolveUser", "User was not found", http.StatusNotFound).WithOrigin(result.Error)
	}

	userDataCache, exception := s.userDataCacheClient.Get(user.Name)
	if exception != nil {
		return nil, exception
	}

	return &capi.GetUserDataResponseDto{
		PublicId:    userDataCache.PublicId,
		Name:        userDataCache.Name,
		DisplayName: userDataCache.DisplayName,
		Email:       userDataCache.Email,
		Role:        userDataCache.Role.String(),
		Plan:        userDataCache.Plan.String(),
		Status:      userDataCache.Status.String(),
		AvatarURL:   userDataCache.AvatarURL,
		CreatedAt:   userDataCache.CreatedAt,
		UpdatedAt:   userDataCache.UpdatedAt,
	}, nil
}

func (s *UserService) GetMe(
	ctx context.Context, requestDto *capi.GetMeRequestDto,
) (*capi.GetMeResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"User",
			"GetMe",
			"User request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)

	user, exception := s.userRepository.GetOneById(actorUserId, nil, options.WithDB(db))
	if exception != nil {
		return nil, exception
	}

	return &capi.GetMeResponseDto{
		PublicId:    user.PublicId,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Role:        user.Role.String(),
		Plan:        user.Plan.String(),
		Status:      user.Status.String(),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

func (s *UserService) UpdateMe(
	ctx context.Context, requestDto *capi.UpdateMeRequestDto,
) (*capi.UpdateMeResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"User",
			"UpdateMe",
			"User update request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	var status *enums.UserStatus
	if requestDto.Body.Values.Status != nil {
		parsedStatus, err := enums.ConvertStringToUserStatus(*requestDto.Body.Values.Status)
		if err != nil {
			return nil, cexceptions.InvalidInput("User").WithOrigin(err)
		}
		status = parsedStatus
	}
	var user schemas.User
	if result := s.db.WithContext(ctx).Select("name").Where("id = ?", actorUserId).First(&user); result.Error != nil {
		return nil, cexceptions.New("NotFound", "User", "ResolveUser", "User was not found", http.StatusNotFound).WithOrigin(result.Error)
	}

	db := s.db.WithContext(ctx)

	updatedUser, exception := s.userRepository.UpdateOneById(
		actorUserId,
		inputs.PartialUpdateUserInput{
			Values: inputs.UpdateUserInput{
				DisplayName: requestDto.Body.Values.DisplayName,
				Status:      status,
			},
			SetNull: requestDto.Body.SetNull,
		},
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	if requestDto.Body.Values.DisplayName != nil {
		exception = s.userDataCacheClient.Update(user.Name, cacheinputs.UpdateUserDataCacheInput{
			DisplayName: requestDto.Body.Values.DisplayName,
		})
		if exception != nil && logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(
				ctx,
				exception.Origin(),
				exception.String(),
			)
		}
	}
	if requestDto.Body.Values.Status != nil {
		exception = s.userDataCacheClient.Update(user.Name, cacheinputs.UpdateUserDataCacheInput{
			Status: status,
		})
		if exception != nil && logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(
				ctx,
				exception.Origin(),
				exception.String(),
			)
		}
	}

	return &capi.UpdateMeResponseDto{UpdatedAt: updatedUser.UpdatedAt}, nil
}

/* ============================== Service Methods for Public User (Only available in GraphQL) ============================== */

func (s *UserService) GetPublicUserByPublicId(
	ctx context.Context, publicId uuid.UUID,
) (*cgqlmodels.PublicUser, *cexceptions.Exception) {
	db := s.db.WithContext(ctx)

	user := schemas.User{}
	result := db.Table(schemas.User{}.TableName()).
		Where("public_id = ?", publicId).
		First(&user)
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"GetPublicUserByPublicId",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	return user.ToPublicUser(), nil
}

func (s *UserService) GetPublicAuthorByThemePublicIds(
	ctx context.Context, publicIds []uuid.UUID,
) ([]*cgqlmodels.PublicUser, *cexceptions.Exception) {
	if len(publicIds) == 0 {
		return []*cgqlmodels.PublicUser{}, nil
	}

	db := s.db.WithContext(ctx)

	uniquePublicIds := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]bool)
	for _, publicId := range publicIds {
		if !seen[publicId] {
			uniquePublicIds = append(uniquePublicIds, publicId)
			seen[publicId] = true
		}
	}
	if len(uniquePublicIds) == 0 {
		return make([]*cgqlmodels.PublicUser, len(publicIds)), nil
	}

	var authorsWithPublicThemeIds []*struct {
		schemas.User
		ThemePublicId uuid.UUID `gorm:"theme_public_id"`
	}
	result := db.Table(schemas.User{}.TableName()+" u").
		Select("u.*, t.public_id as theme_public_id").
		Joins(`LEFT JOIN "ThemeTable" t ON t.author_id = u.id`).
		Where("t.public_id IN ?", uniquePublicIds).
		Find(&authorsWithPublicThemeIds)
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
			"QueryFailed",
			"User",
			"GetPublicAuthorByThemePublicIds",
			"Failed to retrieve users",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	publicIdToIndexesMap := make(map[uuid.UUID][]int)
	for index, publicId := range publicIds {
		publicIdToIndexesMap[publicId] = append(publicIdToIndexesMap[publicId], index)
	}

	publicUsers := make([]*cgqlmodels.PublicUser, len(publicIds))
	for _, authorWithPublicThemeId := range authorsWithPublicThemeIds {
		for _, index := range publicIdToIndexesMap[authorWithPublicThemeId.ThemePublicId] {
			publicUsers[index] = authorWithPublicThemeId.ToPublicUser()
		}
	}

	return publicUsers, nil
}

func (s *UserService) SearchPublicUsers(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchUserInput,
) (*cgqlmodels.SearchUserConnection, *cexceptions.Exception) {
	startTime := time.Now()

	db := s.db.WithContext(ctx)

	query := db.Model(&schemas.User{})

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"name ILIKE ? OR display_name ILIKE ? OR email ILIKE ?",
			"%"+gqlInput.Query+"%", "%"+gqlInput.Query+"%", "%"+gqlInput.Query+"%",
		)
	}
	if len(gqlInput.RootShelfIds) > 0 {
		accessibleRootShelfIds := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select(`"UsersToShelvesTable".root_shelf_id`).
			Joins(`INNER JOIN "RootShelfTable" ON "RootShelfTable".id = "UsersToShelvesTable".root_shelf_id`).
			Where(`"UsersToShelvesTable".user_id = ?`, userId).
			Where(`"UsersToShelvesTable".root_shelf_id IN ?`, gqlInput.RootShelfIds).
			Where(`"RootShelfTable".deleted_at IS NULL`)

		usersWithAccessibleRootShelf := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Where(`"UsersToShelvesTable".user_id = "UserTable".id`).
			Where(`"UsersToShelvesTable".root_shelf_id IN (?)`, accessibleRootShelfIds)

		query = query.Where("EXISTS (?)", usersWithAccessibleRootShelf)
	}
	if len(gqlInput.StationIds) > 0 {
		accessibleStationIds := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToStations{}).
			Select(`"UsersToStationsTable".station_id`).
			Joins(`INNER JOIN "StationTable" ON "StationTable".id = "UsersToStationsTable".station_id`).
			Where(`"UsersToStationsTable".user_id = ?`, userId).
			Where(`"UsersToStationsTable".station_id IN ?`, gqlInput.StationIds).
			Where(`"StationTable".deleted_at IS NULL`)

		usersWithAccessibleStation := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToStations{}).
			Select("1").
			Where(`"UsersToStationsTable".user_id = "UserTable".id`).
			Where(`"UsersToStationsTable".station_id IN (?)`, accessibleStationIds)

		query = query.Where("EXISTS (?)", usersWithAccessibleStation)
	}
	if len(gqlInput.BadgePublicIds) > 0 {
		usersWithBadge := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToBadges{}).
			Select("1").
			Joins(`INNER JOIN "BadgeTable" ON "BadgeTable".id = "UsersToBadgesTable".badge_id`).
			Where(`"UsersToBadgesTable".user_id = "UserTable".id`).
			Where(`"BadgeTable".public_id IN ?`, gqlInput.BadgePublicIds)

		query = query.Where("EXISTS (?)", usersWithBadge)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := searchcursor.Decode[cgqlmodels.SearchUserCursorFields](*gqlInput.After)
		if err != nil {
			return nil, cexceptions.New(
				"CursorDecodeFailed",
				"Search",
				"SearchPublicUsers",
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
		case cgqlmodels.SearchUserSortByName:
			query.Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchUserSortByLastActive:
			query.Order("updated_at " + cending).
				Order("name " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchUserSortByCreatedAt:
			query.Order("created_at " + cending).
				Order("name " + cending).
				Order("updated_at " + cending)
		default:
			query.Order("updated_at " + cending).
				Order("name " + cending).
				Order("created_at " + cending)
		}
	}

	limit := constants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, constants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var users []schemas.User
	if err := query.Find(&users).Error; err != nil {
		return nil, cexceptions.New(
			"QueryFailed",
			"User",
			"SearchPublicUsers",
			"Failed to retrieve users",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	hasNextPage := len(users) > limit // since we fetch an additional one
	searchEdges := make([]*cgqlmodels.SearchUserEdge, len(users))

	for index, user := range users {
		searchCursor := searchcursor.SearchCursor[cgqlmodels.SearchUserCursorFields]{
			Fields: cgqlmodels.SearchUserCursorFields{
				PublicID: user.PublicId,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, cexceptions.New(
				"CursorEncodeFailed",
				"Search",
				"SearchPublicUsers",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, cexceptions.New(
				"CursorEncodingFailed",
				"Search",
				"SearchPublicUsers",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			)
		}

		searchEdges[index] = &cgqlmodels.SearchUserEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                user.ToPublicUser(),
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

	return &cgqlmodels.SearchUserConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
