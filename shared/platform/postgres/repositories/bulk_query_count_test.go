package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type queryCountLogger struct {
	count *int
}

func (l queryCountLogger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

func (queryCountLogger) Info(context.Context, string, ...interface{}) {}

func (queryCountLogger) Warn(context.Context, string, ...interface{}) {}

func (queryCountLogger) Error(context.Context, string, ...interface{}) {}

func (l queryCountLogger) Trace(
	_ context.Context,
	_ time.Time,
	fc func() (string, int64),
	_ error,
) {
	query, _ := fc()
	if query != "" {
		*l.count = *l.count + 1
	}
}

func TestAutomationBulkPermissionChecksDoNotScaleQueriesWithInputSize(t *testing.T) {
	tests := []struct {
		name string
		run  func(*gorm.DB, []uuid.UUID)
	}{
		{
			name: "sub shelf",
			run: func(db *gorm.DB, ids []uuid.UUID) {
				bulkInputs := make([]inputs.BulkCheckSubShelfPermissionInput, len(ids))
				for index, id := range ids {
					bulkInputs[index] = inputs.BulkCheckSubShelfPermissionInput{
						Id:     id,
						UserId: uuid.New(),
					}
				}
				repository := NewBulkSubShelfRepository(db, scopes.NewSubShelfScope())
				_, _, _ = repository.BulkCheckPermissionsAndGetManyByIds(
					bulkInputs,
					[]schemas.SubShelfRelation{},
					[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
				)
			},
		},
		{
			name: "routine",
			run: func(db *gorm.DB, ids []uuid.UUID) {
				bulkInputs := make([]inputs.BulkCheckRoutinePermissionInput, len(ids))
				for index, id := range ids {
					bulkInputs[index] = inputs.BulkCheckRoutinePermissionInput{
						Id:     id,
						UserId: uuid.New(),
					}
				}
				repository := NewBulkRoutineRepository(db, scopes.NewRoutineScope())
				_, _, _ = repository.BulkCheckPermissionsAndGetManyByIds(
					bulkInputs,
					[]schemas.RoutineRelation{},
					[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
				)
			},
		},
		{
			name: "block pack",
			run: func(db *gorm.DB, ids []uuid.UUID) {
				bulkInputs := make([]inputs.BulkCheckBlockPackPermissionInput, len(ids))
				for index, id := range ids {
					bulkInputs[index] = inputs.BulkCheckBlockPackPermissionInput{
						Id:     id,
						UserId: uuid.New(),
					}
				}
				repository := NewBulkBlockPackRepository(db, scopes.NewBlockPackScope())
				_, _, _ = repository.BulkCheckPermissionsAndGetManyByIds(
					bulkInputs,
					[]schemas.BlockPackRelation{},
					[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
				)
			},
		},
		{
			name: "material",
			run: func(db *gorm.DB, ids []uuid.UUID) {
				bulkInputs := make([]inputs.BulkCheckMaterialPermissionInput, len(ids))
				for index, id := range ids {
					bulkInputs[index] = inputs.BulkCheckMaterialPermissionInput{
						Id:     id,
						UserId: uuid.New(),
					}
				}
				repository := NewBulkMaterialRepository(db, scopes.NewMaterialScope())
				_, _, _ = repository.BulkCheckPermissionsAndGetManyByIds(
					bulkInputs,
					[]schemas.MaterialRelation{},
					[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			queryCounts := make([]int, 0, 2)
			for _, inputCount := range []int{1, 32} {
				queryCount := 0
				db, err := gorm.Open(
					postgres.New(postgres.Config{
						DSN: "host=localhost user=test dbname=test sslmode=disable",
					}),
					&gorm.Config{
						DisableAutomaticPing: true,
						DryRun:               true,
						Logger:               queryCountLogger{count: &queryCount},
					},
				)
				if err != nil {
					t.Fatalf("failed to create dry-run database: %v", err)
				}

				ids := make([]uuid.UUID, inputCount)
				for index := range ids {
					ids[index] = uuid.New()
				}
				test.run(db, ids)
				queryCounts = append(queryCounts, queryCount)
			}

			if queryCounts[0] == 0 || queryCounts[1] == 0 {
				t.Fatalf("expected bulk operation to execute queries, got counts %v", queryCounts)
			}
			if queryCounts[0] != queryCounts[1] {
				t.Fatalf("expected query count not to scale with input size, got counts %v", queryCounts)
			}
		})
	}
}
