package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/auth"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type AuthControllerInterface interface {
	Register(ctx *gin.Context, requestDto *capi.RegisterRequestDto)
	RegisterViaGoogle(ctx *gin.Context, requestDto *capi.RegisterViaGoogleRequestDto)
	Login(ctx *gin.Context, requestDto *capi.LoginRequestDto)
	LoginViaGoogle(ctx *gin.Context, requestDto *capi.LoginViaGoogleRequestDto)
	Logout(ctx *gin.Context, requestDto *capi.LogoutRequestDto)
	SendAuthCode(ctx *gin.Context, requestDto *capi.SendAuthCodeRequestDto)
	ValidateEmail(ctx *gin.Context, requestDto *capi.ValidateEmailRequestDto)
	ResetEmail(ctx *gin.Context, requestDto *capi.ResetEmailRequestDto)
	ForgetPassword(ctx *gin.Context, requestDto *capi.ForgetPasswordRequestDto)
	ResetMe(ctx *gin.Context, requestDto *capi.ResetMeRequestDto)
	DeleteMe(ctx *gin.Context, requestDto *capi.DeleteMeRequestDto)
}

type AuthController struct {
	coreAdapter               *coreadapters.CoreAdapter
	accessTokenCookieHandler  *scookies.CookieHandler
	refreshTokenCookieHandler *scookies.CookieHandler
}

func NewAuthController(
	coreAdapter *coreadapters.CoreAdapter,
	accessTokenCookieHandler *scookies.CookieHandler,
	refreshTokenCookieHandler *scookies.CookieHandler,
) AuthControllerInterface {
	return &AuthController{
		coreAdapter:               coreAdapter,
		accessTokenCookieHandler:  accessTokenCookieHandler,
		refreshTokenCookieHandler: refreshTokenCookieHandler,
	}
}

