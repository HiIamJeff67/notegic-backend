package scopes

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
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
		permissions []enums.AccessControlPermission
	}{
		{
			name:        "empty policy",
			permissions: []enums.AccessControlPermission{},
		},
		{
			name: "read policy",
			permissions: []enums.AccessControlPermission{
				enums.AccessControlPermission_Read,
			},
		},
		{
			name: "write policy",
			permissions: []enums.AccessControlPermission{
				enums.AccessControlPermission_Write,
			},
		},
		{
			name: "admin policy",
			permissions: []enums.AccessControlPermission{
				enums.AccessControlPermission_Admin,
			},
		},
		{
			name: "owner policy",
			permissions: []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
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
