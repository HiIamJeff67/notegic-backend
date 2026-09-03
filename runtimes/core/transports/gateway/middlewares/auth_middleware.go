package middlewares

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	userdata "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/redis/userdata"
)

func AuthMiddleware(
	userRepository srepositories.UserRepositoryInterface,
	userDataCacheClient *userdata.UserDataCacheClient,
	db *gorm.DB,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userPublicId, exception := contexts.GetActorUserPublicId(ctx.Request.Context())
		if exception != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
					"InvalidDelegation",
					"Core",
					"AuthenticateRequest",
					"a user delegation subject is required",
					http.StatusUnauthorized,
				),
			})
			return
		}

		request := &cgateway.Request[json.RawMessage]{}
		if ctx.Request.ContentLength != 0 {
			_ = ctx.ShouldBindBodyWithJSON(request)
		}
		accessToken := request.Tokens.AccessToken
		refreshToken := request.Tokens.RefreshToken
		accessTokenExists := accessToken != ""
		refreshTokenExists := refreshToken != ""
		userAgent := ctx.GetHeader("User-Agent")
		if accessTokenExists {
			claims, err := sharedtokens.ParseAccessToken(accessToken)
			if err == nil && claims.Subject == userPublicId.String() && claims.UserAgent == userAgent {
				if setActorUserId(ctx, userRepository, userPublicId, db) {
					ctx.Next()
				}
				return
			}
		}

		if !refreshTokenExists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
					"Unauthorized",
					"Core",
					"AuthenticateRequest",
					"the forwarded authentication credentials are invalid",
					http.StatusUnauthorized,
				),
			})
			return
		}

		claims, err := sharedtokens.ParseRefreshToken(refreshToken)
		if err != nil || claims.Subject != userPublicId.String() || claims.UserAgent != userAgent {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
					"Unauthorized",
					"Core",
					"AuthenticateRequest",
					"the forwarded authentication credentials are invalid",
					http.StatusUnauthorized,
				),
			})
			return
		}

		if userRepository == nil {
			ctx.Next()
			return
		}

		user, exception := userRepository.GetOneByPublicId(
			userPublicId,
			nil,
			srepositories.WithDB(db),
		)
		if exception != nil || user == nil || user.RefreshToken != refreshToken || user.UserAgent != userAgent {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
					"Unauthorized",
					"Core",
					"AuthenticateRequest",
					"the forwarded refresh credential could not be authenticated",
					http.StatusUnauthorized,
				),
			})
			return
		}

		newAccessToken, err := sharedtokens.GenerateAccessToken(
			user.PublicId.String(),
			sharedtokens.AccessTokenClaims{
				Name:      user.Name,
				Email:     user.Email,
				UserAgent: user.UserAgent,
			},
		)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
					"GenerationFailed",
					"Core",
					"AuthenticateRequest",
					"failed to generate a new access token",
					http.StatusInternalServerError,
					true,
				).WithOrigin(err),
			})
			return
		}
		var newCSRFToken *string
		if userDataCacheClient == nil {
			newCSRFToken, err = sharedtokens.GenerateCSRFToken(sharedtokens.CSRFTokenClaims{})
		} else {
			userDataCache, exception := userDataCacheClient.Get(user.Name)
			if exception != nil {
				if exception.Reason != "NotFound" {
					err = exception
				} else {
					newCSRFToken, err = sharedtokens.GenerateCSRFToken(sharedtokens.CSRFTokenClaims{})
					if err == nil {
						err = userDataCacheClient.Set(user.Name, userdata.UserDataCache{
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
							CreatedAt:   user.CreatedAt,
							UpdatedAt:   user.UpdatedAt,
						})
					}
				}
			} else if request.Tokens.CSRFToken != "" && request.Tokens.CSRFToken != userDataCache.CSRFToken {
				newCSRFToken = &userDataCache.CSRFToken
			} else {
				newCSRFToken, err = sharedtokens.GenerateCSRFToken(sharedtokens.CSRFTokenClaims{})
				if err == nil {
					var currentCSRFToken string
					currentCSRFToken, _, exception = userDataCacheClient.RotateCSRFToken(
						user.Name,
						userDataCache.CSRFToken,
						*newCSRFToken,
					)
					if exception != nil {
						err = exception
					} else {
						newCSRFToken = &currentCSRFToken
					}
				}
			}
		}
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
					"RefreshFailed",
					"Core",
					"AuthenticateRequest",
					"failed to refresh the CSRF token",
					http.StatusInternalServerError,
					true,
				).WithOrigin(err),
			})
			return
		}

		ctx.Set(sharedcontexts.ContextFieldName_IsNewTokens.String(), true)
		ctx.Set(sharedcontexts.ContextFieldName_AccessToken.String(), *newAccessToken)
		ctx.Set(sharedcontexts.ContextFieldName_CSRFToken.String(), *newCSRFToken)
		if !setActorUserId(ctx, userRepository, userPublicId, db) {
			return
		}
		ctx.Next()
	}
}

func setActorUserId(
	ctx *gin.Context,
	userRepository srepositories.UserRepositoryInterface,
	userPublicId uuid.UUID,
	db *gorm.DB,
) bool {
	if userRepository == nil {
		return true
	}

	user, exception := userRepository.GetOneByPublicId(
		userPublicId,
		nil,
		srepositories.WithDB(db),
	)
	if exception != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   ctx.GetHeader("X-Request-Id"),
				RespondedAt: time.Now(),
			},
			Data: struct{}{},
			Exception: cexceptions.New(
				"Unauthorized",
				"Core",
				"AuthenticateRequest",
				"the delegated user subject could not be authenticated",
				http.StatusUnauthorized,
			),
		})
		return false
	}

	requestContext := contexts.WithActorUserId(ctx.Request.Context(), user.Id)
	ctx.Request = ctx.Request.WithContext(contexts.WithActorUserName(requestContext, user.Name))
	ctx.Set(sharedcontexts.ContextFieldName_User_Name.String(), user.Name)
	ctx.Set(sharedcontexts.ContextFieldName_User_Email.String(), user.Email)
	ctx.Set(sharedcontexts.ContextFieldName_User_Role.String(), user.Role)
	ctx.Set(sharedcontexts.ContextFieldName_User_Plan.String(), user.Plan)
	return true
}
