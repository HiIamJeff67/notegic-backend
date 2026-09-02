package repositories_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	repositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

func TestRoutineTaskDependencyRepositorySupportsEmptyGraphAndBatchMutations(t *testing.T) {
	db := openRoutineTaskDependencyRepositoryTestDatabase(t, "mutations")
	routineId, taskIds := seedRoutineTaskDependencyRepositoryData(t, db, 4)
	repository := repositories.NewRoutineTaskDependencyRepository(db)

	dependencies, exception := repository.GetAllByRoutineId(routineId)
	if exception != nil {
		t.Fatalf("get empty routine task dependency graph: %v", exception)
	}
	if len(dependencies) != 0 {
		t.Fatalf("empty dependency graph count = %d, want 0", len(dependencies))
	}

	firstDescription := "first dependency"
	secondDescription := "second dependency"
	if exception := repository.CreateManyByRoutineId(
		routineId,
		[]sinputs.CreateRoutineTaskDependencyInput{
			{
				RoutineTaskId:         taskIds[1],
				PreviousRoutineTaskId: taskIds[0],
				Description:           firstDescription,
				Progress:              25,
			},
			{
				RoutineTaskId:         taskIds[2],
				PreviousRoutineTaskId: taskIds[0],
				Description:           secondDescription,
				Progress:              50,
			},
		},
	); exception != nil {
		t.Fatalf("create multiple routine task dependencies: %v", exception)
	}

	assertRoutineDefinitionVersion(t, db, routineId, 2)

	updatedDescription := "updated dependency"
	if exception := repository.UpdateManyByRoutineId(
		routineId,
		[]sinputs.UpdateRoutineTaskDependencyInput{
			{
				RoutineTaskId:         taskIds[1],
				PreviousRoutineTaskId: taskIds[0],
				Description:           updatedDescription,
				Progress:              75,
			},
			{
				RoutineTaskId:         taskIds[2],
				PreviousRoutineTaskId: taskIds[0],
				Description:           secondDescription,
				Progress:              100,
			},
		},
	); exception != nil {
		t.Fatalf("update multiple routine task dependencies: %v", exception)
	}

	assertRoutineDefinitionVersion(t, db, routineId, 3)
	dependencies, exception = repository.GetAllByRoutineId(routineId)
	if exception != nil {
		t.Fatalf("get updated routine task dependencies: %v", exception)
	}
	if len(dependencies) != 2 {
		t.Fatalf("updated dependency count = %d, want 2", len(dependencies))
	}
	if dependencies[0].Description != updatedDescription || dependencies[0].Progress != 75 {
		t.Fatalf("first dependency = %#v, want updated values", dependencies[0])
	}
	if dependencies[1].Progress != 100 {
		t.Fatalf("second dependency progress = %d, want 100", dependencies[1].Progress)
	}

	if exception := repository.UpdateManyByRoutineId(
		routineId,
		[]sinputs.UpdateRoutineTaskDependencyInput{
			{
				RoutineTaskId:         taskIds[1],
				PreviousRoutineTaskId: taskIds[0],
				Description:           updatedDescription,
				Progress:              75,
			},
			{
				RoutineTaskId:         taskIds[2],
				PreviousRoutineTaskId: taskIds[0],
				Description:           secondDescription,
				Progress:              100,
			},
		},
	); exception != nil {
		t.Fatalf("repeat routine task dependency update: %v", exception)
	}
	assertRoutineDefinitionVersion(t, db, routineId, 4)
	dependencies, exception = repository.GetAllByRoutineId(routineId)
	if exception != nil {
		t.Fatalf("get dependencies after repeated update: %v", exception)
	}
	if len(dependencies) != 2 {
		t.Fatalf("dependencies after repeated update = %d, want 2", len(dependencies))
	}

	deletedCount, exception := repository.DeleteManyByRoutineId(
		routineId,
		[]sinputs.RoutineTaskDependencyKey{
			{
				RoutineTaskId:         taskIds[1],
				PreviousRoutineTaskId: taskIds[0],
			},
			{
				RoutineTaskId:         taskIds[2],
				PreviousRoutineTaskId: taskIds[0],
			},
		},
	)
	if exception != nil {
		t.Fatalf("delete multiple routine task dependencies: %v", exception)
	}
	if deletedCount != 2 {
		t.Fatalf("deleted dependency count = %d, want 2", deletedCount)
	}
	assertRoutineDefinitionVersion(t, db, routineId, 5)
}

