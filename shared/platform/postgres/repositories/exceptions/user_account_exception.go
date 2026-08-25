package exceptions

type UserAccountException struct {
	RepositoryException
}

func NewUserAccountException() UserAccountException {
	return UserAccountException{RepositoryException: NewRepositoryException("UserAccount")}
}
