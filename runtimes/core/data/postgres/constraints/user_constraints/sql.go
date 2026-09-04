package userconstraints

import (
	_ "embed"
)

//go:embed user_foreign_keys.sql
var UserForeignKeysSQL string
