package contexts

import (
	"context"
	"testing"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func TestGetOptionalAllowedPermissionsReturnsNilWhenRouteHasNoScope(t *testing.T) {
	allowedPermissions, exception := GetOptionalAllowedPermissions(context.Background())
	if exception != nil {
		t.Fatalf("expected no exception for an absent optional scope: %v", exception)
	}
	if allowedPermissions != nil {
		t.Fatalf("expected nil permissions for an absent optional scope, got %#v", allowedPermissions)
	}
}

func TestGetOptionalAllowedPermissionsClonesRouteScope(t *testing.T) {
	original := []enums.AccessControlPermission{
		enums.AccessControlPermission_Read,
	}
	allowedPermissions, exception := GetOptionalAllowedPermissions(
		WithAllowedPermissions(context.Background(), original),
	)
	if exception != nil {
		t.Fatalf("expected optional scope to be valid: %v", exception)
	}
	allowedPermissions[0] = enums.AccessControlPermission_Admin
	if original[0] != enums.AccessControlPermission_Read {
		t.Fatal("expected optional permissions to be cloned")
	}
}
