package middlewares_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	coremiddlewares "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/middlewares"
)

func TestAuthenticationMiddlewareValidatesForwardedAccessToken(t *testing.T) {
	t.Setenv("CORE_DELEGATION_AUDIENCE", "notegic-core-test")
	t.Setenv("CORE_DELEGATION_ISSUER", "notegic-gateway-test")
	t.Setenv("CORE_DELEGATION_SECRET", "test-delegation-secret")
	t.Setenv("JWT_ACCESS_TOKEN_SECRET_KEY", "test-access-secret")

	const userSubject = "83bdeac1-02de-42fe-a7a8-4e1a83174866"
	delegationTokenValue, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:       "gateway",
		UserSubject: userSubject,
		Operation:   "station.get",
		RequestId:   "request-id",
	})
	if err != nil {
		t.Fatalf("issue delegation token: %v", err)
	}
	delegationToken := *delegationTokenValue
	accessToken, err := sharedtokens.GenerateAccessToken(
		userSubject,
		sharedtokens.AccessTokenClaims{UserAgent: "test-agent"},
	)
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(coremiddlewares.DelegationAuthenticatedMiddleware(""), coremiddlewares.AuthMiddleware(nil, nil))
	router.POST("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	payload, err := json.Marshal(cgateway.Request[struct{}]{
		Version:   cgateway.Version,
		Operation: "station.get",
		Metadata: cgateway.RequestMetadata{
			RequestId: "request-id",
		},
		Tokens: cgateway.Tokens{
			AccessToken: *accessToken,
		},
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	request.Header.Set("Authorization", "Bearer "+delegationToken)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "test-agent")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}
