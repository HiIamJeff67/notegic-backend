package repositories

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type automationBulkPermissionCheck struct {
	name string
	run  func(*gorm.DB, []uuid.UUID, uuid.UUID) (bool, error)
}

type integrationQueryStats struct {
	mu         sync.Mutex
	queryCount int
}

func (s *integrationQueryStats) LogMode(logger.LogLevel) logger.Interface {
	return s
}

func (s *integrationQueryStats) Info(context.Context, string, ...interface{}) {}

func (s *integrationQueryStats) Warn(context.Context, string, ...interface{}) {}

func (s *integrationQueryStats) Error(context.Context, string, ...interface{}) {}

func (s *integrationQueryStats) Trace(
	_ context.Context,
	_ time.Time,
	fc func() (string, int64),
	_ error,
) {
	query, _ := fc()
	if query == "" {
		return
	}

	s.mu.Lock()
	s.queryCount++
	s.mu.Unlock()
}

func (s *integrationQueryStats) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queryCount
}

func TestAutomationBulkPermissionChecksPostgresIntegration(t *testing.T) {
	db := openAutomationPostgresIntegrationDB(t)
	userId, subShelfIds, routineIds, blockPackIds, materialIds := seedAutomationPostgresIntegrationData(t, db)

	checks := []automationBulkPermissionCheck{
		{
			name: "sub shelf",
			run: func(db *gorm.DB, ids []uuid.UUID, userId uuid.UUID) (bool, error) {
				bulkInputs := make([]inputs.BulkCheckSubShelfPermissionInput, len(ids))
				for index, id := range ids {
					bulkInputs[index] = inputs.BulkCheckSubShelfPermissionInput{Id: id, UserId: userId}
				}
				successes, _, exception := NewSubShelfBulkRepositoryWithDB(db, scopes.NewSubShelfScope()).BulkCheckPermissionsAndGetManyByIds(
					bulkInputs,
					[]schemas.SubShelfRelation{},
					[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
				)
				if exception != nil {
					return false, exception
				}
				return allSuccessful(successes), nil
			},
		},
		{
			name: "routine",
			run: func(db *gorm.DB, ids []uuid.UUID, userId uuid.UUID) (bool, error) {
				bulkInputs := make([]inputs.BulkCheckRoutinePermissionInput, len(ids))
				for index, id := range ids {
					bulkInputs[index] = inputs.BulkCheckRoutinePermissionInput{Id: id, UserId: userId}
				}
				successes, _, exception := NewRoutineBulkRepositoryWithDB(db, scopes.NewRoutineScope()).BulkCheckPermissionsAndGetManyByIds(
					bulkInputs,
					[]schemas.RoutineRelation{},
					[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
				)
				if exception != nil {
					return false, exception
				}
				return allSuccessful(successes), nil
			},
		},
		{
			name: "block pack",
			run: func(db *gorm.DB, ids []uuid.UUID, userId uuid.UUID) (bool, error) {
				bulkInputs := make([]inputs.BulkCheckBlockPackPermissionInput, len(ids))
				for index, id := range ids {
					bulkInputs[index] = inputs.BulkCheckBlockPackPermissionInput{Id: id, UserId: userId}
				}
				successes, _, exception := NewBlockPackBulkRepositoryWithDB(db, scopes.NewBlockPackScope()).BulkCheckPermissionsAndGetManyByIds(
					bulkInputs,
					[]schemas.BlockPackRelation{},
					[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
				)
				if exception != nil {
					return false, exception
				}
				return allSuccessful(successes), nil
			},
		},
		{
			name: "material",
			run: func(db *gorm.DB, ids []uuid.UUID, userId uuid.UUID) (bool, error) {
				bulkInputs := make([]inputs.BulkCheckMaterialPermissionInput, len(ids))
				for index, id := range ids {
					bulkInputs[index] = inputs.BulkCheckMaterialPermissionInput{Id: id, UserId: userId}
				}
				successes, _, exception := NewMaterialBulkRepositoryWithDB(db, scopes.NewMaterialScope()).BulkCheckPermissionsAndGetManyByIds(
					bulkInputs,
					[]schemas.MaterialRelation{},
					[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
				)
				if exception != nil {
					return false, exception
				}
				return allSuccessful(successes), nil
			},
		},
	}

	idsByCheck := map[string][]uuid.UUID{
		"sub shelf":  subShelfIds,
		"routine":    routineIds,
		"block pack": blockPackIds,
		"material":   materialIds,
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			ids := idsByCheck[check.name]
			queryCounts := make([]int, 0, 2)
			for _, inputCount := range []int{1, 32} {
				stats := &integrationQueryStats{}
				ok, err := check.run(
					db.Session(&gorm.Session{Logger: stats}),
					ids[:inputCount],
					userId,
				)
				if err != nil {
					t.Fatalf("run bulk permission check with %d inputs: %v", inputCount, err)
				}
				if !ok {
					t.Fatalf("bulk permission check with %d inputs did not authorize all rows", inputCount)
				}
				queryCounts = append(queryCounts, stats.Count())
			}
			if queryCounts[0] == 0 || queryCounts[0] != queryCounts[1] {
				t.Fatalf("query count scaled with input size: %v", queryCounts)
			}

			latencies := make([]time.Duration, 0, 25)
			for run := 0; run < 25; run++ {
				stats := &integrationQueryStats{}
				startedAt := time.Now()
				ok, err := check.run(
					db.Session(&gorm.Session{Logger: stats}),
					ids[:32],
					userId,
				)
				latencies = append(latencies, time.Since(startedAt))
				if err != nil {
					t.Fatalf("run latency sample %d: %v", run, err)
				}
				if !ok {
					t.Fatalf("latency sample %d did not authorize all rows", run)
				}
				if stats.Count() != queryCounts[1] {
					t.Fatalf("query count changed during latency samples: got %d, want %d", stats.Count(), queryCounts[1])
				}
			}

			t.Logf("query count=%d, p95=%s, p99=%s", queryCounts[1], percentileDuration(latencies, 0.95), percentileDuration(latencies, 0.99))
		})
	}
}

func TestAutomationBulkPermissionChecksAreConcurrentSafe(t *testing.T) {
	db := openAutomationPostgresIntegrationDB(t)
	userId, _, _, _, materialIds := seedAutomationPostgresIntegrationData(t, db)

	start := make(chan struct{})
	results := make(chan error, 8)
	for worker := 0; worker < cap(results); worker++ {
		go func() {
			<-start
			bulkInputs := make([]inputs.BulkCheckMaterialPermissionInput, len(materialIds))
			for index, id := range materialIds {
				bulkInputs[index] = inputs.BulkCheckMaterialPermissionInput{Id: id, UserId: userId}
			}
			successes, _, exception := NewMaterialBulkRepositoryWithDB(db, scopes.NewMaterialScope()).BulkCheckPermissionsAndGetManyByIds(
				bulkInputs,
				[]schemas.MaterialRelation{},
				[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
			)
			if exception != nil {
				results <- exception
				return
			}
			if !allSuccessful(successes) {
				results <- fmt.Errorf("concurrent bulk permission check did not authorize all rows")
				return
			}
			results <- nil
		}()
	}
	close(start)

	for worker := 0; worker < cap(results); worker++ {
		select {
		case err := <-results:
			if err != nil {
				t.Fatalf("concurrent bulk permission check failed: %v", err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("concurrent bulk permission checks did not complete")
		}
	}
}

func TestAutomationBulkPermissionChecksRespectPostgresRowLocks(t *testing.T) {
	db := openAutomationPostgresIntegrationDB(t)
	userId, _, _, _, materialIds := seedAutomationPostgresIntegrationData(t, db)

	bulkInputs := []inputs.BulkCheckMaterialPermissionInput{{Id: materialIds[0], UserId: userId}}
	firstTransaction := db.Begin()
	if firstTransaction.Error != nil {
		t.Fatalf("begin first lock transaction: %v", firstTransaction.Error)
	}
	defer firstTransaction.Rollback()

	if _, _, exception := NewMaterialBulkRepositoryWithDB(db, scopes.NewMaterialScope()).BulkCheckPermissionsAndGetManyByIds(
		bulkInputs,
		[]schemas.MaterialRelation{},
		[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
		WithTransactionDB(firstTransaction),
		WithLockingStrength(LockingStrengthUpdate),
	); exception != nil {
		t.Fatalf("acquire first row lock: %v", exception)
	}

	secondTransaction := db.Begin()
	if secondTransaction.Error != nil {
		t.Fatalf("begin second lock transaction: %v", secondTransaction.Error)
	}
	defer secondTransaction.Rollback()
	secondResult := make(chan error, 1)
	go func() {
		_, _, exception := NewMaterialBulkRepositoryWithDB(db, scopes.NewMaterialScope()).BulkCheckPermissionsAndGetManyByIds(
			bulkInputs,
			[]schemas.MaterialRelation{},
			[]cenums.AccessControlPermission{cenums.AccessControlPermission_Read},
			WithTransactionDB(secondTransaction),
			WithLockingStrength(LockingStrengthUpdate),
		)
		if exception != nil {
			secondResult <- exception
			return
		}
		secondResult <- secondTransaction.Commit().Error
	}()

	select {
	case err := <-secondResult:
		t.Fatalf("second transaction completed while first lock was held: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	if err := firstTransaction.Commit().Error; err != nil {
		t.Fatalf("commit first lock transaction: %v", err)
	}
	select {
	case err := <-secondResult:
		if err != nil {
			t.Fatalf("second transaction failed after lock release: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("second transaction did not proceed after first lock release")
	}
}

func openAutomationPostgresIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("NOTEGIC_RUN_POSTGRES_REPOSITORY_INTEGRATION") != "1" {
		t.Skip("set NOTEGIC_RUN_POSTGRES_REPOSITORY_INTEGRATION=1 to run PostgreSQL repository integration tests")
	}

	config, err := platformpostgres.LoadConfig(
		integrationEnv("POSTGRES_REPOSITORY_INTEGRATION_HOST", "127.0.0.1"),
		integrationEnv("POSTGRES_REPOSITORY_INTEGRATION_USER", "notegic"),
		integrationEnv("POSTGRES_REPOSITORY_INTEGRATION_PASSWORD", "notegic"),
		integrationEnv("POSTGRES_REPOSITORY_INTEGRATION_NAME", "notegic_integration"),
		integrationEnv("POSTGRES_REPOSITORY_INTEGRATION_PORT", "15432"),
	)
	if err != nil {
		t.Fatalf("load PostgreSQL repository integration config: %v", err)
	}

	admin, err := platformpostgres.Connect(config)
	if err != nil {
		t.Fatalf("connect PostgreSQL repository integration database: %v", err)
	}
	schemaName := "automation_integration_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if result := admin.Exec(`CREATE SCHEMA "` + schemaName + `"`); result.Error != nil {
		platformpostgres.Disconnect(admin)
		t.Fatalf("create isolated integration schema: %v", result.Error)
	}

	dsn := platformpostgres.ConnectionString(config) + " options='-c search_path=" + schemaName + "'"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		admin.Exec(`DROP SCHEMA "` + schemaName + `" CASCADE`)
		platformpostgres.Disconnect(admin)
		t.Fatalf("connect isolated integration schema: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		admin.Exec(`DROP SCHEMA "` + schemaName + `" CASCADE`)
		platformpostgres.Disconnect(admin)
		t.Fatalf("get isolated integration database handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(16)
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		admin.Exec(`DROP SCHEMA "` + schemaName + `" CASCADE`)
		platformpostgres.Disconnect(admin)
		t.Fatalf("ping isolated integration schema: %v", err)
	}
	if err := createAutomationIntegrationTables(db); err != nil {
		sqlDB.Close()
		admin.Exec(`DROP SCHEMA "` + schemaName + `" CASCADE`)
		platformpostgres.Disconnect(admin)
		t.Fatalf("create isolated integration tables: %v", err)
	}

	t.Cleanup(func() {
		sqlDB.Close()
		admin.Exec(`DROP SCHEMA "` + schemaName + `" CASCADE`)
		platformpostgres.Disconnect(admin)
	})
	return db
}

func createAutomationIntegrationTables(db *gorm.DB) error {
	if result := db.Exec(`CREATE TABLE "SubShelfTable" (id uuid PRIMARY KEY, name text NOT NULL, root_shelf_id uuid NOT NULL, prev_sub_shelf_id uuid, path text NOT NULL, deleted_at timestamptz, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`); result.Error != nil {
		return result.Error
	}
	if result := db.Exec(`CREATE TABLE "RoutineTable" (id uuid PRIMARY KEY, station_id uuid NOT NULL, title text NOT NULL, description text NOT NULL, status text NOT NULL, is_pinned boolean NOT NULL, scheduled_start_at timestamptz NOT NULL, scheduled_end_at timestamptz NOT NULL, period text, timezone text NOT NULL, deleted_at timestamptz, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`); result.Error != nil {
		return result.Error
	}
	if result := db.Exec(`CREATE TABLE "BlockPackTable" (id uuid PRIMARY KEY, parent_sub_shelf_id uuid NOT NULL, name text NOT NULL, icon text, header_background_url text, block_count bigint NOT NULL, deleted_at timestamptz, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`); result.Error != nil {
		return result.Error
	}
	if result := db.Exec(`CREATE TABLE "MaterialTable" (id uuid PRIMARY KEY, parent_sub_shelf_id uuid NOT NULL, name text NOT NULL, size bigint NOT NULL, content_key text NOT NULL, content_type text NOT NULL, parse_media_type text NOT NULL, deleted_at timestamptz, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`); result.Error != nil {
		return result.Error
	}
	if result := db.Exec(`CREATE TABLE "UsersToShelvesTable" (user_id uuid NOT NULL, root_shelf_id uuid NOT NULL, permission text NOT NULL)`); result.Error != nil {
		return result.Error
	}
	if result := db.Exec(`CREATE TABLE "UsersToStationsTable" (user_id uuid NOT NULL, station_id uuid NOT NULL, permission text NOT NULL)`); result.Error != nil {
		return result.Error
	}

	return nil
}

func seedAutomationPostgresIntegrationData(
	t *testing.T,
	db *gorm.DB,
) (uuid.UUID, []uuid.UUID, []uuid.UUID, []uuid.UUID, []uuid.UUID) {
	t.Helper()
	const rowCount = 64
	now := time.Now().UTC()
	userId := uuid.New()
	rootShelfId := uuid.New()
	stationId := uuid.New()
	subShelfIds := make([]uuid.UUID, rowCount)
	routineIds := make([]uuid.UUID, rowCount)
	blockPackIds := make([]uuid.UUID, rowCount)
	materialIds := make([]uuid.UUID, rowCount)
	for index := 0; index < rowCount; index++ {
		subShelfIds[index] = uuid.New()
		routineIds[index] = uuid.New()
		blockPackIds[index] = uuid.New()
		materialIds[index] = uuid.New()
	}

	if result := db.Exec(
		`INSERT INTO "UsersToShelvesTable" (user_id, root_shelf_id, permission) VALUES (?, ?, ?)`,
		userId,
		rootShelfId,
		cenums.AccessControlPermission_Read,
	); result.Error != nil {
		t.Fatalf("seed shelf permission: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "UsersToStationsTable" (user_id, station_id, permission) VALUES (?, ?, ?)`,
		userId,
		stationId,
		cenums.AccessControlPermission_Read,
	); result.Error != nil {
		t.Fatalf("seed station permission: %v", result.Error)
	}

	subShelfValues := make([]string, rowCount)
	subShelfArgs := make([]any, 0, rowCount*8)
	for index, id := range subShelfIds {
		subShelfValues[index] = "(?, ?, ?, ?, ?, ?, ?, ?)"
		subShelfArgs = append(subShelfArgs, id, fmt.Sprintf("SubShelf %d", index), rootShelfId, nil, "{}", nil, now, now)
	}
	if result := db.Exec(
		`INSERT INTO "SubShelfTable" (id, name, root_shelf_id, prev_sub_shelf_id, path, deleted_at, updated_at, created_at) VALUES `+strings.Join(subShelfValues, ", "),
		subShelfArgs...,
	); result.Error != nil {
		t.Fatalf("seed sub shelves: %v", result.Error)
	}

	routineValues := make([]string, rowCount)
	routineArgs := make([]any, 0, rowCount*13)
	for index, id := range routineIds {
		routineValues[index] = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		routineArgs = append(routineArgs, id, stationId, fmt.Sprintf("Routine %d", index), "", cenums.RoutineStatus_Scheduled, false, now, now.Add(time.Hour), nil, "UTC", nil, now, now)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineTable" (id, station_id, title, description, status, is_pinned, scheduled_start_at, scheduled_end_at, period, timezone, deleted_at, updated_at, created_at) VALUES `+strings.Join(routineValues, ", "),
		routineArgs...,
	); result.Error != nil {
		t.Fatalf("seed routines: %v", result.Error)
	}

	blockPackValues := make([]string, rowCount)
	blockPackArgs := make([]any, 0, rowCount*9)
	for index, id := range blockPackIds {
		blockPackValues[index] = "(?, ?, ?, ?, ?, ?, ?, ?, ?)"
		blockPackArgs = append(blockPackArgs, id, subShelfIds[index], fmt.Sprintf("BlockPack %d", index), nil, nil, 0, nil, now, now)
	}
	if result := db.Exec(
		`INSERT INTO "BlockPackTable" (id, parent_sub_shelf_id, name, icon, header_background_url, block_count, deleted_at, updated_at, created_at) VALUES `+strings.Join(blockPackValues, ", "),
		blockPackArgs...,
	); result.Error != nil {
		t.Fatalf("seed block packs: %v", result.Error)
	}

	materialValues := make([]string, rowCount)
	materialArgs := make([]any, 0, rowCount*10)
	for index, id := range materialIds {
		materialValues[index] = "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		materialArgs = append(materialArgs, id, subShelfIds[index], fmt.Sprintf("Material %d", index), 0, fmt.Sprintf("integration-material-%d", index), "none", "", nil, now, now)
	}
	if result := db.Exec(
		`INSERT INTO "MaterialTable" (id, parent_sub_shelf_id, name, size, content_key, content_type, parse_media_type, deleted_at, updated_at, created_at) VALUES `+strings.Join(materialValues, ", "),
		materialArgs...,
	); result.Error != nil {
		t.Fatalf("seed materials: %v", result.Error)
	}

	return userId, subShelfIds, routineIds, blockPackIds, materialIds
}

func integrationEnv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func allSuccessful(successes []bool) bool {
	if len(successes) == 0 {
		return false
	}
	for _, success := range successes {
		if !success {
			return false
		}
	}
	return true
}

func percentileDuration(values []time.Duration, percentile float64) time.Duration {
	sortedValues := append([]time.Duration{}, values...)
	sort.Slice(sortedValues, func(left, right int) bool {
		return sortedValues[left] < sortedValues[right]
	})
	index := int(math.Ceil(float64(len(sortedValues))*percentile)) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sortedValues) {
		index = len(sortedValues) - 1
	}
	return sortedValues[index]
}
