package views

import _ "embed"

//go:embed user_view.sql
var userViewSQL string

var MigratingViewSQLs = []string{
	userViewSQL,
}
