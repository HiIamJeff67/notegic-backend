package developmentroutes

import (
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

func TestNewRouterAPIGatewayAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := NewRouter(APIRouteDependencies{
		CoreAdapter: coreadapters.NewCoreAdapter("http://core:7778", time.Second),
	})

	routes := router.Routes()
	for _, domain := range []string{
		"/stations",
		"/routines",
		"/routine-tags",
		"/routine-tasks",
		"/root-shelves",
		"/sub-shelves",
		"/materials",
		"/block-packs",
		"/blocks",
	} {
		if !hasRouteUnderDomain(routes, domain) {
			t.Errorf("APIGateway route allowlist is missing domain %q", domain)
		}
	}

	for _, domain := range []string{
		"/auth",
		"/users",
		"/me",
		"/notifications",
		"/realtime",
		"/graphql",
		"/static",
	} {
		if hasRouteUnderDomain(routes, domain) {
			t.Errorf("APIGateway unexpectedly exposed client-only domain %q", domain)
		}
	}
}

func hasRouteUnderDomain(routes gin.RoutesInfo, domain string) bool {
	prefix := "/api/development/v1" + domain
	for _, route := range routes {
		if strings.HasPrefix(route.Path, prefix) {
			return true
		}
	}
	return false
}
