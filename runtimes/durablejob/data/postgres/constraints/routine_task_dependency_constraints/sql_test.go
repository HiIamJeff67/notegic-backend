package routinetaskdependencyconstraints

import (
	"strings"
	"testing"
)

func TestRoutineTaskDependencyForeignKeysCascadeDeletes(t *testing.T) {
	for _, fragment := range []string{
		"FOREIGN KEY (routine_task_id)",
		"FOREIGN KEY (previous_routine_task_id)",
		"REFERENCES \"RoutineTaskTable\" (id)",
		"ON DELETE CASCADE",
	} {
		if !strings.Contains(RoutineTaskDependencyForeignKeysSQL, fragment) {
			t.Fatalf("RoutineTaskDependencyForeignKeysSQL must contain %q", fragment)
		}
	}
}
