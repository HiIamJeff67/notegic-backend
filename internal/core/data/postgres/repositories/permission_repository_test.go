package repositories

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

func TestCheckPermissionAndGetOneByIdUsesOneQuery(t *testing.T) {
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

	var statements []string
	if err := db.Callback().Query().After("gorm:query").Register(
		"capture_permission_query",
		func(db *gorm.DB) {
			statements = append(statements, db.Statement.SQL.String())
		},
	); err != nil {
		t.Fatalf("failed to register query callback: %v", err)
	}

	resources := []struct {
		name  string
		table string
		join  string
		check func([]enums.AccessControlPermission)
	}{
		{
			name:  "root shelf",
			table: "RootShelfTable",
			join:  "UsersToShelvesTable",
			check: func(allowedPermissions []enums.AccessControlPermission) {
				repository := NewRootShelfRepository(scopes.NewRootShelfScope())
				_, _, _ = repository.CheckPermissionAndGetOneById(
					uuid.New(),
					uuid.New(),
					nil,
					allowedPermissions,
					options.WithDB(db),
				)
			},
		},
		{
			name:  "station",
			table: "StationTable",
			join:  "UsersToStationsTable",
			check: func(allowedPermissions []enums.AccessControlPermission) {
				repository := NewStationRepository(scopes.NewStationScope())
				_, _, _ = repository.CheckPermissionAndGetOneById(
					uuid.New(),
					uuid.New(),
					nil,
					allowedPermissions,
					options.WithDB(db),
				)
			},
		},
	}
	permissions := []struct {
		name                string
		allowedPermissions  []enums.AccessControlPermission
		wantPermissionScope bool
	}{
		{
			name:                "nil permissions",
			allowedPermissions:  nil,
			wantPermissionScope: false,
		},
		{
			name:                "empty permissions",
			allowedPermissions:  []enums.AccessControlPermission{},
			wantPermissionScope: false,
		},
		{
			name: "non-empty permissions",
			allowedPermissions: []enums.AccessControlPermission{
				enums.AccessControlPermission_Read,
			},
			wantPermissionScope: true,
		},
	}

	for _, resource := range resources {
		t.Run(resource.name, func(t *testing.T) {
			for _, permission := range permissions {
				t.Run(permission.name, func(t *testing.T) {
					statements = nil
					resource.check(permission.allowedPermissions)

					var resourceStatements []string
					for _, statement := range statements {
						if strings.Contains(statement, `FROM "`+resource.table+`"`) {
							resourceStatements = append(resourceStatements, statement)
						}
					}
					if len(resourceStatements) != 1 {
						t.Fatalf(
							"expected one %s query, got %d: %v",
							resource.table,
							len(resourceStatements),
							resourceStatements,
						)
					}

					sql := resourceStatements[0]
					if !strings.Contains(sql, `INNER JOIN "`+resource.join+`"`) {
						t.Fatalf(
							"expected the %s query to include %s",
							resource.table,
							resource.join,
						)
					}
					if strings.Contains(sql, "EXISTS") != permission.wantPermissionScope {
						t.Fatalf(
							"expected permission scope presence to be %t, SQL: %s",
							permission.wantPermissionScope,
							sql,
						)
					}
				})
			}
		})
	}
}
