package middlewares

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEitherMiddlewareSelectsOnlyTheVerifiedRequestBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name      string
		condition bool
		expected  string
	}{
		{name: "passed", condition: true, expected: "client"},
		{name: "failed", condition: false, expected: "api"},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/", append(
				EitherMiddleware(
					[]gin.HandlerFunc{func(ctx *gin.Context) { ctx.Set("branch", "client"); ctx.Next() }},
					[]gin.HandlerFunc{func(ctx *gin.Context) { ctx.Set("branch", "api"); ctx.Next() }},
					func(*gin.Context) bool { return test.condition },
				),
				func(ctx *gin.Context) {
					value, _ := ctx.Get("branch")
					ctx.String(200, value.(string))
				},
			)...)

			response := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "/", nil)
			router.ServeHTTP(response, request)
			if response.Body.String() != test.expected {
				t.Fatalf("expected branch %q, got %q", test.expected, response.Body.String())
			}
		})
	}
}
