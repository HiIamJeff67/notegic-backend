package repositories

import (
	"testing"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func TestAllowedPermissionsPresence(t *testing.T) {
	t.Run("option omitted", func(t *testing.T) {
		parsedOptions := ParseRepositoryOptions()

		if parsedOptions.HasAllowedPermissions() {
			t.Fatal("expected omitted allowed permissions to disable the permission filter")
		}
	})

	t.Run("empty option", func(t *testing.T) {
		parsedOptions := ParseRepositoryOptions(
			WithAllowedPermissions([]enums.AccessControlPermission{}),
		)

		if !parsedOptions.HasAllowedPermissions() {
			t.Fatal("expected an explicit empty permission policy to enable the permission filter")
		}
	})

	t.Run("non-empty option", func(t *testing.T) {
		parsedOptions := ParseRepositoryOptions(
			WithAllowedPermissions([]enums.AccessControlPermission{
				enums.AccessControlPermission_Read,
			}),
		)

		if !parsedOptions.HasAllowedPermissions() {
			t.Fatal("expected a non-empty permission policy to enable the permission filter")
		}
	})
}
