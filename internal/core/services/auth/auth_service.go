package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	stringutil "github.com/HiIamJeff67/notegic-backend/shared/lib/strings"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	authcode "github.com/HiIamJeff67/notegic-backend/shared/lib/authcode"
	snowflake "github.com/HiIamJeff67/notegic-backend/shared/lib/snowflake"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/auth"
	coreeventscontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	emaildto "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"
	notificationtypescontract "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/types"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	authsql "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/sqls/auth"
	badgesql "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/sqls/badge"
	usersql "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/sqls/user"
	userdata "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata"
	cacheinputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata/inputs"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	emailtransport "github.com/HiIamJeff67/notegic-backend/internal/core/transports/email"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, requestDto *apicontract.RegisterRequestDto) (*apicontract.RegisterResponseDto, *exceptions.Exception)
	RegisterViaGoogle(ctx context.Context, requestDto *apicontract.RegisterViaGoogleRequestDto) (*apicontract.RegisterViaGoogleResponseDto, *exceptions.Exception)
	Login(ctx context.Context, requestDto *apicontract.LoginRequestDto) (*apicontract.LoginResponseDto, *exceptions.Exception)
	LoginViaGoogle(ctx context.Context, requestDto *apicontract.LoginViaGoogleRequestDto) (*apicontract.LoginViaGoogleResponseDto, *exceptions.Exception)
	Logout(ctx context.Context, requestDto *apicontract.LogoutRequestDto) (*apicontract.LogoutResponseDto, *exceptions.Exception)
	SendAuthCode(ctx context.Context, requestDto *apicontract.SendAuthCodeRequestDto) (*apicontract.SendAuthCodeResponseDto, *exceptions.Exception)
	ValidateEmail(ctx context.Context, requestDto *apicontract.ValidateEmailRequestDto) (*apicontract.ValidateEmailResponseDto, *exceptions.Exception)
	ResetEmail(ctx context.Context, requestDto *apicontract.ResetEmailRequestDto) (*apicontract.ResetEmailResponseDto, *exceptions.Exception)
	ForgetPassword(ctx context.Context, requestDto *apicontract.ForgetPasswordRequestDto) (*apicontract.ForgetPasswordResponseDto, *exceptions.Exception)
	ResetMe(ctx context.Context, requestDto *apicontract.ResetMeRequestDto) (*apicontract.ResetMeResponseDto, *exceptions.Exception)
	DeleteMe(ctx context.Context, requestDto *apicontract.DeleteMeRequestDto) (*apicontract.DeleteMeResponseDto, *exceptions.Exception)
}

type AuthService struct {
	validator             *validator.Validate
	db                    *gorm.DB
	userRepository        repositories.UserRepositoryInterface
	userInfoRepository    repositories.UserInfoRepositoryInterface
	userAccountRepository repositories.UserAccountRepositoryInterface
	userSettingRepository repositories.UserSettingRepositoryInterface
	rootShelfRepository   repositories.RootShelfRepositoryInterface
	outboxRepository      repositories.OutboxEventRepositoryInterface
	oauthService          OAuthServiceInterface
	emailClient           emailtransport.ClientInterface
	userDataCacheClient   *userdata.UserDataCacheClient
	authCodeGenerator     *authcode.AuthCodeGenerator
}

func NewAuthService(
	validator *validator.Validate,
	db *gorm.DB,
	userRepository repositories.UserRepositoryInterface,
	userInfoRepository repositories.UserInfoRepositoryInterface,
	userAccountRepository repositories.UserAccountRepositoryInterface,
	userSettingRepository repositories.UserSettingRepositoryInterface,
	rootShelfRepository repositories.RootShelfRepositoryInterface,
	outboxRepository repositories.OutboxEventRepositoryInterface,
	oauthService OAuthServiceInterface,
	emailClient emailtransport.ClientInterface,
	userDataCacheClient *userdata.UserDataCacheClient,
	authCodeGenerator *authcode.AuthCodeGenerator,
) AuthServiceInterface {
	if db == nil {
		db = data.DB
	}
	if authCodeGenerator == nil {
		authCodeGenerator = authcode.New()
	}
	if outboxRepository == nil {
		outboxRepository = repositories.NewOutboxEventRepository()
	}
	return &AuthService{
		validator:             validator,
		db:                    db,
		userRepository:        userRepository,
		userInfoRepository:    userInfoRepository,
		userAccountRepository: userAccountRepository,
		userSettingRepository: userSettingRepository,
		rootShelfRepository:   rootShelfRepository,
		outboxRepository:      outboxRepository,
		oauthService:          oauthService,
		emailClient:           emailClient,
		userDataCacheClient:   userDataCacheClient,
		authCodeGenerator:     authCodeGenerator,
	}
}

/* ============================== Auxiliary Functions ============================== */

var loginCountToBlockDurationMap = map[int32]time.Duration{
	3:  5 * time.Minute,
	5:  15 * time.Minute,
	7:  30 * time.Minute,
	10: 1 * time.Hour,
	15: 6 * time.Hour,
	20: 24 * time.Hour,
	30: 7 * 24 * time.Hour,
}

func (s *AuthService) generateRandomFakeDisplayName() string {
	gofakeit.Seed(0)
	return fmt.Sprintf("%s%s%d", gofakeit.AdjectiveDescriptive(), gofakeit.LastName(), gofakeit.Number(100000, 999999))
}

