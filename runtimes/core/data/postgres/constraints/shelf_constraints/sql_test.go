package shelfconstraints

import (
	"strings"
	"testing"
)

func TestShelfForeignKeysCascadeDeletes(t *testing.T) {
	for _, fragment := range []string{
		"FOREIGN KEY (root_shelf_id) REFERENCES \"RootShelfTable\" (id)",
		"FOREIGN KEY (parent_sub_shelf_id) REFERENCES \"SubShelfTable\" (id)",
		"FOREIGN KEY (root_shelf_id) REFERENCES \"RootShelfTable\" (id)",
		"ON DELETE CASCADE",
	} {
		if !strings.Contains(ShelfForeignKeysSQL, fragment) {
			t.Fatalf("ShelfForeignKeysSQL must contain %q", fragment)
		}
	}
}
