package shelfconstraints

import (
	_ "embed"
)

//go:embed shelf_foreign_keys.sql
var ShelfForeignKeysSQL string
