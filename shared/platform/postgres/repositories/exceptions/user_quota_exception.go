package exceptions

type UserQuotaException struct {
	RepositoryException
}

func NewUserQuotaException() UserQuotaException {
	return UserQuotaException{RepositoryException: NewRepositoryException("UserQuota")}
}
