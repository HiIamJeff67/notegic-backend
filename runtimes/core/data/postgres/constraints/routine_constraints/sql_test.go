package routineconstraints

import (
	"strings"
	"testing"
)

func TestRoutineForeignKeysCascadeDeletes(t *testing.T) {
	for _, fragment := range []string{
		"FOREIGN KEY (station_id) REFERENCES \"StationTable\" (id)",
		"REFERENCES \"RoutineTable\" (id, station_id)",
		"REFERENCES \"RoutineTagTable\" (id, owner_id)",
		"REFERENCES \"UsersToStationsTable\" (user_id, station_id)",
		"FOREIGN KEY (item_id, type)",
		"REFERENCES \"ItemTable\" (id, type)",
		"ON DELETE CASCADE",
	} {
		if !strings.Contains(RoutineForeignKeysSQL, fragment) {
			t.Fatalf("RoutineForeignKeysSQL must contain %q", fragment)
		}
	}
}
