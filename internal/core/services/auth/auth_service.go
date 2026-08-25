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

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/auth"
	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cemaildto "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"
	cnotificationtypes "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/types"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	sauthcode "github.com/HiIamJeff67/notegic-backend/shared/lib/authcode"
	snowflake "github.com/HiIamJeff67/notegic-backend/shared/lib/snowflake"
	sstrings "github.com/HiIamJeff67/notegic-backend/shared/lib/strings"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	authsql "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/sqls/auth"
	badgesql "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/sqls/badge"
	usersql "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/sqls/user"
	userdata "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata"
	cacheinputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata/inputs"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	emailtransport "github.com/HiIamJeff67/notegic-backend/internal/core/transports/email"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, requestDto *capi.RegisterRequestDto) (*capi.RegisterResponseDto, *cexceptions.Exception)
	RegisterViaGoogle(ctx context.Context, requestDto *capi.RegisterViaGoogleRequestDto) (*capi.RegisterViaGoogleResponseDto, *cexceptions.Exception)
	Login(ctx context.Context, requestDto *capi.LoginRequestDto) (*capi.LoginResponseDto, *cexceptions.Exception)
	LoginViaGoogle(ctx context.Context, requestDto *capi.LoginViaGoogleRequestDto) (*capi.LoginViaGoogleResponseDto, *cexceptions.Exception)
	Logout(ctx context.Context, requestDto *capi.LogoutRequestDto) (*capi.LogoutResponseDto, *cexceptions.Exception)
	SendAuthCode(ctx context.Context, requestDto *capi.SendAuthCodeRequestDto) (*capi.SendAuthCodeResponseDto, *cexceptions.Exception)
	ValidateEmail(ctx context.Context, requestDto *capi.ValidateEmailRequestDto) (*capi.ValidateEmailResponseDto, *cexceptions.Exception)
	ResetEmail(ctx context.Context, requestDto *capi.ResetEmailRequestDto) (*capi.ResetEmailResponseDto, *cexceptions.Exception)
	ForgetPassword(ctx context.Context, requestDto *capi.ForgetPasswordRequestDto) (*capi.ForgetPasswordResponseDto, *cexceptions.Exception)
	ResetMe(ctx context.Context, requestDto *capi.ResetMeRequestDto) (*capi.ResetMeResponseDto, *cexceptions.Exception)
	DeleteMe(ctx context.Context, requestDto *capi.DeleteMeRequestDto) (*capi.DeleteMeResponseDto, *cexceptions.Exception)
}

type AuthService struct {
	validator             *validator.Validate
	db                    *gorm.DB
	userRepository        srepositories.UserRepositoryInterface
	userInfoRepository    srepositories.UserInfoRepositoryInterface
	userAccountRepository srepositories.UserAccountRepositoryInterface
	userSettingRepository srepositories.UserSettingRepositoryInterface
	rootShelfRepository   srepositories.RootShelfRepositoryInterface
	outboxRepository      srepositories.OutboxEventRepositoryInterface
	oauthService          OAuthServiceInterface
	emailClient           emailtransport.ClientInterface
	userDataCacheClient   *userdata.UserDataCacheClient
	authCodeGenerator     *sauthcode.AuthCodeGenerator
}

