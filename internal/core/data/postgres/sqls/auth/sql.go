package authsql

import (
	_ "embed"
)

//go:embed update_auth_code_by_email.sql
var UpdateAuthCodeSQL string

//go:embed reset_email.sql
var ResetEmailSQL string

//go:embed validate_email_by_auth_code.sql
var ValidateEmailSQL string

//go:embed delete_me.sql
var DeleteMeSQL string
