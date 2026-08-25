package views

import (
	usersql "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/sqls/user"
)

var MigratingViewSQLs = []string{
	usersql.UserViewSQL,
}
