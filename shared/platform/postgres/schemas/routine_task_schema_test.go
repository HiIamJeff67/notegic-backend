package schemas

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRoutineTaskUsesExplicitDependencyAssociations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:routine-task-schema?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	statement := &gorm.Statement{DB: db}
	if err := statement.Parse(&RoutineTask{}); err != nil {
		t.Fatalf("parse routine task schema: %v", err)
	}
	phaseField, exists := statement.Schema.FieldsByDBName["phase"]
	if !exists {
		t.Fatal("routine task phase field is not mapped")
	}
	if phaseField.NotNull {
		t.Fatal("routine task phase field must be nullable")
	}

	for _, relationName := range []string{"PreviousDependencies", "NextDependencies"} {
		relation := statement.Schema.Relationships.Relations[relationName]
		if relation == nil || relation.JoinTable != nil {
			t.Fatalf("routine task relation %s is not an explicit dependency association", relationName)
		}
		if relation.FieldSchema == nil || relation.FieldSchema.Table != "RoutineDependencyTable" {
			t.Fatalf("routine task relation %s table = %q, want RoutineDependencyTable", relationName, relation.FieldSchema.Table)
		}
	}
}
