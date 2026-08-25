package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/contexts"
)

func TestAllowedPermissionsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name                string
		middleware          gin.HandlerFunc
		expectedPermissions []enums.AccessControlPermission
	}{
		{
			name:       "above includes the requested permission",
			middleware: AllowedPermissionsAbove(enums.AccessControlPermission_Admin),
			expectedPermissions: []enums.AccessControlPermission{
				enums.AccessControlPermission_Admin,
				enums.AccessControlPermission_Owner,
			},
		},
		{
			name:       "below includes the requested permission",
			middleware: AllowedPermissionsBelow(enums.AccessControlPermission_Write),
			expectedPermissions: []enums.AccessControlPermission{
				enums.AccessControlPermission_Read,
				enums.AccessControlPermission_Write,
			},
		},
		{
			name: "within preserves the explicit permission set",
			middleware: AllowedPermissionsWithin(
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Write,
			),
			expectedPermissions: []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Write,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/", testCase.middleware, func(ctx *gin.Context) {
				allowedPermissions, exception := contexts.GetAllowedPermissions(ctx.Request.Context())
				if exception != nil {
					t.Fatal(exception)
				}

				if len(allowedPermissions) != len(testCase.expectedPermissions) {
					t.Fatalf("expected %d permissions, got %d", len(testCase.expectedPermissions), len(allowedPermissions))
				}
				for index, expectedPermission := range testCase.expectedPermissions {
					if allowedPermissions[index] != expectedPermission {
						t.Fatalf("expected permission %s at index %d, got %s", expectedPermission, index, allowedPermissions[index])
					}
				}

				ctx.Status(http.StatusNoContent)
			})

			responseRecorder := httptest.NewRecorder()
			router.ServeHTTP(responseRecorder, httptest.NewRequest(http.MethodGet, "/", nil))

			if responseRecorder.Code != http.StatusNoContent {
				t.Fatalf("expected status %d, got %d", http.StatusNoContent, responseRecorder.Code)
			}
		})
	}
}
