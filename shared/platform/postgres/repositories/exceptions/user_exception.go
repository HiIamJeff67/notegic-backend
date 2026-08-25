package exceptions

type UserException struct {
	RepositoryException
}

func NewUserException() UserException {
	return UserException{RepositoryException: NewRepositoryException("User")}
}
