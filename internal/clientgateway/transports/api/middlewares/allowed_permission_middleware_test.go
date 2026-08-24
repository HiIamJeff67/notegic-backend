package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/contexts"
)

func TestAllowedPermissionsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name                string
		middleware          gin.HandlerFunc
		expectedPermissions []enumcontract.AccessControlPermission
	}{
		{
			name:       "above includes the requested permission",
			middleware: AllowedPermissionsAbove(enumcontract.AccessControlPermission_Admin),
			expectedPermissions: []enumcontract.AccessControlPermission{
				enumcontract.AccessControlPermission_Admin,
				enumcontract.AccessControlPermission_Owner,
			},
		},
		{
			name:       "below includes the requested permission",
			middleware: AllowedPermissionsBelow(enumcontract.AccessControlPermission_Write),
			expectedPermissions: []enumcontract.AccessControlPermission{
				enumcontract.AccessControlPermission_Read,
				enumcontract.AccessControlPermission_Write,
			},
		},
		{
			name: "within preserves the explicit permission set",
			middleware: AllowedPermissionsWithin(
				enumcontract.AccessControlPermission_Owner,
				enumcontract.AccessControlPermission_Write,
			),
			expectedPermissions: []enumcontract.AccessControlPermission{
				enumcontract.AccessControlPermission_Owner,
				enumcontract.AccessControlPermission_Write,
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
