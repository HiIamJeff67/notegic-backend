package user

import (
	_ "embed"
)

//go:embed user_quota_claim.sql
var ConsumeUserQuotaSQL string