func TestRoutineTaskDependencyRepositoryRollsBackFailedBatchAndReportsConflict(t *testing.T) {
	db := openRoutineTaskDependencyRepositoryTestDatabase(t, "rollback")
	routineId, taskIds := seedRoutineTaskDependencyRepositoryData(t, db, 3)
	repository := repositories.NewRoutineTaskDependencyRepository(db)

	if exception := repository.CreateManyByRoutineId(
		routineId,
		[]sinputs.CreateRoutineTaskDependencyInput{
			{
				RoutineTaskId:         taskIds[1],
				PreviousRoutineTaskId: taskIds[0],
				Description:           "valid dependency",
				Progress:              10,
			},
			{
				RoutineTaskId:         taskIds[2],
				PreviousRoutineTaskId: taskIds[0],
				Description:           "invalid dependency",
				Progress:              101,
			},
		},
	); exception == nil {
		t.Fatal("expected invalid batch to fail")
	}

	dependencies, exception := repository.GetAllByRoutineId(routineId)
	if exception != nil {
		t.Fatalf("get dependencies after failed batch: %v", exception)
	}
	if len(dependencies) != 0 {
		t.Fatalf("dependencies after failed batch = %d, want 0", len(dependencies))
	}
	assertRoutineDefinitionVersion(t, db, routineId, 1)

	if _, exception := repository.CreateOneByRoutineId(
		routineId,
		sinputs.CreateRoutineTaskDependencyInput{
			RoutineTaskId:         taskIds[1],
			PreviousRoutineTaskId: taskIds[0],
		},
	); exception != nil {
		t.Fatalf("create dependency before duplicate check: %v", exception)
	}

	if _, exception := repository.CreateOneByRoutineId(
		routineId,
		sinputs.CreateRoutineTaskDependencyInput{
			RoutineTaskId:         taskIds[1],
			PreviousRoutineTaskId: taskIds[0],
		},
	); exception == nil {
		t.Fatal("expected duplicate dependency to fail")
	} else if exception.Reason != "DependencyAlreadyExists" || exception.HTTPStatusCode() != http.StatusConflict {
		t.Fatalf("duplicate dependency exception = %v, want conflict", exception)
	}
	assertRoutineDefinitionVersion(t, db, routineId, 2)
}

func openRoutineTaskDependencyRepositoryTestDatabase(t *testing.T, name string) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(
		sqlite.Open("file:routine-task-dependency-repository-"+name+"?mode=memory&cache=shared"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatalf("open routine task dependency test database: %v", err)
	}
	for _, statement := range []string{
		`CREATE TABLE "RoutineTable" (id BLOB PRIMARY KEY, definition_version INTEGER NOT NULL, status TEXT NOT NULL, updated_at DATETIME NOT NULL)`,
		`CREATE TABLE "RoutineTaskTable" (id BLOB PRIMARY KEY, routine_id BLOB NOT NULL)`,
		`CREATE TABLE "RoutineDependencyTable" (routine_task_id BLOB NOT NULL, previous_routine_task_id BLOB NOT NULL, description TEXT NOT NULL DEFAULT '', progress INTEGER NOT NULL DEFAULT 0, updated_at DATETIME NOT NULL, created_at DATETIME NOT NULL, PRIMARY KEY (routine_task_id, previous_routine_task_id), CHECK (progress >= 0 AND progress <= 100), CHECK (length(description) <= 128))`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("create routine task dependency test table: %v", err)
		}
	}
	return db
}

func seedRoutineTaskDependencyRepositoryData(t *testing.T, db *gorm.DB, taskCount int) (uuid.UUID, []uuid.UUID) {
	t.Helper()

	routineId := uuid.New()
	if err := db.Exec(
		`INSERT INTO "RoutineTable" (id, definition_version, status, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`,
		routineId,
		1,
		"Scheduled",
	).Error; err != nil {
		t.Fatalf("seed routine task dependency routine: %v", err)
	}

	taskIds := make([]uuid.UUID, taskCount)
	for index := range taskIds {
		taskIds[index] = uuid.New()
		if err := db.Exec(
			`INSERT INTO "RoutineTaskTable" (id, routine_id) VALUES (?, ?)`,
			taskIds[index],
			routineId,
		).Error; err != nil {
			t.Fatalf("seed routine task dependency task %d: %v", index, err)
		}
	}
	return routineId, taskIds
}

func assertRoutineDefinitionVersion(t *testing.T, db *gorm.DB, routineId uuid.UUID, expected int64) {
	t.Helper()

	var routine sschemas.Routine
	if err := db.Model(&sschemas.Routine{}).Where("id = ?", routineId).First(&routine).Error; err != nil {
		t.Fatalf("load routine definition version: %v", err)
	}
	if routine.DefinitionVersion != expected {
		t.Fatalf("routine definition version = %d, want %d", routine.DefinitionVersion, expected)
	}
}
