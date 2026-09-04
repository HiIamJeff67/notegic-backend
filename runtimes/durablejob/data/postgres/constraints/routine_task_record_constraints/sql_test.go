package routinetaskrecordconstraints

import (
	"strings"
	"testing"
)

func TestRoutineTaskRecordForeignKeysCascadeDeletes(t *testing.T) {
	for _, fragment := range []string{
		"FOREIGN KEY (routine_record_id) REFERENCES \"RoutineRecordTable\" (id)",
		"FOREIGN KEY (routine_task_id) REFERENCES \"RoutineTaskTable\" (id)",
		"ON DELETE CASCADE",
	} {
		if !strings.Contains(RoutineTaskRecordForeignKeysSQL, fragment) {
			t.Fatalf("RoutineTaskRecordForeignKeysSQL must contain %q", fragment)
		}
	}
}
