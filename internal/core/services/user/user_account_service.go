package user

import (
	"context"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	cauthdto "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/auth"
	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-accounts"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	authservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/auth"
)

type UserAccountServiceInterface interface {
	GetMyAccount(ctx context.Context, requestDto *capi.GetMyAccountRequestDto) (*capi.GetMyAccountResponseDto, *cexceptions.Exception)
	UpdateMyAccount(ctx context.Context, requestDto *capi.UpdateMyAccountRequestDto) (*capi.UpdateMyAccountResponseDto, *cexceptions.Exception)
	BindGoogleAccount(ctx context.Context, requestDto *capi.BindGoogleAccountRequestDto) (*capi.BindGoogleAccountResponseDto, *cexceptions.Exception)
	UnbindGoogleAccount(ctx context.Context, requestDto *capi.UnbindGoogleAccountRequestDto) (*capi.UnbindGoogleAccountResponseDto, *cexceptions.Exception)
}

type UserAccountService struct {
	validator             *validator.Validate
	db                    *gorm.DB
	userRepository        srepositories.UserRepositoryInterface
	userAccountRepository srepositories.UserAccountRepositoryInterface
	userQuotaRepository   srepositories.UserQuotaRepositoryInterface
	oauthService          authservices.OAuthServiceInterface
}

func NewUserAccountService(
	validator *validator.Validate,
	db *gorm.DB,
	userRepository srepositories.UserRepositoryInterface,
	userAccountRepository srepositories.UserAccountRepositoryInterface,
	userQuotaRepository srepositories.UserQuotaRepositoryInterface,
	oauthService authservices.OAuthServiceInterface,
) UserAccountServiceInterface {
	return &UserAccountService{
		validator:             validator,
		db:                    db,
		userRepository:        userRepository,
		userAccountRepository: userAccountRepository,
		userQuotaRepository:   userQuotaRepository,
		oauthService:          oauthService,
	}
}

/* ============================== Service Methods for UserAccount ============================== */

func (s *UserAccountService) GetMyAccount(
	ctx context.Context, requestDto *capi.GetMyAccountRequestDto,
) (*capi.GetMyAccountResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"UserAccount",
			"GetMyAccount",
			"User account request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)

	routineTaskCostUnitUsed, exception := s.userQuotaRepository.GetRoutineTaskCostUnitUsed(
		ctx,
		actorUserId,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	userAccount, exception := s.userAccountRepository.GetOneByUserId(actorUserId, srepositories.WithDB(db))
	if exception != nil {
		return nil, exception
	}

	var countryCode *string
	if userAccount.CountryCode != nil {
		countryCodeString := string(*userAccount.CountryCode)
		countryCode = &countryCodeString
	}
	return &capi.GetMyAccountResponseDto{
		CountryCode:              countryCode,
		PhoneNumber:              userAccount.PhoneNumber,
		GoogleCredential:         userAccount.GoogleCredential,
		DiscordCredential:        userAccount.DiscordCredential,
		RootShelfCount:           userAccount.RootShelfCount,
		BlockPackCount:           userAccount.BlockPackCount,
		BlockCount:               userAccount.BlockCount,
		MaterialCount:            userAccount.MaterialCount,
		WorkflowCount:            userAccount.WorkflowCount,
		AdditionalItemCount:      userAccount.AdditionalItemCount,
		StationCount:             userAccount.StationCount,
		RoutineCount:             userAccount.RoutineCount,
		RoutineTaskCostUnitCount: routineTaskCostUnitUsed,
		RoutineTagCount:          userAccount.RoutineTagCount,
		UpdatedAt:                userAccount.UpdatedAt,
	}, nil
}

func (s *UserAccountService) UpdateMyAccount(
	ctx context.Context, requestDto *capi.UpdateMyAccountRequestDto,
) (*capi.UpdateMyAccountResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"UserAccount",
			"UpdateMyAccount",
			"User account request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	var countryCode *cenums.CountryCode
	if requestDto.Body.Values.CountryCode != nil {
		parsedCountryCode, err := cenums.ConvertStringToCountryCode(*requestDto.Body.Values.CountryCode)
		if err != nil {
			return nil, cexceptions.InvalidInput("UserAccount").WithOrigin(err)
		}
		contractCountryCode := cenums.CountryCode(*parsedCountryCode)
		countryCode = &contractCountryCode
	}

	db := s.db.WithContext(ctx)

	result := db.Model(&sschemas.UserAccount{}).
		Where("user_id = ? AND auth_code = ?", actorUserId, requestDto.Body.AuthCode).
		First(&sschemas.UserAccount{})
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
			"NotFound",
			"UserAccount",
			"UpdateMyAccount",
			"User account was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	_, exception = s.userAccountRepository.UpdateOneByUserId(
		actorUserId,
		sinputs.PartialUpdateUserAccountInput{
			Values: sinputs.UpdateUserAccountInput{
				BackupEmail: requestDto.Body.Values.BackupEmail,
				CountryCode: countryCode,
				PhoneNumber: requestDto.Body.Values.PhoneNumber,
			},
			SetNull: requestDto.Body.SetNull,
		},
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.UpdateMyAccountResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

/* ============================== Service Methods for Binding Accounts ============================== */

func (s *UserAccountService) BindGoogleAccount(
	ctx context.Context, requestDto *capi.BindGoogleAccountRequestDto,
) (*capi.BindGoogleAccountResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"UserAccount",
			"BindGoogleAccount",
			"Google account binding request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	// Start transaction
	db := s.db.WithContext(ctx)

	userInfo, exception := s.oauthService.GetGoogleUserInfo(
		ctx,
		&cauthdto.GetGoogleUserInfoRequestDto{
			AuthenticationCode: requestDto.Body.AuthorizationCode,
		},
	)
	if exception != nil {
		return nil, exception
	}

	user, exception := s.userRepository.GetOneById(
		actorUserId,
		[]sschemas.UserRelation{sschemas.UserRelation_UserAccount},
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	if user.UserAccount.GoogleCredential != nil {
		return nil, cexceptions.New(
			"GoogleCredentialAlreadyBound",
			"UserAccount",
			"BindGoogleAccount",
			"Google credential is already bound",
			http.StatusInternalServerError,
		)
	}

	_, exception = s.userAccountRepository.UpdateOneByUserId(
		actorUserId,
		sinputs.PartialUpdateUserAccountInput{
			Values: sinputs.UpdateUserAccountInput{
				GoogleCredential: &userInfo.Id,
			},
			SetNull: nil,
		},
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.BindGoogleAccountResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *UserAccountService) UnbindGoogleAccount(
	ctx context.Context, requestDto *capi.UnbindGoogleAccountRequestDto,
) (*capi.UnbindGoogleAccountResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
			"InvalidRequest",
			"UserAccount",
			"UnbindGoogleAccount",
			"Google account unbinding request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	// Start transaction
	db := s.db.WithContext(ctx)

	result := db.Model(&sschemas.UserAccount{}).
		Where("user_id = ? AND auth_code = ?", actorUserId, requestDto.Body.AuthCode).
		First(&sschemas.UserAccount{})
	if err := result.Error; err != nil {
		return nil, cexceptions.New(
			"NotFound",
			"UserAccount",
			"UnbindGoogleAccount",
			"User account was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	_, exception = s.userAccountRepository.UpdateOneByUserId(
		actorUserId,
		sinputs.PartialUpdateUserAccountInput{
			Values: sinputs.UpdateUserAccountInput{
				GoogleCredential: nil,
			},
			SetNull: &map[string]bool{
				"GoogleCredential": true,
			},
		},
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.UnbindGoogleAccountResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}
