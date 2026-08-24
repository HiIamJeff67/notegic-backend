package usersql

import (
	_ "embed"
)

//go:embed get_user_data_cache_by_id.sql
var GetUserDataCacheByIdSQL string
