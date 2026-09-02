package exceptions

type UserQuotaException struct {
	CoreException
}

func NewUserQuotaException() UserQuotaException {
	return UserQuotaException{
		CoreException: CoreException{
			Domain: "UserQuota",
		},
	}
}
