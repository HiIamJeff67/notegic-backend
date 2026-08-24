package views

import (
	cusersql "github.com/HiIamJeff67/notegic-backend/contracts/types/models/sqls/user"
)

var MigratingViewSQLs = []string{
	cusersql.UserViewSQL,
}
