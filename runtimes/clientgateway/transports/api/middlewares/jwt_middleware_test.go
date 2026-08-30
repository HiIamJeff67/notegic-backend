package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"
)

func newTestCookieHandlers() (*scookies.CookieHandler, *scookies.CookieHandler) {
	return scookies.New(scookies.Config{
			Name:     scookies.ValidCookieName_AccessToken,
			Path:     "/",
			Duration: 30 * time.Minute,
			HTTPOnly: true,
			SameSite: http.SameSiteLaxMode,
		}), scookies.New(scookies.Config{
			Name:     scookies.ValidCookieName_RefreshToken,
			Path:     "/",
			Duration: 14 * 24 * time.Hour,
			HTTPOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
}

func TestJWTMiddlewareExtractsAccessTokenSubjectWithoutDataStore(t *testing.T) {
	t.Setenv("JWT_ACCESS_TOKEN_SECRET_KEY", "test-access-secret")
	accessToken, err := sharedtokens.GenerateAccessToken(
		"83bdeac1-02de-42fe-a7a8-4e1a83174866",
		sharedtokens.AccessTokenClaims{
			Name:      "notegic",
			Email:     "notegic@example.com",
			UserAgent: "test-agent",
		},
	)
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	accessTokenCookieHandler, refreshTokenCookieHandler := newTestCookieHandlers()
	router.Use(JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler))
	router.GET("/", func(ctx *gin.Context) {
		value, exists := ctx.Get(sharedcontexts.ContextFieldName_User_PublicId.String())
		if !exists || value != "83bdeac1-02de-42fe-a7a8-4e1a83174866" {
			t.Fatalf("unexpected user public ID context: %#v", value)
		}
		ctx.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("User-Agent", "test-agent")
	request.AddCookie(&http.Cookie{Name: "accessToken", Value: *accessToken})
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}

func TestJWTMiddlewareFallsBackToRefreshTokenSubject(t *testing.T) {
	t.Setenv("JWT_REFRESH_TOKEN_SECRET_KEY", "test-refresh-secret")
	refreshToken, err := sharedtokens.GenerateRefreshToken(
		"83bdeac1-02de-42fe-a7a8-4e1a83174866",
		sharedtokens.RefreshTokenClaims{
			Name:      "notegic",
			Email:     "notegic@example.com",
			UserAgent: "test-agent",
		},
	)
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	accessTokenCookieHandler, refreshTokenCookieHandler := newTestCookieHandlers()
	router.Use(JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler))
	router.GET("/", func(ctx *gin.Context) {
		value, exists := ctx.Get(sharedcontexts.ContextFieldName_User_PublicId.String())
		if !exists || value != "83bdeac1-02de-42fe-a7a8-4e1a83174866" {
			t.Fatalf("unexpected user public ID context: %#v", value)
		}
		ctx.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("User-Agent", "test-agent")
	request.AddCookie(&http.Cookie{Name: "refreshToken", Value: *refreshToken})
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}
