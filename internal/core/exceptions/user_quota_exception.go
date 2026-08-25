package apiexceptions

type UserQuotaException struct {
	Exception
}

func NewUserQuotaException() UserQuotaException {
	return UserQuotaException{Exception: NewException("UserQuota")}
}
