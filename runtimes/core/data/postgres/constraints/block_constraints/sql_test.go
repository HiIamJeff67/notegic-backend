package blockconstraints

import (
	"strings"
	"testing"
)

func TestBlockForeignKeysUseConfiguredDeleteActions(t *testing.T) {
	for _, fragment := range []string{
		"FOREIGN KEY (block_pack_id) REFERENCES \"BlockPackTable\" (id)",
		"FOREIGN KEY (parent_block_id) REFERENCES \"BlockTable\" (id)",
		"FOREIGN KEY (prev_block_id) REFERENCES \"BlockTable\" (id)",
		"FOREIGN KEY (next_block_id) REFERENCES \"BlockTable\" (id)",
		"ON DELETE CASCADE",
		"ON DELETE SET NULL",
	} {
		if !strings.Contains(BlockForeignKeysSQL, fragment) {
			t.Fatalf("BlockForeignKeysSQL must contain %q", fragment)
		}
	}
}