func (c *AuthController) Register(ctx *gin.Context, requestDto *capi.RegisterRequestDto) {
	c.accessTokenCookieHandler.Delete(ctx)
	c.refreshTokenCookieHandler.Delete(ctx)

	response, exception := coreadapters.Call[capi.RegisterRequestDto, capi.RegisterResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RegisterOperation,
		"/core/v1/auth/register",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	c.accessTokenCookieHandler.Set(ctx, response.Data.AccessToken)
	c.refreshTokenCookieHandler.Set(ctx, response.Data.RefreshToken)
	writeClientResponse(ctx, struct {
		PublicId    uuid.UUID `json:"publicId"`
		Name        string    `json:"name"`
		DisplayName string    `json:"displayName"`
		Email       string    `json:"email"`
		CSRFToken   string    `json:"csrfToken"`
		CreatedAt   time.Time `json:"createdAt"`
	}{
		PublicId:    response.Data.PublicId,
		Name:        response.Data.Name,
		DisplayName: response.Data.DisplayName,
		Email:       response.Data.Email,
		CSRFToken:   response.Data.CSRFToken,
		CreatedAt:   response.Data.CreatedAt,
	})
}

func (c *AuthController) RegisterViaGoogle(ctx *gin.Context, requestDto *capi.RegisterViaGoogleRequestDto) {
	c.accessTokenCookieHandler.Delete(ctx)
	c.refreshTokenCookieHandler.Delete(ctx)

	response, exception := coreadapters.Call[capi.RegisterViaGoogleRequestDto, capi.RegisterViaGoogleResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.RegisterViaGoogleOperation,
		"/core/v1/auth/register/google",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	c.accessTokenCookieHandler.Set(ctx, response.Data.AccessToken)
	c.refreshTokenCookieHandler.Set(ctx, response.Data.RefreshToken)
	writeClientResponse(ctx, struct {
		PublicId    uuid.UUID `json:"publicId"`
		Name        string    `json:"name"`
		DisplayName string    `json:"displayName"`
		Email       string    `json:"email"`
		CSRFToken   string    `json:"csrfToken"`
		CreatedAt   time.Time `json:"createdAt"`
	}{
		PublicId:    response.Data.PublicId,
		Name:        response.Data.Name,
		DisplayName: response.Data.DisplayName,
		Email:       response.Data.Email,
		CSRFToken:   response.Data.CSRFToken,
		CreatedAt:   response.Data.CreatedAt,
	})
}

func (c *AuthController) Login(ctx *gin.Context, requestDto *capi.LoginRequestDto) {
	c.accessTokenCookieHandler.Delete(ctx)
	c.refreshTokenCookieHandler.Delete(ctx)

	response, exception := coreadapters.Call[capi.LoginRequestDto, capi.LoginResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.LoginOperation,
		"/core/v1/auth/login",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	c.accessTokenCookieHandler.Set(ctx, response.Data.AccessToken)
	c.refreshTokenCookieHandler.Set(ctx, response.Data.RefreshToken)
	updatedAt := response.Data.UpdatedAt
	writeClientResponse(ctx, struct {
		PublicId    uuid.UUID  `json:"publicId"`
		Name        string     `json:"name"`
		DisplayName string     `json:"displayName"`
		Email       string     `json:"email"`
		CSRFToken   string     `json:"csrfToken"`
		UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
		CreatedAt   time.Time  `json:"createdAt"`
	}{
		PublicId:    response.Data.PublicId,
		Name:        response.Data.Name,
		DisplayName: response.Data.DisplayName,
		Email:       response.Data.Email,
		CSRFToken:   response.Data.CSRFToken,
		UpdatedAt:   &updatedAt,
		CreatedAt:   response.Data.CreatedAt,
	})
}

func (c *AuthController) LoginViaGoogle(ctx *gin.Context, requestDto *capi.LoginViaGoogleRequestDto) {
	c.accessTokenCookieHandler.Delete(ctx)
	c.refreshTokenCookieHandler.Delete(ctx)

	response, exception := coreadapters.Call[capi.LoginViaGoogleRequestDto, capi.LoginViaGoogleResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.LoginViaGoogleOperation,
		"/core/v1/auth/login/google",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	c.accessTokenCookieHandler.Set(ctx, response.Data.AccessToken)
	c.refreshTokenCookieHandler.Set(ctx, response.Data.RefreshToken)
	updatedAt := response.Data.UpdatedAt
	writeClientResponse(ctx, struct {
		PublicId    uuid.UUID  `json:"publicId"`
		Name        string     `json:"name"`
		DisplayName string     `json:"displayName"`
		Email       string     `json:"email"`
		CSRFToken   string     `json:"csrfToken"`
		UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
		CreatedAt   time.Time  `json:"createdAt"`
	}{
		PublicId:    response.Data.PublicId,
		Name:        response.Data.Name,
		DisplayName: response.Data.DisplayName,
		Email:       response.Data.Email,
		CSRFToken:   response.Data.CSRFToken,
		UpdatedAt:   &updatedAt,
		CreatedAt:   response.Data.CreatedAt,
	})
}

func (c *AuthController) Logout(ctx *gin.Context, requestDto *capi.LogoutRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.LogoutRequestDto, capi.LogoutResponseDto](
		ctx,
		c.coreAdapter,
		requestDto,
		capi.LogoutOperation,
		"/core/v1/auth/logout",
	)
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	c.accessTokenCookieHandler.Delete(ctx)
	c.refreshTokenCookieHandler.Delete(ctx)
	writeClientResponse(ctx, response.Data)
}

func (c *AuthController) SendAuthCode(ctx *gin.Context, requestDto *capi.SendAuthCodeRequestDto) {
	response, exception := coreadapters.Call[capi.SendAuthCodeRequestDto, capi.SendAuthCodeResponseDto](ctx, c.coreAdapter, requestDto, capi.SendAuthCodeOperation, "/core/v1/auth/email/code")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *AuthController) ValidateEmail(ctx *gin.Context, requestDto *capi.ValidateEmailRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.ValidateEmailRequestDto, capi.ValidateEmailResponseDto](ctx, c.coreAdapter, requestDto, capi.ValidateEmailOperation, "/core/v1/auth/email/validate")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *AuthController) ResetEmail(ctx *gin.Context, requestDto *capi.ResetEmailRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.ResetEmailRequestDto, capi.ResetEmailResponseDto](ctx, c.coreAdapter, requestDto, capi.ResetEmailOperation, "/core/v1/auth/email/reset")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *AuthController) ForgetPassword(ctx *gin.Context, requestDto *capi.ForgetPasswordRequestDto) {
	response, exception := coreadapters.Call[capi.ForgetPasswordRequestDto, capi.ForgetPasswordResponseDto](ctx, c.coreAdapter, requestDto, capi.ForgetPasswordOperation, "/core/v1/auth/password/forget")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *AuthController) ResetMe(ctx *gin.Context, requestDto *capi.ResetMeRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.ResetMeRequestDto, capi.ResetMeResponseDto](ctx, c.coreAdapter, requestDto, capi.ResetMeOperation, "/core/v1/auth/me/reset")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}

func (c *AuthController) DeleteMe(ctx *gin.Context, requestDto *capi.DeleteMeRequestDto) {
	response, exception := coreadapters.CallSecurly[capi.DeleteMeRequestDto, capi.DeleteMeResponseDto](ctx, c.coreAdapter, requestDto, capi.DeleteMeOperation, "/core/v1/auth/me/delete")
	if exception != nil {
		sexceptionwriter.SafelyAbortAndResponseWithJSON(exception, ctx)
		return
	}

	writeClientResponse(ctx, response.Data)
}
