package user

import (
	"context"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-infos"
	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	userdata "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata"
	cacheinputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata/inputs"
)

type UserInfoServiceInterface interface {
	GetMyInfo(ctx context.Context, requestDto *apicontract.GetMyInfoRequestDto) (*apicontract.GetMyInfoResponseDto, *exceptions.Exception)
	UpdateMyInfo(ctx context.Context, requestDto *apicontract.UpdateMyInfoRequestDto) (*apicontract.UpdateMyInfoResponseDto, *exceptions.Exception)

	// services for public userInfos
	GetPublicUserInfoByUserPublicId(ctx context.Context, publicId uuid.UUID) (*gqlmodels.PublicUserInfo, *exceptions.Exception)
	GetPublicUserInfosByUserPublicIds(ctx context.Context, publicIds []uuid.UUID) ([]*gqlmodels.PublicUserInfo, *exceptions.Exception)
}

type UserInfoService struct {
	validator           *validator.Validate
	db                  *gorm.DB
	userInfoRepository  repositories.UserInfoRepositoryInterface
	userDataCacheClient *userdata.UserDataCacheClient
}

func NewUserInfoService(
	validator *validator.Validate,
	db *gorm.DB,
	userInfoRepository repositories.UserInfoRepositoryInterface,
	userDataCacheClient *userdata.UserDataCacheClient,
) UserInfoServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &UserInfoService{
		validator:           validator,
		db:                  db,
		userInfoRepository:  userInfoRepository,
		userDataCacheClient: userDataCacheClient,
	}
}

/* ============================== Service Methods for UserInfo ============================== */

func (s *UserInfoService) GetMyInfo(
	ctx context.Context, requestDto *apicontract.GetMyInfoRequestDto,
) (*apicontract.GetMyInfoResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"UserInfo",
			"GetMyInfo",
			"User info request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	userInfo, exception := s.userInfoRepository.GetOneByUserId(
		actorUserId,
		options.WithDB(s.db.WithContext(ctx)),
	)
	if exception != nil {
		return nil, exception
	}

	var country *string
	if userInfo.Country != nil {
		countryString := userInfo.Country.String()
		country = &countryString
	}
	return &apicontract.GetMyInfoResponseDto{
		CoverBackgroundURL: userInfo.CoverBackgroundURL,
		AvatarURL:          userInfo.AvatarURL,
		Header:             userInfo.Header,
		Introduction:       userInfo.Introduction,
		Gender:             userInfo.Gender.String(),
		Country:            country,
		BirthDate:          userInfo.BirthDate,
		UpdatedAt:          userInfo.UpdatedAt,
	}, nil
}

func (s *UserInfoService) UpdateMyInfo(
	ctx context.Context, requestDto *apicontract.UpdateMyInfoRequestDto,
) (*apicontract.UpdateMyInfoResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"UserInfo",
			"UpdateMyInfo",
			"User info request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	var gender *enums.UserGender
	if requestDto.Body.Values.Gender != nil {
		parsedGender, err := enums.ConvertStringToUserGender(*requestDto.Body.Values.Gender)
		if err != nil {
			return nil, exceptions.InvalidInput("UserInfo").WithOrigin(err)
		}
		gender = parsedGender
	}
	var country *enums.Country
	if requestDto.Body.Values.Country != nil {
		parsedCountry, err := enums.ConvertStringToCountry(*requestDto.Body.Values.Country)
		if err != nil {
			return nil, exceptions.InvalidInput("UserInfo").WithOrigin(err)
		}
		country = parsedCountry
	}

	updatedUserInfo, exception := s.userInfoRepository.UpdateOneByUserId(
		actorUserId,
		inputs.PartialUpdateUserInfoInput{
			Values: inputs.UpdateUserInfoInput{
				CoverBackgroundURL: requestDto.Body.Values.CoverBackgroundURL,
				AvatarURL:          requestDto.Body.Values.AvatarURL,
				Header:             requestDto.Body.Values.Header,
				Introduction:       requestDto.Body.Values.Introduction,
				Gender:             gender,
				Country:            country,
				BirthDate:          requestDto.Body.Values.BirthDate,
			},
			SetNull: requestDto.Body.SetNull,
		},
		options.WithDB(s.db.WithContext(ctx)),
	)
	if exception != nil {
		return nil, exception
	}

	var user schemas.User
	if result := s.db.WithContext(ctx).
		Select("name").
		Where("id = ?", actorUserId).
		First(&user); result.Error != nil {
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}

	exception = s.userDataCacheClient.Update(user.Name, cacheinputs.UpdateUserDataCacheInput{
		AvatarURL: requestDto.Body.Values.AvatarURL,
	})
	if exception != nil && logs.NotegicLogger != nil {
		logs.NotegicLogger.Error(
			ctx,
			exception.Origin(),
			exception.String(),
		)
	}

	return &apicontract.UpdateMyInfoResponseDto{
		UpdatedAt: updatedUserInfo.UpdatedAt,
	}, nil
}

/* ============================== Service Methods for Public UserInfo (Only available in GraphQL) ============================== */

// use the searchable user cursor (we only give the search functionality on users)
func (s *UserInfoService) GetPublicUserInfoByUserPublicId(
	ctx context.Context,
	publicId uuid.UUID,
) (*gqlmodels.PublicUserInfo, *exceptions.Exception) {
	db := s.db.WithContext(ctx)

	userInfo := schemas.UserInfo{}
	result := db.Model(&schemas.UserInfo{}).
		Joins(`LEFT JOIN "UserTable" u ON u.id = "UserInfoTable".user_id`).
		Where("u.public_id = ?", publicId).
		First(&userInfo)
	if err := result.Error; err != nil {
		return nil, exceptions.New(
			"NotFound",
			"UserInfo",
			"GetPublicUserInfoByUserPublicId",
			"User info was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	return userInfo.ToPublicUserInfo(), nil
}

func (s *UserInfoService) GetPublicUserInfosByUserPublicIds(
	ctx context.Context, publicIds []uuid.UUID,
) ([]*gqlmodels.PublicUserInfo, *exceptions.Exception) {
	if len(publicIds) == 0 {
		return []*gqlmodels.PublicUserInfo{}, nil
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
		return make([]*gqlmodels.PublicUserInfo, len(publicIds)), nil
	}

	var userInfosWithPublicUserIds []*struct {
		schemas.UserInfo
		UserPublicId uuid.UUID `gorm:"column:user_public_id"`
	}
	result := db.Model(&schemas.UserInfo{}).
		Select(`"UserInfoTable".*, u.public_id AS user_public_id`).
		Joins(`LEFT JOIN "UserTable" u ON u.id = "UserInfoTable".user_id`).
		Where("u.public_id IN ?", uniquePublicIds).
		Find(&userInfosWithPublicUserIds)
	if err := result.Error; err != nil {
		return nil, exceptions.New(
			"QueryFailed",
			"UserInfo",
			"GetPublicUserInfosByUserPublicIds",
			"Failed to retrieve user infos",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	publicIdToIndexesMap := make(map[uuid.UUID][]int)
	for index, publidId := range publicIds {
		publicIdToIndexesMap[publidId] = append(publicIdToIndexesMap[publidId], index)
	}

	publicUserInfos := make([]*gqlmodels.PublicUserInfo, len(publicIds))
	for _, userInfoWithPublicUserId := range userInfosWithPublicUserIds {
		for _, index := range publicIdToIndexesMap[userInfoWithPublicUserId.UserPublicId] {
			publicUserInfos[index] = userInfoWithPublicUserId.UserInfo.ToPublicUserInfo()
		}
	}

	return publicUserInfos, nil
}
