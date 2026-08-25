package user

import (
	"context"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-infos"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	userdata "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata"
	cacheinputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata/inputs"
)

type UserInfoServiceInterface interface {
	GetMyInfo(ctx context.Context, requestDto *capi.GetMyInfoRequestDto) (*capi.GetMyInfoResponseDto, *cexceptions.Exception)
	UpdateMyInfo(ctx context.Context, requestDto *capi.UpdateMyInfoRequestDto) (*capi.UpdateMyInfoResponseDto, *cexceptions.Exception)

	// services for public userInfos
	GetPublicUserInfoByUserPublicId(ctx context.Context, publicId uuid.UUID) (*cgqlmodels.PublicUserInfo, *cexceptions.Exception)
	GetPublicUserInfosByUserPublicIds(ctx context.Context, publicIds []uuid.UUID) ([]*cgqlmodels.PublicUserInfo, *cexceptions.Exception)
}

type UserInfoService struct {
	validator           *validator.Validate
	db                  *gorm.DB
	userInfoRepository  srepositories.UserInfoRepositoryInterface
	userDataCacheClient *userdata.UserDataCacheClient
}

func NewUserInfoService(
	validator *validator.Validate,
	db *gorm.DB,
	userInfoRepository srepositories.UserInfoRepositoryInterface,
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
	ctx context.Context, requestDto *capi.GetMyInfoRequestDto,
) (*capi.GetMyInfoResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
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
		srepositories.WithDB(s.db.WithContext(ctx)),
	)
	if exception != nil {
		return nil, exception
	}

	var country *string
	if userInfo.Country != nil {
		countryString := userInfo.Country.String()
		country = &countryString
	}
	return &capi.GetMyInfoResponseDto{
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
	ctx context.Context, requestDto *capi.UpdateMyInfoRequestDto,
) (*capi.UpdateMyInfoResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
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
	var gender *cenums.UserGender
	if requestDto.Body.Values.Gender != nil {
		parsedGender, err := cenums.ConvertStringToUserGender(*requestDto.Body.Values.Gender)
		if err != nil {
			return nil, cexceptions.InvalidInput("UserInfo").WithOrigin(err)
		}
		gender = parsedGender
	}
	var country *cenums.Country
	if requestDto.Body.Values.Country != nil {
		parsedCountry, err := cenums.ConvertStringToCountry(*requestDto.Body.Values.Country)
		if err != nil {
			return nil, cexceptions.InvalidInput("UserInfo").WithOrigin(err)
		}
		country = parsedCountry
	}

	updatedUserInfo, exception := s.userInfoRepository.UpdateOneByUserId(
		actorUserId,
		sinputs.PartialUpdateUserInfoInput{
			Values: sinputs.UpdateUserInfoInput{
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
		srepositories.WithDB(s.db.WithContext(ctx)),
	)
	if exception != nil {
		return nil, exception
	}

	var user sschemas.User
	if result := s.db.WithContext(ctx).
		Select("name").
		Where("id = ?", actorUserId).
		First(&user); result.Error != nil {
		return nil, cexceptions.New(
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
	if exception != nil && slogs.NotegicLogger != nil {
		slogs.NotegicLogger.Error(
			ctx,
			exception.Origin(),
			exception.String(),
		)
	}

	return &capi.UpdateMyInfoResponseDto{
		UpdatedAt: updatedUserInfo.UpdatedAt,
	}, nil
}

/* ============================== Service Methods for Public UserInfo (Only available in GraphQL) ============================== */

// use the searchable user cursor (we only give the search functionality on users)
func (s *UserInfoService) GetPublicUserInfoByUserPublicId(
	ctx context.Context,
	publicId uuid.UUID,
) (*cgqlmodels.PublicUserInfo, *cexceptions.Exception) {
	db := s.db.WithContext(ctx)

	userInfo := sschemas.UserInfo{}
	result := db.Model(&sschemas.UserInfo{}).
		Joins(`LEFT JOIN "UserTable" u ON u.id = "UserInfoTable".user_id`).
		Where("u.public_id = ?", publicId).
		First(&userInfo)
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
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
) ([]*cgqlmodels.PublicUserInfo, *cexceptions.Exception) {
	if len(publicIds) == 0 {
		return []*cgqlmodels.PublicUserInfo{}, nil
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
		return make([]*cgqlmodels.PublicUserInfo, len(publicIds)), nil
	}

	var userInfosWithPublicUserIds []*struct {
		sschemas.UserInfo
		UserPublicId uuid.UUID `gorm:"column:user_public_id"`
	}
	result := db.Model(&sschemas.UserInfo{}).
		Select(`"UserInfoTable".*, u.public_id AS user_public_id`).
		Joins(`LEFT JOIN "UserTable" u ON u.id = "UserInfoTable".user_id`).
		Where("u.public_id IN ?", uniquePublicIds).
		Find(&userInfosWithPublicUserIds)
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
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

	publicUserInfos := make([]*cgqlmodels.PublicUserInfo, len(publicIds))
	for _, userInfoWithPublicUserId := range userInfosWithPublicUserIds {
		for _, index := range publicIdToIndexesMap[userInfoWithPublicUserId.UserPublicId] {
			publicUserInfos[index] = userInfoWithPublicUserId.UserInfo.ToPublicUserInfo()
		}
	}

	return publicUserInfos, nil
}