func (s *AuthService) getLoginBlockedUntilByLoginCount(loginCount int32) (*time.Time, *exceptions.Exception) {
	if loginCount < 0 {
		return nil, exceptions.New("InvalidLoginCount", "Auth", "GetLoginBlockedUntil", "Login count is invalid", http.StatusInternalServerError, true)
	}

	var blockDuration *time.Duration
	for count, duration := range loginCountToBlockDurationMap {
		if loginCount >= count {
			blockDuration = &duration
		}
	}
	if blockDuration == nil {
		return nil, nil
	}
	blockedUntil := time.Now().Add(*blockDuration)
	return &blockedUntil, nil
}

func (s *AuthService) hashPassword(password string) (string, *exceptions.Exception) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", exceptions.New(
			"FailedToGenerateHashValue",
			"Auth",
			"Hash",
			"Failed to generate a hash value",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return string(bytes), nil
}

func (s *AuthService) generateAccessToken(userPublicId uuid.UUID, name string, email string, userAgent string) (*string, *exceptions.Exception) {
	token, err := sharedtokens.GenerateAccessToken(
		userPublicId.String(),
		sharedtokens.AccessTokenClaims{
			Name:      name,
			Email:     email,
			UserAgent: userAgent,
		},
	)
	if err != nil {
		return nil, exceptions.New(
			"GenerationFailed",
			"Token",
			"GenerateAccessToken",
			"Failed to generate the access token",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return token, nil
}

func (s *AuthService) generateRefreshToken(userPublicId uuid.UUID, name string, email string, userAgent string) (*string, *exceptions.Exception) {
	token, err := sharedtokens.GenerateRefreshToken(
		userPublicId.String(),
		sharedtokens.RefreshTokenClaims{
			Name:      name,
			Email:     email,
			UserAgent: userAgent,
		},
	)
	if err != nil {
		return nil, exceptions.New(
			"GenerationFailed",
			"Token",
			"GenerateRefreshToken",
			"Failed to generate the refresh token",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return token, nil
}

func (s *AuthService) generateCSRFToken() (*string, *exceptions.Exception) {
	token, err := sharedtokens.GenerateCSRFToken(sharedtokens.CSRFTokenClaims{})
	if err != nil {
		return nil, exceptions.New(
			"GenerationFailed",
			"Token",
			"GenerateCSRFToken",
			"Failed to generate the CSRF token",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return token, nil
}

func (s *AuthService) enqueueWelcomeNotification(
	tx *gorm.DB,
	userPublicId uuid.UUID,
) *exceptions.Exception {
	payload, err := json.Marshal(notificationtypescontract.NewsPayload{
		Title:   "Welcome to Notegic",
		Summary: "Your Notegic account is ready.",
		Body:    "Start organizing your notes, shelves, and routines in one place.",
	})
	if err != nil {
		return exceptions.New(
			"FailedToMarshal",
			"Notification",
			"Request",
			"Failed to encode the welcome notification payload",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := s.outboxRepository.EnqueueNotificationRequested(
		tx,
		uuid.NewString(),
		coreeventscontract.NotificationRequestedData{
			RecipientUserPublicId: userPublicId,
			Type:                  coreeventscontract.NotificationType_News,
			Priority:              coreeventscontract.NotificationPriority_Normal,
			TemplateKey:           notificationtypescontract.TemplateKey_News,
			TemplateVersion:       1,
			Payload:               payload,
			DedupeKey:             "welcome:" + userPublicId.String(),
		},
	); err != nil {
		return exceptions.New(
			"FailedToCreate",
			"Notification",
			"Request",
			"Failed to enqueue the welcome notification",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return nil
}

/* ============================== Service Methods for Authentication ============================== */

func (s *AuthService) Register(
	ctx context.Context, reqDto *apicontract.RegisterRequestDto,
) (*apicontract.RegisterResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewAuthException().InvalidDto().WithOrigin(err)
	}

	// put the hash part outside the transaction to decrease its blocking time from heavily hashing operation
	hashedPassword, exception := s.hashPassword(reqDto.Body.Password)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	createUserInput := inputs.CreateUserInput{
		Name:        reqDto.Body.Name,
		DisplayName: s.generateRandomFakeDisplayName(), // we generate a default display name for the new user
		Email:       reqDto.Body.Email,
		Password:    hashedPassword,
		UserAgent:   reqDto.Header.UserAgent,
	}
	newUserId, exception := s.userRepository.CreateOne(
		createUserInput,
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	createdUser, exception := s.userRepository.GetOneById(
		*newUserId,
		nil,
		options.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	newAccessToken, exception := s.generateAccessToken(
		createdUser.PublicId,
		createUserInput.Name,
		createUserInput.Email,
		createUserInput.UserAgent,
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newRefreshToken, exception := s.generateRefreshToken(
		createdUser.PublicId,
		createUserInput.Name,
		createUserInput.Email,
		createUserInput.UserAgent,
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newCSRFToken, exception := s.generateCSRFToken()
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	authCode := s.authCodeGenerator.Generate()
	authCodeExpiredAt := s.authCodeGenerator.ExpireAt(time.Now())

	newUser, exception := s.userRepository.UpdateOneById(
		*newUserId,
		inputs.PartialUpdateUserInput{
			Values: inputs.UpdateUserInput{
				RefreshToken: newRefreshToken,
			},
			SetNull: nil,
		},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userInfoRepository.CreateOneByUserId(
		*newUserId,
		inputs.CreateUserInfoInput{},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userAccountRepository.CreateOneByUserId(
		*newUserId,
		inputs.CreateUserAccountInput{
			AuthCode:          authCode,
			AuthCodeExpiredAt: authCodeExpiredAt,
		},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userSettingRepository.CreateOneByUserId(
		*newUserId,
		inputs.CreateUserSettingInput{},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if exception = s.enqueueWelcomeNotification(tx, newUser.PublicId); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	exception = s.userDataCacheClient.Set(
		newUser.Name,
		userdata.UserDataCache{
			Id:          *newUserId,
			PublicId:    newUser.PublicId,
			Name:        newUser.Name,
			DisplayName: newUser.DisplayName,
			Email:       newUser.Email,
			AccessToken: *newAccessToken,
			CSRFToken:   *newCSRFToken,
			Role:        newUser.Role,
			Plan:        newUser.Plan,
			Status:      newUser.Status,
			AvatarURL:   "",
			CreatedAt:   newUser.CreatedAt,
			UpdatedAt:   newUser.UpdatedAt,
		},
	)
	if exception != nil {
		_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
	}

	if exception = s.emailClient.SendWelcomeEmail(ctx, emaildto.SendWelcomeEmailRequestDto{
		To:       newUser.Email,
		UserName: newUser.Name,
		Status:   newUser.Status.String(),
	}); exception != nil {
		_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
	}

	return &apicontract.RegisterResponseDto{
		PublicId:     newUser.PublicId,
		Name:         newUser.Name,
		DisplayName:  newUser.DisplayName,
		Email:        newUser.Email,
		AccessToken:  *newAccessToken,
		RefreshToken: *newRefreshToken,
		CSRFToken:    *newCSRFToken,
		CreatedAt:    newUser.CreatedAt,
	}, nil
}

func (s *AuthService) RegisterViaGoogle(
	ctx context.Context, reqDto *apicontract.RegisterViaGoogleRequestDto,
) (*apicontract.RegisterViaGoogleResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewAuthException().InvalidDto().WithOrigin(err)
	}

	userInfo, exception := s.oauthService.GetGoogleUserInfo(ctx, reqDto.Body.AuthorizationCode)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	existingUser, lookupException := s.userRepository.GetOneByEmail(
		userInfo.Email,
		[]schemas.UserRelation{schemas.UserRelation_UserAccount},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if lookupException == nil && existingUser != nil {
		loginResponse, loginException := s.loginViaGoogleUser(
			ctx,
			tx,
			existingUser,
			userInfo,
			reqDto.Header.UserAgent,
		)
		if loginException != nil {
			return nil, loginException
		}

		return &apicontract.RegisterViaGoogleResponseDto{
			PublicId:     loginResponse.PublicId,
			Name:         loginResponse.Name,
			DisplayName:  loginResponse.DisplayName,
			Email:        loginResponse.Email,
			AccessToken:  loginResponse.AccessToken,
			RefreshToken: loginResponse.RefreshToken,
			CSRFToken:    loginResponse.CSRFToken,
			CreatedAt:    loginResponse.CreatedAt,
		}, nil
	}
	if lookupException != nil &&
		(lookupException.Reason != "NotFound" || lookupException.Domain != "User") {
		tx.Rollback()
		return nil, lookupException
	}

	fakePasswordBytes, err := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"FailedToGenerateHashValue",
			"Auth",
			"Hash",
			"Failed to generate a hash value",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	fakePassword := string(fakePasswordBytes)

	hashedPassword, exception := s.hashPassword(fakePassword)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	reg, err := regexp.Compile("[^a-z0-9]+")
	if err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewAuthException().FailedToCompileRegularExpression().WithOrigin(err)
	}
	fakeName := strings.ToLower(uuid.New().String())
	fakeName = reg.ReplaceAllString(fakeName, "")
	if len(fakeName) < 6 {
		fakeName += snowflake.GenerateRepeatableID()
	}
	if len(fakeName) > constants.MaxNameLength {
		fakeName = fakeName[:constants.MaxNameLength]
	}

	createUserInput := inputs.CreateUserInput{
		Name:        fakeName,
		DisplayName: s.generateRandomFakeDisplayName(), // we generate a default display name for the new user
		Email:       userInfo.Email,
		Password:    hashedPassword,
		UserAgent:   reqDto.Header.UserAgent,
	}
	newUserId, exception := s.userRepository.CreateOne(
		createUserInput,
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	createdUser, exception := s.userRepository.GetOneById(
		*newUserId,
		nil,
		options.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	newAccessToken, exception := s.generateAccessToken(
		createdUser.PublicId,
		createUserInput.Name,
		createUserInput.Email,
		createUserInput.UserAgent,
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newRefreshToken, exception := s.generateRefreshToken(
		createdUser.PublicId,
		createUserInput.Name,
		createUserInput.Email,
		createUserInput.UserAgent,
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newCSRFToken, exception := s.generateCSRFToken()
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	authCode := s.authCodeGenerator.Generate()
	authCodeExpiredAt := s.authCodeGenerator.ExpireAt(time.Now())

	newUser, exception := s.userRepository.UpdateOneById(
		*newUserId,
		inputs.PartialUpdateUserInput{
			Values: inputs.UpdateUserInput{
				RefreshToken: newRefreshToken,
			},
			SetNull: nil,
		},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userInfoRepository.CreateOneByUserId(
		*newUserId,
		inputs.CreateUserInfoInput{},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userAccountRepository.CreateOneByUserId(
		*newUserId,
		inputs.CreateUserAccountInput{
			AuthCode:          authCode,
			AuthCodeExpiredAt: authCodeExpiredAt,
			GoogleCredential:  &userInfo.Id,
		},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userSettingRepository.CreateOneByUserId(
		*newUserId,
		inputs.CreateUserSettingInput{},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	exception = s.userDataCacheClient.Set(
		newUser.Name,
		userdata.UserDataCache{
			Id:          *newUserId,
			PublicId:    newUser.PublicId,
			Name:        newUser.Name,
			DisplayName: newUser.DisplayName,
			Email:       newUser.Email,
			AccessToken: *newAccessToken,
			CSRFToken:   *newCSRFToken,
			Role:        newUser.Role,
			Plan:        newUser.Plan,
			Status:      newUser.Status,
			AvatarURL:   "",
			CreatedAt:   newUser.CreatedAt,
			UpdatedAt:   newUser.UpdatedAt,
		},
	)
	if exception != nil {
		_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
	}

	if exception = s.enqueueWelcomeNotification(tx, newUser.PublicId); exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	// send the welcome email to the registered user
	if exception = s.emailClient.SendWelcomeEmail(ctx, emaildto.SendWelcomeEmailRequestDto{
		To:       newUser.Email,
		UserName: newUser.Name,
		Status:   newUser.Status.String(),
	}); exception != nil {
		_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
	}

	return &apicontract.RegisterViaGoogleResponseDto{
		PublicId:     newUser.PublicId,
		Name:         newUser.Name,
		DisplayName:  newUser.DisplayName,
		Email:        newUser.Email,
		AccessToken:  *newAccessToken,
		RefreshToken: *newRefreshToken,
		CSRFToken:    *newCSRFToken,
		CreatedAt:    newUser.CreatedAt,
	}, nil
}

func (s *AuthService) Login(
	ctx context.Context, reqDto *apicontract.LoginRequestDto,
) (*apicontract.LoginResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewUserException().InvalidInput().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()

	// otherwise, the user should provide their account and password
	var user *schemas.User = nil
	var exception *exceptions.Exception = nil
	if stringutil.IsAlphaAndNumberString(reqDto.Body.Account) { // if the account field contains user name
		if user, exception = s.userRepository.GetOneByName(
			reqDto.Body.Account,
			nil,
			options.WithTransactionDB(tx),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		); exception != nil {
			tx.Rollback()
			return nil, exception
		}
	} else if stringutil.IsEmailString(reqDto.Body.Account) { // if the account field contains email
		if user, exception = s.userRepository.GetOneByEmail(
			reqDto.Body.Account,
			nil,
			options.WithTransactionDB(tx),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		); exception != nil {
			tx.Rollback()
			return nil, exception
		}
	} else {
		tx.Rollback()
		return nil, apiexceptions.NewAuthException().InvalidDto()
	}

	if user == nil {
		tx.Rollback()
		return nil, apiexceptions.NewAuthException().InvalidDto()
	}

	if user.BlockLoginUntil.After(time.Now()) {
		tx.Rollback()
		return nil, apiexceptions.NewAuthException().LoginBlockedDueToTryingTooManyTimes(user.BlockLoginUntil)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reqDto.Body.Password)) != nil {
		newLoginCount := user.LoginCount + 1
		blockLoginUntil, exception := s.getLoginBlockedUntilByLoginCount(newLoginCount)
		if exception != nil {
			tx.Rollback()
			return nil, exception
		}

		_, exception = s.userRepository.UpdateOneById(
			user.Id,
			inputs.PartialUpdateUserInput{
				Values: inputs.UpdateUserInput{
					LoginCount:     &newLoginCount,
					BlockLoginUtil: blockLoginUntil,
				},
				SetNull: nil,
			},
			options.WithTransactionDB(tx),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		)
		if exception != nil {
			tx.Rollback()
			return nil, exception
		}

		if blockLoginUntil != nil {
			tx.Rollback()
			return nil, apiexceptions.NewAuthException().LoginBlockedDueToTryingTooManyTimes(*blockLoginUntil)
		}

		tx.Rollback()
		return nil, apiexceptions.NewAuthException().WrongPassword() // login procedure early ends here
	}

	if user.UserAgent != reqDto.Header.UserAgent {
		// send a security email to warn the user
		if exception := s.emailClient.SendSecurityAlertEmail(ctx, emaildto.SendSecurityAlertEmailRequestDto{
			To:               user.Email,
			UserName:         user.Name,
			Status:           user.Status.String(),
			AlertType:        "Login in Different Place",
			Reason:           "Your account has a recent login action in other place",
			TimeOfOccurrence: time.Now(),
			OtherDetails:     "",
		}); exception != nil {
			_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
		}
	}

	newAccessToken, exception := s.generateAccessToken(user.PublicId, user.Name, user.Email, user.UserAgent)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newRefreshToken, exception := s.generateRefreshToken(user.PublicId, user.Name, user.Email, user.UserAgent)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newCSRFToken, exception := s.generateCSRFToken()
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	// check if the user data cache exists
	if _, exception := s.userDataCacheClient.Get(user.Name); exception == nil {
		// then just update the existing user data cache
		if exception = s.userDataCacheClient.Update(
			user.Name,
			cacheinputs.UpdateUserDataCacheInput{
				AccessToken: newAccessToken,
				CSRFToken:   newCSRFToken,
			},
		); exception != nil {
			_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
		}
	} else { // else if it does not exist
		// then we have to first get the relative data from different tables
		// we done this by one custom sql so it's not that slow...
		// once we have the required data, we set it as the user data cache
		output := struct {
			Id          uuid.UUID        `gorm:"id"`
			PublicId    uuid.UUID        `gorm:"public_id"`
			Name        string           `gorm:"name"`
			DisplayName string           `gorm:"display_name"`
			Email       string           `gorm:"email"`
			Role        enums.UserRole   `gorm:"role"`
			Plan        enums.UserPlan   `gorm:"plan"`
			Status      enums.UserStatus `gorm:"status"`
			AvatarURL   *string          `gorm:"avatar_url"`
			CreatedAt   time.Time        `gorm:"created_at"`
			UpdatedAt   time.Time        `gorm:"updated_at"`
		}{}
		err := tx.Raw(usersql.GetUserDataCacheByIdSQL, user.Id).
			Row().
			Scan(
				&output.Id,
				&output.PublicId,
				&output.Name,
				&output.DisplayName,
				&output.Email,
				&output.Role,
				&output.Plan,
				&output.Status,
				&output.AvatarURL,
				&output.CreatedAt,
				&output.UpdatedAt,
			)
		if err != nil {
			tx.Rollback()
			return nil, apiexceptions.NewUserException().NotFound().WithOrigin(err)
		}

		newUserDataCache := userdata.UserDataCache{
			Id:          user.Id,
			PublicId:    output.PublicId,
			Name:        output.Name,
			DisplayName: output.DisplayName,
			Email:       output.Email,
			AccessToken: *newAccessToken,
			CSRFToken:   *newCSRFToken,
			Role:        output.Role,
			Plan:        output.Plan,
			Status:      output.Status,
			AvatarURL:   "",
			CreatedAt:   output.CreatedAt,
			UpdatedAt:   output.UpdatedAt,
		}
		if output.AvatarURL != nil {
			newUserDataCache.AvatarURL = *output.AvatarURL
		}
		exception := s.userDataCacheClient.Set(
			user.Name,
			newUserDataCache,
		)
		if exception != nil {
			tx.Rollback()
			return nil, exception
		}
	}

	// update the refresh token and the status of the user
	var zeroLoginCount int32 = 0 // reset the login count if the login procedure is valid
	updatedUser, exception := s.userRepository.UpdateOneById(
		user.Id,
		inputs.PartialUpdateUserInput{
			Values: inputs.UpdateUserInput{
				Status:       &user.PrevStatus,
				RefreshToken: newRefreshToken,
				UserAgent:    &reqDto.Header.UserAgent,
				LoginCount:   &zeroLoginCount,
			},
			SetNull: nil,
		},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &apicontract.LoginResponseDto{
		PublicId:     user.PublicId,
		Name:         user.Name,
		DisplayName:  user.DisplayName,
		Email:        user.Email,
		AccessToken:  *newAccessToken,
		RefreshToken: updatedUser.RefreshToken,
		CSRFToken:    *newCSRFToken,
		UpdatedAt:    updatedUser.UpdatedAt,
		CreatedAt:    user.CreatedAt,
	}, nil
}

func (s *AuthService) LoginViaGoogle(
	ctx context.Context, reqDto *apicontract.LoginViaGoogleRequestDto,
) (*apicontract.LoginViaGoogleResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewAuthException().InvalidDto().WithOrigin(err)
	}

	userInfo, exception := s.oauthService.GetGoogleUserInfo(ctx, reqDto.Body.AuthorizationCode)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	user, exception := s.userRepository.GetOneByEmail(
		userInfo.Email,
		[]schemas.UserRelation{schemas.UserRelation_UserAccount},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if user == nil {
		tx.Rollback()
		return nil, apiexceptions.NewAuthException().InvalidDto()
	}

	return s.loginViaGoogleUser(ctx, tx, user, userInfo, reqDto.Header.UserAgent)
}

func (s *AuthService) loginViaGoogleUser(
	ctx context.Context,
	tx *gorm.DB,
	user *schemas.User,
	userInfo *googleUserInfo,
	userAgent string,
) (*apicontract.LoginViaGoogleResponseDto, *exceptions.Exception) {

	if user.BlockLoginUntil.After(time.Now()) {
		tx.Rollback()
		return nil, apiexceptions.NewAuthException().LoginBlockedDueToTryingTooManyTimes(user.BlockLoginUntil)
	}

	if user.UserAccount.GoogleCredential == nil || userInfo.Id != *user.UserAccount.GoogleCredential {
		newLoginCount := user.LoginCount + 1
		blockLoginUntil, exception := s.getLoginBlockedUntilByLoginCount(newLoginCount)
		if exception != nil {
			tx.Rollback()
			return nil, exception
		}

		_, exception = s.userRepository.UpdateOneById(
			user.Id,
			inputs.PartialUpdateUserInput{
				Values: inputs.UpdateUserInput{
					LoginCount:     &newLoginCount,
					BlockLoginUtil: blockLoginUntil,
				},
				SetNull: nil,
			},
			options.WithTransactionDB(tx),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		)
		if exception != nil {
			tx.Rollback()
			return nil, exception
		}

		if blockLoginUntil != nil {
			tx.Rollback()
			return nil, apiexceptions.NewAuthException().LoginBlockedDueToTryingTooManyTimes(*blockLoginUntil)
		}

		tx.Rollback()
		return nil, apiexceptions.NewAuthException().WrongPassword() // login via google procedure early ends here
	}

	if user.UserAgent != userAgent {
		// send a security email to warn the user
		if exception := s.emailClient.SendSecurityAlertEmail(ctx, emaildto.SendSecurityAlertEmailRequestDto{
			To:               user.Email,
			UserName:         user.Name,
			Status:           user.Status.String(),
			AlertType:        "Login in Different Place",
			Reason:           "Your account has a recent login action in other place",
			TimeOfOccurrence: time.Now(),
			OtherDetails:     "",
		}); exception != nil {
			_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
		}
	}

	newAccessToken, exception := s.generateAccessToken(user.PublicId, user.Name, user.Email, user.UserAgent)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newRefreshToken, exception := s.generateRefreshToken(user.PublicId, user.Name, user.Email, user.UserAgent)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newCSRFToken, exception := s.generateCSRFToken()
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	// check if the user data cache exists
	if _, exception := s.userDataCacheClient.Get(user.Name); exception == nil {
		// then just update the existing user data cache
		if exception = s.userDataCacheClient.Update(
			user.Name,
			cacheinputs.UpdateUserDataCacheInput{
				AccessToken: newAccessToken,
				CSRFToken:   newCSRFToken,
			},
		); exception != nil {
			_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
		}
	} else { // else if it does not exist
		// then we have to first get the relative data from different tables
		// we done this by one custom sql so it's not that slow...
		// once we have the required data, we set it as the user data cache
		output := struct {
			Id          uuid.UUID        `gorm:"id"`
			PublicId    uuid.UUID        `gorm:"public_id"`
			Name        string           `gorm:"name"`
			DisplayName string           `gorm:"display_name"`
			Email       string           `gorm:"email"`
			Role        enums.UserRole   `gorm:"role"`
			Plan        enums.UserPlan   `gorm:"plan"`
			Status      enums.UserStatus `gorm:"status"`
			AvatarURL   *string          `gorm:"avatar_url"`
			CreatedAt   time.Time        `gorm:"created_at"`
			UpdatedAt   time.Time        `gorm:"updated_at"`
		}{}
		err := tx.Raw(usersql.GetUserDataCacheByIdSQL, user.Id).
			Row().
			Scan(
				&output.Id,
				&output.PublicId,
				&output.Name,
				&output.DisplayName,
				&output.Email,
				&output.Role,
				&output.Plan,
				&output.Status,
				&output.AvatarURL,
				&output.CreatedAt,
				&output.UpdatedAt,
			)
		if err != nil {
			tx.Rollback()
			return nil, apiexceptions.NewUserException().NotFound().WithOrigin(err)
		}

		newUserDataCache := userdata.UserDataCache{
			Id:          user.Id,
			PublicId:    output.PublicId,
			Name:        output.Name,
			DisplayName: output.DisplayName,
			Email:       output.Email,
			AccessToken: *newAccessToken,
			CSRFToken:   *newCSRFToken,
			Role:        output.Role,
			Plan:        output.Plan,
			Status:      output.Status,
			AvatarURL:   "",
			CreatedAt:   output.CreatedAt,
			UpdatedAt:   output.UpdatedAt,
		}
		if output.AvatarURL != nil {
			newUserDataCache.AvatarURL = *output.AvatarURL
		}
		exception := s.userDataCacheClient.Set(
			user.Name,
			newUserDataCache,
		)
		if exception != nil {
			tx.Rollback()
			return nil, exception
		}
	}

	// update the refresh token and the status of the user
	var zeroLoginCount int32 = 0 // reset the login count if the login procedure is valid
	updatedUser, exception := s.userRepository.UpdateOneById(
		user.Id,
		inputs.PartialUpdateUserInput{
			Values: inputs.UpdateUserInput{
				Status:       &user.PrevStatus,
				RefreshToken: newRefreshToken,
				UserAgent:    &userAgent,
				LoginCount:   &zeroLoginCount,
			},
			SetNull: nil,
		},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &apicontract.LoginViaGoogleResponseDto{
		PublicId:     user.PublicId,
		Name:         user.Name,
		DisplayName:  user.DisplayName,
		Email:        user.Email,
		AccessToken:  *newAccessToken,
		RefreshToken: updatedUser.RefreshToken,
		CSRFToken:    *newCSRFToken,
		UpdatedAt:    updatedUser.UpdatedAt,
		CreatedAt:    user.CreatedAt,
	}, nil
}

func (s *AuthService) Logout(
	ctx context.Context, reqDto *apicontract.LogoutRequestDto,
) (*apicontract.LogoutResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewAuthException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserPublicId, exception := contexts.GetActorUserPublicId(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserName, exception := contexts.GetActorUserName(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, exceptions.New(
			"TransactionBeginFailed",
			"Auth",
			"Logout",
			"Failed to begin the logout transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	offlineStatus := enums.UserStatus_Offline
	emptyString := ""
	updatedUser, exception := s.userRepository.UpdateOneById(
		actorUserId,
		inputs.PartialUpdateUserInput{
			Values: inputs.UpdateUserInput{
				Status:       &offlineStatus,
				RefreshToken: &emptyString,
			},
			SetNull: nil,
		},
		options.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := s.outboxRepository.EnqueueUserSessionsRevoked(
		tx,
		actorUserPublicId.String(),
		actorUserPublicId,
	); err != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"FailedToCreate",
			"Outbox",
			"Logout",
			"Failed to create user session revocation event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	exception = s.userDataCacheClient.Delete(actorUserName)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.LogoutResponseDto{
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}

func (s *AuthService) SendAuthCode(
	ctx context.Context, reqDto *apicontract.SendAuthCodeRequestDto,
) (*apicontract.SendAuthCodeResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewUserException().InvalidInput().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	authCode := s.authCodeGenerator.Generate()
	authCodeExpiredAt := s.authCodeGenerator.ExpireAt(time.Now())
	blockAuthCodeUntil := time.Now().Add(60 * time.Second)
	output := struct {
		Name               string    `json:"name" gorm:"column:name;"`
		UserAgent          string    `json:"userAgent" gorm:"column:user_agent;"`
		BlockAuthCodeUntil time.Time `json:"blockAuthCodeUntil" gorm:"column:block_auth_code_until;"`
		Now                time.Time `json:"now" gorm:"column:now;"`
	}{}
	err := db.Raw(authsql.UpdateAuthCodeSQL,
		authCode, authCodeExpiredAt, blockAuthCodeUntil, reqDto.Body.Email,
	).Row().
		Scan(&output.Name, &output.UserAgent, &output.BlockAuthCodeUntil, &output.Now)
	if err != nil {
		return nil, apiexceptions.NewAuthException().AuthCodeBlockedDueToTryingTooManyTimes(output.BlockAuthCodeUntil).WithOrigin(err)
	}

	if exception := s.emailClient.SendValidationEmail(ctx, emaildto.SendValidationEmailRequestDto{
		To:        reqDto.Body.Email,
		UserName:  output.Name,
		AuthCode:  authCode,
		UserAgent: output.UserAgent,
		ExpiredAt: authCodeExpiredAt,
	}); exception != nil {
		return nil, exception
	}

	return &apicontract.SendAuthCodeResponseDto{
		AuthCodeExpiredAt:  authCodeExpiredAt,
		BlockAuthCodeUntil: blockAuthCodeUntil,
		UpdatedAt:          time.Now(),
	}, nil
}

func (s *AuthService) ValidateEmail(
	ctx context.Context, reqDto *apicontract.ValidateEmailRequestDto,
) (*apicontract.ValidateEmailResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewUserException().InvalidInput().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)

	var updatedAt time.Time
	err := db.Raw(authsql.ValidateEmailSQL, actorUserId, reqDto.Body.AuthCode).
		Row().
		Scan(&updatedAt)
	if err != nil {
		return nil, apiexceptions.NewUserException().FailedToUpdate().WithOrigin(err)
	}

	return &apicontract.ValidateEmailResponseDto{
		UpdatedAt: updatedAt,
	}, nil
}

func (s *AuthService) ResetEmail(
	ctx context.Context, reqDto *apicontract.ResetEmailRequestDto,
) (*apicontract.ResetEmailResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewUserException().InvalidInput().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	var updatedAt time.Time
	err := tx.Raw(authsql.ResetEmailSQL, reqDto.Body.NewEmail, reqDto.Body.AuthCode, actorUserId).
		Row().
		Scan(&updatedAt)
	if err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToUpdate().WithOrigin(err)
	}

	authCode := s.authCodeGenerator.Generate()
	authCodeExpiredAt := s.authCodeGenerator.ExpireAt(time.Now())
	_, exception = s.userAccountRepository.UpdateOneByUserId(
		actorUserId,
		inputs.PartialUpdateUserAccountInput{
			Values: inputs.UpdateUserAccountInput{
				AuthCode:          &authCode,
				AuthCodeExpiredAt: &authCodeExpiredAt,
			},
			SetNull: nil,
		},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &apicontract.ResetEmailResponseDto{
		UpdatedAt: updatedAt,
	}, nil
}

func (s *AuthService) ForgetPassword(
	ctx context.Context, reqDto *apicontract.ForgetPasswordRequestDto,
) (*apicontract.ForgetPasswordResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewUserException().InvalidInput().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()

	var user *schemas.User = nil
	var exception *exceptions.Exception = nil
	var preloads = []schemas.UserRelation{schemas.UserRelation_UserAccount, schemas.UserRelation_UserInfo, schemas.UserRelation_UserSetting}
	if stringutil.IsEmailString(reqDto.Body.Account) { // if the account field contains email
		if user, exception = s.userRepository.GetOneByEmail(
			reqDto.Body.Account,
			preloads,
			options.WithTransactionDB(tx),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		); exception != nil {
			tx.Rollback()
			return nil, exception
		}
	} else if stringutil.IsAlphaAndNumberString(reqDto.Body.Account) { // if the account field contains user name
		if user, exception = s.userRepository.GetOneByName(
			reqDto.Body.Account,
			preloads,
			options.WithTransactionDB(tx),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		); exception != nil {
			tx.Rollback()
			return nil, exception
		}
	} else {
		tx.Rollback()
		return nil, apiexceptions.NewAuthException().InvalidDto()
	}

	if reqDto.Body.AuthCode != user.UserAccount.AuthCode {
		tx.Rollback()
		return nil, apiexceptions.NewAuthException().WrongAuthCode()
	}

	newAccessToken, exception := s.generateAccessToken(user.PublicId, user.Name, user.Email, user.UserAgent)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newRefreshToken, exception := s.generateRefreshToken(user.PublicId, user.Name, user.Email, user.UserAgent)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newCSRFToken, exception := s.generateCSRFToken()
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	// update the access token of the user
	exception = s.userDataCacheClient.Update(user.Name, cacheinputs.UpdateUserDataCacheInput{AccessToken: newAccessToken})
	if exception != nil {
		_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
		// and also try to set the new user cache data
		exception = s.userDataCacheClient.Set(user.Name, userdata.UserDataCache{
			Id:          user.Id,
			PublicId:    user.PublicId,
			Name:        user.Name,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			AccessToken: *newAccessToken,
			CSRFToken:   *newCSRFToken,
			Role:        user.Role,
			Plan:        user.Plan,
			Status:      user.Status,
			AvatarURL:   *user.UserInfo.AvatarURL,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
		if exception != nil {
			_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
		}
	}

	hashedPassword, exception := s.hashPassword(reqDto.Body.NewPassword)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	// update the refresh token and the status of the user
	var zeroLoginCount int32 = 0 // reset the login count if the login procedure is valid
	updatedUser, exception := s.userRepository.UpdateOneById(
		user.Id,
		inputs.PartialUpdateUserInput{
			Values: inputs.UpdateUserInput{
				Password:     &hashedPassword,
				RefreshToken: newRefreshToken,
				UserAgent:    &reqDto.Header.UserAgent,
				LoginCount:   &zeroLoginCount,
			},
			SetNull: nil,
		},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := repositories.NewOutboxEventRepository().EnqueueUserSessionsRevoked(
		tx,
		user.PublicId.String(),
		user.PublicId,
	); err != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"FailedToCreate",
			"Outbox",
			"ForgetPassword",
			"Failed to create user session revocation event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &apicontract.ForgetPasswordResponseDto{
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}

func (s *AuthService) ResetMe(
	ctx context.Context, reqDto *apicontract.ResetMeRequestDto,
) (*apicontract.ResetMeResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewUserException().InvalidInput().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	tx := s.db.WithContext(ctx).Begin()

	// Instead of deleting the user, we recreate their relative data in the database
	// and make sure not to update the access token and refresh token, and csrf token in the reset logic
	// Note that the user will not logged out after the reset operation

	// try to retrieve the target user to reset and validate his/her auth code first
	var resetUserAccount schemas.UserAccount
	result := tx.Model(&resetUserAccount).
		Where("user_id = ? AND auth_code = ?", actorUserId, reqDto.Body.AuthCode).
		First(&resetUserAccount)
	if err := result.Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserAccountException().NotFound().WithOrigin(err)
	}

	// delete the user info
	if err := tx.Where("user_id = ?", actorUserId).Delete(&schemas.UserInfo{}).Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserInfoException().FailedToDelete().WithOrigin(err)
	}
	// and then re-create a new user info
	if _, exception := s.userInfoRepository.CreateOneByUserId(
		resetUserAccount.UserId,
		inputs.CreateUserInfoInput{},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}

	// delete the user setting
	if err := tx.Where("user_id = ?", actorUserId).Delete(&schemas.UserSetting{}).Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserSettingException().FailedToDelete().WithOrigin(err)
	}
	// and then re-create a new user setting
	if _, exception := s.userSettingRepository.CreateOneByUserId(
		resetUserAccount.UserId,
		inputs.CreateUserSettingInput{},
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}

	// delete all the badges of the user
	if err := tx.Exec(badgesql.DeleteAllMyBadgesSQL, actorUserId).Error; err != nil {
		// skip if there's no users to badges to delete
	}

	// soft delete all the root shelves of the user
	if exception := s.rootShelfRepository.SoftDeleteManyByUserId(
		actorUserId,
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	); exception != nil {
		// skip if there's no root shelves to soft delete
	} else {
		// then hard delete all the root shelves of the user
		if exception := s.rootShelfRepository.HardDeleteManyByUserId(
			actorUserId,
			options.WithTransactionDB(tx),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		); exception != nil {
			// skip if there's no root shelves to hard delete
		}
	}

	// delete other stuff in the future...

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithDetails(err)
	}

	return &apicontract.ResetMeResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *AuthService) DeleteMe(
	ctx context.Context, reqDto *apicontract.DeleteMeRequestDto,
) (*apicontract.DeleteMeResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewUserException().InvalidInput().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserName, exception := contexts.GetActorUserName(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserPublicId, exception := contexts.GetActorUserPublicId(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, exceptions.New(
			"TransactionBeginFailed",
			"Auth",
			"DeleteMe",
			"Failed to begin the delete account transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	deleteResult := tx.Exec(authsql.DeleteMeSQL, actorUserId, reqDto.Body.AuthCode)
	if deleteResult.Error != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToDelete().WithOrigin(deleteResult.Error)
	}
	if deleteResult.RowsAffected == 0 {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToDelete()
	}
	if err := s.outboxRepository.EnqueueUserSessionsRevoked(
		tx,
		actorUserPublicId.String(),
		actorUserPublicId,
	); err != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMe",
			"Failed to create user session revocation event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	exception = s.userDataCacheClient.Delete(actorUserName)
	if exception != nil {
		_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
	}

	return &apicontract.DeleteMeResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *AuthService) RegisterViaMeta() {}

func (s *AuthService) RegisterViaGithub() {}

func (s *AuthService) LoginViaMeta() {}

func (s *AuthService) LoginViaGithub() {}