func NewAuthService(
	validator *validator.Validate,
	db *gorm.DB,
	userRepository srepositories.UserRepositoryInterface,
	userInfoRepository srepositories.UserInfoRepositoryInterface,
	userAccountRepository srepositories.UserAccountRepositoryInterface,
	userSettingRepository srepositories.UserSettingRepositoryInterface,
	rootShelfRepository srepositories.RootShelfRepositoryInterface,
	outboxRepository srepositories.OutboxEventRepositoryInterface,
	oauthService OAuthServiceInterface,
	emailClient emailtransport.ClientInterface,
	userDataCacheClient *userdata.UserDataCacheClient,
	authCodeGenerator *sauthcode.AuthCodeGenerator,
) AuthServiceInterface {
	if db == nil {
		db = data.DB
	}
	if authCodeGenerator == nil {
		authCodeGenerator = sauthcode.New()
	}
	if outboxRepository == nil {
		outboxRepository = srepositories.NewOutboxEventRepository()
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

func (s *AuthService) loginByGoogleUserInfo(
	ctx context.Context,
	tx *gorm.DB,
	user *sschemas.User,
	userInfo *capi.GetGoogleUserInfoResponseDto,
	userAgent string,
) (*capi.LoginViaGoogleResponseDto, *cexceptions.Exception) {
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
			sinputs.PartialUpdateUserInput{
				Values: sinputs.UpdateUserInput{
					LoginCount:     &newLoginCount,
					BlockLoginUtil: blockLoginUntil,
				},
				SetNull: nil,
			},
			srepositories.WithTransactionDB(tx),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
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
		if exception := s.emailClient.SendSecurityAlertEmail(ctx, cemaildto.SendSecurityAlertEmailRequestDto{
			To:               user.Email,
			UserName:         user.Name,
			Status:           user.Status.String(),
			AlertType:        "Login in Different Place",
			Reason:           "Your account has a recent login action in other place",
			TimeOfOccurrence: time.Now(),
			OtherDetails:     "",
		}); exception != nil {
			_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
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
			_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
		}
	} else { // else if it does not exist
		// then we have to first get the relative data from different tables
		// we done this by one custom sql so it's not that slow...
		// once we have the required data, we set it as the user data cache
		output := struct {
			Id          uuid.UUID         `gorm:"id"`
			PublicId    uuid.UUID         `gorm:"public_id"`
			Name        string            `gorm:"name"`
			DisplayName string            `gorm:"display_name"`
			Email       string            `gorm:"email"`
			Role        cenums.UserRole   `gorm:"role"`
			Plan        cenums.UserPlan   `gorm:"plan"`
			Status      cenums.UserStatus `gorm:"status"`
			AvatarURL   *string           `gorm:"avatar_url"`
			CreatedAt   time.Time         `gorm:"created_at"`
			UpdatedAt   time.Time         `gorm:"updated_at"`
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
		sinputs.PartialUpdateUserInput{
			Values: sinputs.UpdateUserInput{
				Status:       &user.PrevStatus,
				RefreshToken: newRefreshToken,
				UserAgent:    &userAgent,
				LoginCount:   &zeroLoginCount,
			},
			SetNull: nil,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.LoginViaGoogleResponseDto{
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

func (s *AuthService) getLoginBlockedUntilByLoginCount(loginCount int32) (*time.Time, *cexceptions.Exception) {
	if loginCount < 0 {
		return nil, cexceptions.New("InvalidLoginCount", "Auth", "GetLoginBlockedUntil", "Login count is invalid", http.StatusInternalServerError, true)
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

func (s *AuthService) hashPassword(password string) (string, *cexceptions.Exception) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", cexceptions.New(
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

func (s *AuthService) generateAccessToken(userPublicId uuid.UUID, name string, email string, userAgent string) (*string, *cexceptions.Exception) {
	token, err := sharedtokens.GenerateAccessToken(
		userPublicId.String(),
		sharedtokens.AccessTokenClaims{
			Name:      name,
			Email:     email,
			UserAgent: userAgent,
		},
	)
	if err != nil {
		return nil, cexceptions.New(
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

func (s *AuthService) generateRefreshToken(userPublicId uuid.UUID, name string, email string, userAgent string) (*string, *cexceptions.Exception) {
	token, err := sharedtokens.GenerateRefreshToken(
		userPublicId.String(),
		sharedtokens.RefreshTokenClaims{
			Name:      name,
			Email:     email,
			UserAgent: userAgent,
		},
	)
	if err != nil {
		return nil, cexceptions.New(
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

func (s *AuthService) generateCSRFToken() (*string, *cexceptions.Exception) {
	token, err := sharedtokens.GenerateCSRFToken(sharedtokens.CSRFTokenClaims{})
	if err != nil {
		return nil, cexceptions.New(
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
) *cexceptions.Exception {
	payload, err := json.Marshal(cnotificationtypes.NewsPayload{
		Title:   "Welcome to Notegic",
		Summary: "Your Notegic account is ready.",
		Body:    "Start organizing your notes, shelves, and routines in one place.",
	})
	if err != nil {
		return cexceptions.New(
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
		coreevents.NotificationRequestedData{
			RecipientUserPublicId: userPublicId,
			Type:                  coreevents.NotificationType_News,
			Priority:              coreevents.NotificationPriority_Normal,
			TemplateKey:           cnotificationtypes.TemplateKey_News,
			TemplateVersion:       1,
			Payload:               payload,
			DedupeKey:             "welcome:" + userPublicId.String(),
		},
	); err != nil {
		return cexceptions.New(
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
	ctx context.Context, reqDto *capi.RegisterRequestDto,
) (*capi.RegisterResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewAuthException().InvalidDto().WithOrigin(err)
	}

	// put the hash part outside the transaction to decrease its blocking time from heavily hashing operation
	hashedPassword, exception := s.hashPassword(reqDto.Body.Password)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	createUserInput := sinputs.CreateUserInput{
		Name:        reqDto.Body.Name,
		DisplayName: s.generateRandomFakeDisplayName(), // we generate a default display name for the new user
		Email:       reqDto.Body.Email,
		Password:    hashedPassword,
		UserAgent:   reqDto.Header.UserAgent,
	}
	newUserId, exception := s.userRepository.CreateOne(
		createUserInput,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	createdUser, exception := s.userRepository.GetOneById(
		*newUserId,
		nil,
		srepositories.WithTransactionDB(tx),
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
		sinputs.PartialUpdateUserInput{
			Values: sinputs.UpdateUserInput{
				RefreshToken: newRefreshToken,
			},
			SetNull: nil,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userInfoRepository.CreateOneByUserId(
		*newUserId,
		sinputs.CreateUserInfoInput{},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userAccountRepository.CreateOneByUserId(
		*newUserId,
		sinputs.CreateUserAccountInput{
			AuthCode:          authCode,
			AuthCodeExpiredAt: authCodeExpiredAt,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userSettingRepository.CreateOneByUserId(
		*newUserId,
		sinputs.CreateUserSettingInput{},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
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
		_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
	}

	if exception = s.emailClient.SendWelcomeEmail(ctx, cemaildto.SendWelcomeEmailRequestDto{
		To:       newUser.Email,
		UserName: newUser.Name,
		Status:   newUser.Status.String(),
	}); exception != nil {
		_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
	}

	return &capi.RegisterResponseDto{
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
	ctx context.Context, reqDto *capi.RegisterViaGoogleRequestDto,
) (*capi.RegisterViaGoogleResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewAuthException().InvalidDto().WithOrigin(err)
	}

	userInfo, exception := s.oauthService.GetGoogleUserInfo(
		ctx,
		&capi.GetGoogleUserInfoRequestDto{
			AuthenticationCode: reqDto.Body.AuthorizationCode,
		},
	)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	existingUser, lookupException := s.userRepository.GetOneByEmail(
		userInfo.Email,
		[]sschemas.UserRelation{sschemas.UserRelation_UserAccount},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if lookupException == nil && existingUser != nil {
		loginResponse, loginException := s.loginByGoogleUserInfo(
			ctx,
			tx,
			existingUser,
			userInfo,
			reqDto.Header.UserAgent,
		)
		if loginException != nil {
			return nil, loginException
		}

		return &capi.RegisterViaGoogleResponseDto{
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
		return nil, cexceptions.New(
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
	if len(fakeName) > sconstants.MaxNameLength {
		fakeName = fakeName[:sconstants.MaxNameLength]
	}

	createUserInput := sinputs.CreateUserInput{
		Name:        fakeName,
		DisplayName: s.generateRandomFakeDisplayName(), // we generate a default display name for the new user
		Email:       userInfo.Email,
		Password:    hashedPassword,
		UserAgent:   reqDto.Header.UserAgent,
	}
	newUserId, exception := s.userRepository.CreateOne(
		createUserInput,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	createdUser, exception := s.userRepository.GetOneById(
		*newUserId,
		nil,
		srepositories.WithTransactionDB(tx),
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
		sinputs.PartialUpdateUserInput{
			Values: sinputs.UpdateUserInput{
				RefreshToken: newRefreshToken,
			},
			SetNull: nil,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userInfoRepository.CreateOneByUserId(
		*newUserId,
		sinputs.CreateUserInfoInput{},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userAccountRepository.CreateOneByUserId(
		*newUserId,
		sinputs.CreateUserAccountInput{
			AuthCode:          authCode,
			AuthCodeExpiredAt: authCodeExpiredAt,
			GoogleCredential:  &userInfo.Id,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	_, exception = s.userSettingRepository.CreateOneByUserId(
		*newUserId,
		sinputs.CreateUserSettingInput{},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
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
		_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
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
	if exception = s.emailClient.SendWelcomeEmail(ctx, cemaildto.SendWelcomeEmailRequestDto{
		To:       newUser.Email,
		UserName: newUser.Name,
		Status:   newUser.Status.String(),
	}); exception != nil {
		_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
	}

	return &capi.RegisterViaGoogleResponseDto{
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
	ctx context.Context, reqDto *capi.LoginRequestDto,
) (*capi.LoginResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewUserException().InvalidInput().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()

	// otherwise, the user should provide their account and password
	var user *sschemas.User = nil
	var exception *cexceptions.Exception = nil
	if sstrings.IsAlphaAndNumberString(reqDto.Body.Account) { // if the account field contains user name
		if user, exception = s.userRepository.GetOneByName(
			reqDto.Body.Account,
			nil,
			srepositories.WithTransactionDB(tx),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		); exception != nil {
			tx.Rollback()
			return nil, exception
		}
	} else if sstrings.IsEmailString(reqDto.Body.Account) { // if the account field contains email
		if user, exception = s.userRepository.GetOneByEmail(
			reqDto.Body.Account,
			nil,
			srepositories.WithTransactionDB(tx),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
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
			sinputs.PartialUpdateUserInput{
				Values: sinputs.UpdateUserInput{
					LoginCount:     &newLoginCount,
					BlockLoginUtil: blockLoginUntil,
				},
				SetNull: nil,
			},
			srepositories.WithTransactionDB(tx),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
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
		if exception := s.emailClient.SendSecurityAlertEmail(ctx, cemaildto.SendSecurityAlertEmailRequestDto{
			To:               user.Email,
			UserName:         user.Name,
			Status:           user.Status.String(),
			AlertType:        "Login in Different Place",
			Reason:           "Your account has a recent login action in other place",
			TimeOfOccurrence: time.Now(),
			OtherDetails:     "",
		}); exception != nil {
			_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
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
			_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
		}
	} else { // else if it does not exist
		// then we have to first get the relative data from different tables
		// we done this by one custom sql so it's not that slow...
		// once we have the required data, we set it as the user data cache
		output := struct {
			Id          uuid.UUID         `gorm:"id"`
			PublicId    uuid.UUID         `gorm:"public_id"`
			Name        string            `gorm:"name"`
			DisplayName string            `gorm:"display_name"`
			Email       string            `gorm:"email"`
			Role        cenums.UserRole   `gorm:"role"`
			Plan        cenums.UserPlan   `gorm:"plan"`
			Status      cenums.UserStatus `gorm:"status"`
			AvatarURL   *string           `gorm:"avatar_url"`
			CreatedAt   time.Time         `gorm:"created_at"`
			UpdatedAt   time.Time         `gorm:"updated_at"`
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
		sinputs.PartialUpdateUserInput{
			Values: sinputs.UpdateUserInput{
				Status:       &user.PrevStatus,
				RefreshToken: newRefreshToken,
				UserAgent:    &reqDto.Header.UserAgent,
				LoginCount:   &zeroLoginCount,
			},
			SetNull: nil,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.LoginResponseDto{
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
	ctx context.Context, reqDto *capi.LoginViaGoogleRequestDto,
) (*capi.LoginViaGoogleResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewAuthException().InvalidDto().WithOrigin(err)
	}

	userInfo, exception := s.oauthService.GetGoogleUserInfo(
		ctx,
		&capi.GetGoogleUserInfoRequestDto{
			AuthenticationCode: reqDto.Body.AuthorizationCode,
		},
	)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	user, exception := s.userRepository.GetOneByEmail(
		userInfo.Email,
		[]sschemas.UserRelation{sschemas.UserRelation_UserAccount},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if user == nil {
		tx.Rollback()
		return nil, apiexceptions.NewAuthException().InvalidDto()
	}

	return s.loginByGoogleUserInfo(ctx, tx, user, userInfo, reqDto.Header.UserAgent)
}

func (s *AuthService) Logout(
	ctx context.Context, reqDto *capi.LogoutRequestDto,
) (*capi.LogoutResponseDto, *cexceptions.Exception) {
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
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"Auth",
			"Logout",
			"Failed to begin the logout transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	offlineStatus := cenums.UserStatus_Offline
	emptyString := ""
	updatedUser, exception := s.userRepository.UpdateOneById(
		actorUserId,
		sinputs.PartialUpdateUserInput{
			Values: sinputs.UpdateUserInput{
				Status:       &offlineStatus,
				RefreshToken: &emptyString,
			},
			SetNull: nil,
		},
		srepositories.WithTransactionDB(tx),
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
		return nil, cexceptions.New(
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

	return &capi.LogoutResponseDto{
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}

func (s *AuthService) SendAuthCode(
	ctx context.Context, reqDto *capi.SendAuthCodeRequestDto,
) (*capi.SendAuthCodeResponseDto, *cexceptions.Exception) {
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

	if exception := s.emailClient.SendValidationEmail(ctx, cemaildto.SendValidationEmailRequestDto{
		To:        reqDto.Body.Email,
		UserName:  output.Name,
		AuthCode:  authCode,
		UserAgent: output.UserAgent,
		ExpiredAt: authCodeExpiredAt,
	}); exception != nil {
		return nil, exception
	}

	return &capi.SendAuthCodeResponseDto{
		AuthCodeExpiredAt:  authCodeExpiredAt,
		BlockAuthCodeUntil: blockAuthCodeUntil,
		UpdatedAt:          time.Now(),
	}, nil
}

func (s *AuthService) ValidateEmail(
	ctx context.Context, reqDto *capi.ValidateEmailRequestDto,
) (*capi.ValidateEmailResponseDto, *cexceptions.Exception) {
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

	return &capi.ValidateEmailResponseDto{
		UpdatedAt: updatedAt,
	}, nil
}

func (s *AuthService) ResetEmail(
	ctx context.Context, reqDto *capi.ResetEmailRequestDto,
) (*capi.ResetEmailResponseDto, *cexceptions.Exception) {
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
		sinputs.PartialUpdateUserAccountInput{
			Values: sinputs.UpdateUserAccountInput{
				AuthCode:          &authCode,
				AuthCodeExpiredAt: &authCodeExpiredAt,
			},
			SetNull: nil,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.ResetEmailResponseDto{
		UpdatedAt: updatedAt,
	}, nil
}

func (s *AuthService) ForgetPassword(
	ctx context.Context, reqDto *capi.ForgetPasswordRequestDto,
) (*capi.ForgetPasswordResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(reqDto); err != nil {
		return nil, apiexceptions.NewUserException().InvalidInput().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()

	var user *sschemas.User = nil
	var exception *cexceptions.Exception = nil
	var preloads = []sschemas.UserRelation{sschemas.UserRelation_UserAccount, sschemas.UserRelation_UserInfo, sschemas.UserRelation_UserSetting}
	if sstrings.IsEmailString(reqDto.Body.Account) { // if the account field contains email
		if user, exception = s.userRepository.GetOneByEmail(
			reqDto.Body.Account,
			preloads,
			srepositories.WithTransactionDB(tx),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		); exception != nil {
			tx.Rollback()
			return nil, exception
		}
	} else if sstrings.IsAlphaAndNumberString(reqDto.Body.Account) { // if the account field contains user name
		if user, exception = s.userRepository.GetOneByName(
			reqDto.Body.Account,
			preloads,
			srepositories.WithTransactionDB(tx),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
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
		_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
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
			_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
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
		sinputs.PartialUpdateUserInput{
			Values: sinputs.UpdateUserInput{
				Password:     &hashedPassword,
				RefreshToken: newRefreshToken,
				UserAgent:    &reqDto.Header.UserAgent,
				LoginCount:   &zeroLoginCount,
			},
			SetNull: nil,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueUserSessionsRevoked(
		tx,
		user.PublicId.String(),
		user.PublicId,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
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

	return &capi.ForgetPasswordResponseDto{
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}

func (s *AuthService) ResetMe(
	ctx context.Context, reqDto *capi.ResetMeRequestDto,
) (*capi.ResetMeResponseDto, *cexceptions.Exception) {
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
	var resetUserAccount sschemas.UserAccount
	result := tx.Model(&resetUserAccount).
		Where("user_id = ? AND auth_code = ?", actorUserId, reqDto.Body.AuthCode).
		First(&resetUserAccount)
	if err := result.Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserAccountException().NotFound().WithOrigin(err)
	}

	// delete the user info
	if err := tx.Where("user_id = ?", actorUserId).Delete(&sschemas.UserInfo{}).Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserInfoException().FailedToDelete().WithOrigin(err)
	}
	// and then re-create a new user info
	if _, exception := s.userInfoRepository.CreateOneByUserId(
		resetUserAccount.UserId,
		sinputs.CreateUserInfoInput{},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}

	// delete the user setting
	if err := tx.Where("user_id = ?", actorUserId).Delete(&sschemas.UserSetting{}).Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserSettingException().FailedToDelete().WithOrigin(err)
	}
	// and then re-create a new user setting
	if _, exception := s.userSettingRepository.CreateOneByUserId(
		resetUserAccount.UserId,
		sinputs.CreateUserSettingInput{},
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
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
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	); exception != nil {
		// skip if there's no root shelves to soft delete
	} else {
		// then hard delete all the root shelves of the user
		if exception := s.rootShelfRepository.HardDeleteManyByUserId(
			actorUserId,
			srepositories.WithTransactionDB(tx),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		); exception != nil {
			// skip if there's no root shelves to hard delete
		}
	}

	// delete other stuff in the future...

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewUserException().FailedToCommitTransaction().WithDetails(err)
	}

	return &capi.ResetMeResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *AuthService) DeleteMe(
	ctx context.Context, reqDto *capi.DeleteMeRequestDto,
) (*capi.DeleteMeResponseDto, *cexceptions.Exception) {
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
		return nil, cexceptions.New(
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
		return nil, cexceptions.New(
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
		_ = slogs.NotegicLogger.JSON(ctx, slog.LevelError, exception.String(), exception)
	}

	return &capi.DeleteMeResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *AuthService) RegisterViaMeta() {}

func (s *AuthService) RegisterViaGithub() {}

func (s *AuthService) LoginViaMeta() {}

func (s *AuthService) LoginViaGithub() {}
