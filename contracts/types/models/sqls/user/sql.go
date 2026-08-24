package user

import _ "embed"

var (
	//go:embed user_view.sql
	UserViewSQL string
)
