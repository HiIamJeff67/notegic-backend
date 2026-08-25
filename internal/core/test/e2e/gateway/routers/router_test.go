package routers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	crootshelves "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/root-shelves"
	cstations "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/stations"
	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	corerouters "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/routers"
)

func TestRouterValidatesRootShelfEnvelopeBeforeCallingService(t *testing.T) {
	t.Setenv("CORE_DELEGATION_AUDIENCE", "notegic-api-test")
	t.Setenv("CORE_DELEGATION_ISSUER", "notegic-gateway-test")
	t.Setenv("CORE_DELEGATION_SECRET", "test-delegation-secret")
	t.Setenv("JWT_ACCESS_TOKEN_SECRET_KEY", "test-access-secret")

	tokenValue, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:              "gateway",
		UserSubject:        "83bdeac1-02de-42fe-a7a8-4e1a83174866",
		AllowedPermissions: []string{"Read"},
		Operation:          crootshelves.GetMyRootShelfByIdOperation,
		RequestId:          "request-id",
	})
	if err != nil {
		t.Fatalf("issue delegation token: %v", err)
	}
	tokenString := *tokenValue

	requestBody := []byte(`{"version":"v1","operation":"root-shelf.get-my-root-shelf-by-id","metadata":{"requestId":"request-id"},"dto":"invalid"}`)

	router := newRouterForDelegationValidation()
	request := httptest.NewRequest(
		http.MethodPost,
		"/core/v1/root-shelves/get-by-id",
		bytes.NewReader(requestBody),
	)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	accessToken, err := sharedtokens.GenerateAccessToken(
		"83bdeac1-02de-42fe-a7a8-4e1a83174866",
		sharedtokens.AccessTokenClaims{
			UserAgent: "test-agent",
		},
	)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}
	request.AddCookie(&http.Cookie{Name: "accessToken", Value: *accessToken})
	request.Header.Set("User-Agent", "test-agent")
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

func TestRouterValidatesStationEnvelopeBeforeCallingService(t *testing.T) {
	t.Setenv("CORE_DELEGATION_AUDIENCE", "notegic-api-test")
	t.Setenv("CORE_DELEGATION_ISSUER", "notegic-gateway-test")
	t.Setenv("CORE_DELEGATION_SECRET", "test-delegation-secret")
	t.Setenv("JWT_ACCESS_TOKEN_SECRET_KEY", "test-access-secret")

	tokenValue, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:              "gateway",
		UserSubject:        "83bdeac1-02de-42fe-a7a8-4e1a83174866",
		AllowedPermissions: []string{"Read"},
		Operation:          cstations.GetMyStationByIdOperation,
		RequestId:          "request-id",
	})
	if err != nil {
		t.Fatalf("issue delegation token: %v", err)
	}
	tokenString := *tokenValue

	requestBody := []byte(`{"version":"v1","operation":"station.get-my-station-by-id","metadata":{"requestId":"request-id"},"dto":"invalid"}`)

	router := newRouterForDelegationValidation()
	request := httptest.NewRequest(
		http.MethodPost,
		"/core/v1/stations/get-by-id",
		bytes.NewReader(requestBody),
	)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	accessToken, err := sharedtokens.GenerateAccessToken(
		"83bdeac1-02de-42fe-a7a8-4e1a83174866",
		sharedtokens.AccessTokenClaims{
			UserAgent: "test-agent",
		},
	)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}
	request.AddCookie(&http.Cookie{Name: "accessToken", Value: *accessToken})
	request.Header.Set("User-Agent", "test-agent")
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, responseRecorder.Code)
	}
}

func TestRouterRejectsDelegationForAnotherOperation(t *testing.T) {
	t.Setenv("CORE_DELEGATION_AUDIENCE", "notegic-api-test")
	t.Setenv("CORE_DELEGATION_ISSUER", "notegic-gateway-test")
	t.Setenv("CORE_DELEGATION_SECRET", "test-delegation-secret")
	t.Setenv("JWT_ACCESS_TOKEN_SECRET_KEY", "test-access-secret")

	tokenValue, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:       "gateway",
		UserSubject: "83bdeac1-02de-42fe-a7a8-4e1a83174866",
		Operation:   "root-shelf.create",
		RequestId:   "request-id",
	})
	if err != nil {
		t.Fatalf("issue delegation token: %v", err)
	}
	tokenString := *tokenValue
	requestBody, err := json.Marshal(cgateway.Request[struct{}]{
		Version:   cgateway.Version,
		Operation: crootshelves.GetMyRootShelfByIdOperation,
		Metadata: cgateway.RequestMetadata{
			RequestId: "request-id",
		},
		Dto: struct{}{},
	})
	if err != nil {
		t.Fatalf("marshal Core service request: %v", err)
	}

	router := newRouterForDelegationValidation()
	request := httptest.NewRequest(
		http.MethodPost,
		"/core/v1/root-shelves/get-by-id",
		bytes.NewReader(requestBody),
	)
	request.Header.Set("Authorization", "Bearer "+tokenString)
	accessToken, err := sharedtokens.GenerateAccessToken(
		"83bdeac1-02de-42fe-a7a8-4e1a83174866",
		sharedtokens.AccessTokenClaims{
			UserAgent: "test-agent",
		},
	)
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}
	request.AddCookie(&http.Cookie{Name: "accessToken", Value: *accessToken})
	request.Header.Set("User-Agent", "test-agent")
	request.Header.Set("Content-Type", "application/json")
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)
	if responseRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, responseRecorder.Code)
	}
}

func newRouterForDelegationValidation() *gin.Engine {
	return corerouters.NewRouter(corerouters.RouterDependencies{
		Auth: corerouters.AuthRouterDependencies{
			AuthMiddleware: func(ctx *gin.Context) {
				ctx.Next()
			},
		},
		RootShelf: corerouters.RootShelfRouterDependencies{
			AuthMiddleware: func(ctx *gin.Context) {
				ctx.Next()
			},
		},
		Station: corerouters.StationRouterDependencies{
			AuthMiddleware: func(ctx *gin.Context) {
				ctx.Next()
			},
		},
	})
}
