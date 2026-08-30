package badgesql

import (
	_ "embed"
)

var (
	//go:embed delete_all_my_badges.sql
	DeleteAllMyBadgesSQL string
)
