package user

import _ "embed"

var (
	//go:embed user_view.sql
	UserViewSQL string

	//go:embed user_quota.sql
	UserQuotaSQL string

	//go:embed user_quota_claim.sql
	ConsumeUserQuotaSQL string
)
