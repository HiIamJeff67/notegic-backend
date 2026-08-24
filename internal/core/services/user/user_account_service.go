package user

import (
	"context"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-accounts"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	authservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/auth"
)

type UserAccountServiceInterface interface {
	GetMyAccount(ctx context.Context, requestDto *apicontract.GetMyAccountRequestDto) (*apicontract.GetMyAccountResponseDto, *exceptions.Exception)
	UpdateMyAccount(ctx context.Context, requestDto *apicontract.UpdateMyAccountRequestDto) (*apicontract.UpdateMyAccountResponseDto, *exceptions.Exception)
	BindGoogleAccount(ctx context.Context, requestDto *apicontract.BindGoogleAccountRequestDto) (*apicontract.BindGoogleAccountResponseDto, *exceptions.Exception)
	UnbindGoogleAccount(ctx context.Context, requestDto *apicontract.UnbindGoogleAccountRequestDto) (*apicontract.UnbindGoogleAccountResponseDto, *exceptions.Exception)
}

type UserAccountService struct {
	validator             *validator.Validate
	db                    *gorm.DB
	userRepository        repositories.UserRepositoryInterface
	userAccountRepository repositories.UserAccountRepositoryInterface
	userQuotaRepository   repositories.UserQuotaRepositoryInterface
	oauthService          authservices.OAuthServiceInterface
}

func NewUserAccountService(
	validator *validator.Validate,
	db *gorm.DB,
	userRepository repositories.UserRepositoryInterface,
	userAccountRepository repositories.UserAccountRepositoryInterface,
	userQuotaRepository repositories.UserQuotaRepositoryInterface,
	oauthService authservices.OAuthServiceInterface,
) UserAccountServiceInterface {
	if db == nil {
		db = data.DB
	}
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
	ctx context.Context, requestDto *apicontract.GetMyAccountRequestDto,
) (*apicontract.GetMyAccountResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
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
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	userAccount, exception := s.userAccountRepository.GetOneByUserId(actorUserId, options.WithDB(db))
	if exception != nil {
		return nil, exception
	}

	var countryCode *string
	if userAccount.CountryCode != nil {
		countryCodeString := userAccount.CountryCode.String()
		countryCode = &countryCodeString
	}
	return &apicontract.GetMyAccountResponseDto{
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
	ctx context.Context, requestDto *apicontract.UpdateMyAccountRequestDto,
) (*apicontract.UpdateMyAccountResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
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
	var countryCode *enums.CountryCode
	if requestDto.Body.Values.CountryCode != nil {
		parsedCountryCode, err := enums.ConvertStringToCountryCode(*requestDto.Body.Values.CountryCode)
		if err != nil {
			return nil, exceptions.InvalidInput("UserAccount").WithOrigin(err)
		}
		countryCode = parsedCountryCode
	}

	db := s.db.WithContext(ctx)

	result := db.Model(&schemas.UserAccount{}).
		Where("user_id = ? AND auth_code = ?", actorUserId, requestDto.Body.AuthCode).
		First(&schemas.UserAccount{})
	if err := result.Error; err != nil {
		return nil, exceptions.New(
			"NotFound",
			"UserAccount",
			"UpdateMyAccount",
			"User account was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	_, exception = s.userAccountRepository.UpdateOneByUserId(
		actorUserId,
		inputs.PartialUpdateUserAccountInput{
			Values: inputs.UpdateUserAccountInput{
				BackupEmail: requestDto.Body.Values.BackupEmail,
				CountryCode: countryCode,
				PhoneNumber: requestDto.Body.Values.PhoneNumber,
			},
			SetNull: requestDto.Body.SetNull,
		},
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.UpdateMyAccountResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

/* ============================== Service Methods for Binding Accounts ============================== */

func (s *UserAccountService) BindGoogleAccount(
	ctx context.Context, requestDto *apicontract.BindGoogleAccountRequestDto,
) (*apicontract.BindGoogleAccountResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
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

	userInfo, exception := s.oauthService.GetGoogleUserInfo(ctx, requestDto.Body.AuthorizationCode)
	if exception != nil {
		return nil, exception
	}

	user, exception := s.userRepository.GetOneById(
		actorUserId,
		[]schemas.UserRelation{schemas.UserRelation_UserAccount},
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	if user.UserAccount.GoogleCredential != nil {
		return nil, exceptions.New(
			"GoogleCredentialAlreadyBound",
			"UserAccount",
			"BindGoogleAccount",
			"Google credential is already bound",
			http.StatusInternalServerError,
		)
	}

	_, exception = s.userAccountRepository.UpdateOneByUserId(
		actorUserId,
		inputs.PartialUpdateUserAccountInput{
			Values: inputs.UpdateUserAccountInput{
				GoogleCredential: &userInfo.Id,
			},
			SetNull: nil,
		},
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.BindGoogleAccountResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *UserAccountService) UnbindGoogleAccount(
	ctx context.Context, requestDto *apicontract.UnbindGoogleAccountRequestDto,
) (*apicontract.UnbindGoogleAccountResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
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

	result := db.Model(&schemas.UserAccount{}).
		Where("user_id = ? AND auth_code = ?", actorUserId, requestDto.Body.AuthCode).
		First(&schemas.UserAccount{})
	if err := result.Error; err != nil {
		return nil, exceptions.New(
			"NotFound",
			"UserAccount",
			"UnbindGoogleAccount",
			"User account was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	_, exception = s.userAccountRepository.UpdateOneByUserId(
		actorUserId,
		inputs.PartialUpdateUserAccountInput{
			Values: inputs.UpdateUserAccountInput{
				GoogleCredential: nil,
			},
			SetNull: &map[string]bool{
				"GoogleCredential": true,
			},
		},
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.UnbindGoogleAccountResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}
