package routers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	crootshelves "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/root-shelves"
	cusers "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/users"

	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	corerouters "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/routers"
)

func TestEitherMiddlewareAllowsAPIOnlyForPublishedResourceDomains(t *testing.T) {
	configureDelegationTestEnvironment(t)
	apiToken := issueAPIDelegationToken(t, crootshelves.GetMyRootShelfByIdOperation)
	requestBody := requestBodyForOperation(crootshelves.GetMyRootShelfByIdOperation)

	router := newRouterForAuthSelection()
	request := httptest.NewRequest(http.MethodPost, "/core/v1/root-shelves/get-by-id", bytes.NewReader(requestBody))
	request.Header.Set("Authorization", "Bearer "+apiToken)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusTeapot {
		t.Fatalf("published resource selected the wrong middleware branch: got %d, want %d", response.Code, http.StatusTeapot)
	}

	apiToken = issueAPIDelegationToken(t, cusers.GetMeOperation)
	requestBody = requestBodyForOperation(cusers.GetMeOperation)
	request = httptest.NewRequest(http.MethodPost, "/core/v1/users/me", bytes.NewReader(requestBody))
	request.Header.Set("Authorization", "Bearer "+apiToken)
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUpgradeRequired {
		t.Fatalf("client-only resource selected API middleware: got %d, want %d", response.Code, http.StatusUpgradeRequired)
	}
}

func configureDelegationTestEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("CORE_DELEGATION_AUDIENCE", "notegic-api-test")
	t.Setenv("CORE_DELEGATION_ISSUER", "notegic-gateway-test")
	t.Setenv("CORE_DELEGATION_SECRET", "test-delegation-secret")
}

func issueAPIDelegationToken(t *testing.T, operation string) string {
	t.Helper()
	token, err := sharedtokens.GenerateDelegationToken(sharedtokens.DelegationTokenClaims{
		Actor:         "gateway",
		GatewaySource: sharedtokens.GatewaySourceAPI,
		AuthMethod:    sharedtokens.AuthMethodAPIKey,
		Operation:     operation,
		RequestId:     "request-id",
	})
	if err != nil {
		t.Fatalf("issue API delegation token: %v", err)
	}
	return *token
}

func requestBodyForOperation(operation string) []byte {
	return []byte(`{"version":"v1","operation":"` + operation + `","metadata":{"requestId":"request-id"},"dto":{}}`)
}

func newRouterForAuthSelection() *gin.Engine {
	return corerouters.NewRouter(corerouters.RouterDependencies{
		Auth: corerouters.AuthRouterDependencies{
			AuthMiddleware: func(ctx *gin.Context) { ctx.AbortWithStatus(http.StatusUpgradeRequired) },
		},
		RootShelf: corerouters.RootShelfRouterDependencies{
			AuthMiddleware:   func(ctx *gin.Context) { ctx.AbortWithStatus(http.StatusUpgradeRequired) },
			APIKeyMiddleware: func(ctx *gin.Context) { ctx.AbortWithStatus(http.StatusTeapot) },
		},
		User: corerouters.UserRouterDependencies{
			AuthMiddleware: func(ctx *gin.Context) { ctx.AbortWithStatus(http.StatusUpgradeRequired) },
		},
	})
}
