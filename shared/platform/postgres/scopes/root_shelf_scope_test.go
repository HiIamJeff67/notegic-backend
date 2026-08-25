package scopes

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

func TestRootShelfPermissionPolicy(t *testing.T) {
	db, err := gorm.Open(
		postgres.New(postgres.Config{
			DSN: "host=localhost user=test dbname=test sslmode=disable",
		}),
		&gorm.Config{
			DisableAutomaticPing: true,
			DryRun:               true,
		},
	)
	if err != nil {
		t.Fatalf("failed to create dry-run database: %v", err)
	}

	scope := NewRootShelfScope()
	rootShelfId := uuid.New()
	userId := uuid.New()

	t.Run("omitted policy skips permission filter", func(t *testing.T) {
		var rootShelves []schemas.RootShelf
		result := db.
			Scopes(scope.PassPermissionCheck(rootShelfId, userId, nil)).
			Find(&rootShelves)

		if strings.Contains(result.Statement.SQL.String(), "EXISTS") {
			t.Fatal("expected an omitted policy to skip the permission filter")
		}
	})

	tests := []struct {
		name        string
		permissions []cenums.AccessControlPermission
	}{
		{
			name:        "empty policy",
			permissions: []cenums.AccessControlPermission{},
		},
		{
			name: "read policy",
			permissions: []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Read,
			},
		},
		{
			name: "write policy",
			permissions: []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Write,
			},
		},
		{
			name: "admin policy",
			permissions: []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Admin,
			},
		},
		{
			name: "owner policy",
			permissions: []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rootShelves []schemas.RootShelf
			result := db.
				Scopes(scope.PassPermissionCheck(rootShelfId, userId, tt.permissions)).
				Find(&rootShelves)

			if !strings.Contains(result.Statement.SQL.String(), "EXISTS") {
				t.Fatal("expected an explicit policy to apply the permission filter")
			}
		})
	}
}
