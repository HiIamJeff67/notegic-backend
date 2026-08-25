package middlewares_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	coremiddlewares "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/middlewares"
)

func TestDelegationMiddlewareSetsRoutePermissions(t *testing.T) {
	t.Setenv("CORE_DELEGATION_AUDIENCE", "notegic-api-test")
	t.Setenv("CORE_DELEGATION_ISSUER", "notegic-gateway-test")
	t.Setenv("CORE_DELEGATION_SECRET", "test-delegation-secret")

	tokenValue, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:              "gateway",
		AllowedPermissions: []string{"Read", "Write"},
		Operation:          "station.update",
		RequestId:          "request-id",
	})
	if err != nil {
		t.Fatalf("issue delegation token: %v", err)
	}
	tokenString := *tokenValue

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(coremiddlewares.DelegationMiddleware(""))
	router.GET("/", func(ctx *gin.Context) {
		allowedPermissions, exception := contexts.GetAllowedPermissions(ctx.Request.Context())
		if exception != nil || len(allowedPermissions) != 2 {
			t.Fatal("expected delegated route permissions")
		}
		ctx.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	request.Header.Set("X-Request-Id", "request-id")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}

func TestDelegationAuthenticatedMiddlewareLeavesRoutePermissionsAbsentWhenNotDelegated(t *testing.T) {
	t.Setenv("CORE_DELEGATION_AUDIENCE", "notegic-api-test")
	t.Setenv("CORE_DELEGATION_ISSUER", "notegic-gateway-test")
	t.Setenv("CORE_DELEGATION_SECRET", "test-delegation-secret")

	tokenValue, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:       "gateway",
		UserSubject: uuid.NewString(),
		Operation:   "user.me",
		RequestId:   "request-id",
	})
	if err != nil {
		t.Fatalf("issue delegation token: %v", err)
	}
	tokenString := *tokenValue

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(coremiddlewares.DelegationAuthenticatedMiddleware("user.me"))
	router.GET("/", func(ctx *gin.Context) {
		_, exception := contexts.GetAllowedPermissions(ctx.Request.Context())
		if exception == nil {
			t.Fatal("expected route permissions to remain absent")
		}
		ctx.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	request.Header.Set("X-Request-Id", "request-id")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}

func TestDelegationMiddlewarePreservesBodyForEndpointBinding(t *testing.T) {
	t.Setenv("CORE_DELEGATION_AUDIENCE", "notegic-api-test")
	t.Setenv("CORE_DELEGATION_ISSUER", "notegic-gateway-test")
	t.Setenv("CORE_DELEGATION_SECRET", "test-delegation-secret")

	const operation = "station.update"
	tokenValue, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:     "gateway",
		Operation: operation,
		RequestId: "request-id",
	})
	if err != nil {
		t.Fatalf("issue delegation token: %v", err)
	}
	tokenString := *tokenValue

	payload, err := json.Marshal(cgateway.Request[map[string]string]{
		Version:   cgateway.Version,
		Operation: operation,
		Metadata: cgateway.RequestMetadata{
			RequestId: "request-id",
		},
		Dto: map[string]string{"name": "station"},
	})
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST(
		"/",
		coremiddlewares.DelegationMiddleware(operation),
		func(ctx *gin.Context) {
			request := &cgateway.Request[map[string]string]{}
			if err := ctx.ShouldBindBodyWithJSON(request); err != nil {
				t.Fatalf("bind request after delegation middleware: %v", err)
			}
			if request.Dto["name"] != "station" {
				t.Fatalf("expected endpoint DTO to survive middleware binding, got %#v", request.Dto)
			}
			ctx.Status(http.StatusNoContent)
		},
	)

	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(payload))
	request.Header.Set("Authorization", "Bearer "+tokenString)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-Id", "request-id")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, responseRecorder.Code)
	}
}

func TestDelegationAuthenticatedMiddlewareRequiresUserSubject(t *testing.T) {
	t.Setenv("CORE_DELEGATION_AUDIENCE", "notegic-api-test")
	t.Setenv("CORE_DELEGATION_ISSUER", "notegic-gateway-test")
	t.Setenv("CORE_DELEGATION_SECRET", "test-delegation-secret")

	tokenValue, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:     "gateway",
		Operation: "station.update",
		RequestId: "request-id",
	})
	if err != nil {
		t.Fatalf("issue delegation token: %v", err)
	}
	tokenString := *tokenValue

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(coremiddlewares.DelegationAuthenticatedMiddleware(""))
	router.GET("/", func(ctx *gin.Context) {
		ctx.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, responseRecorder.Code)
	}
}
