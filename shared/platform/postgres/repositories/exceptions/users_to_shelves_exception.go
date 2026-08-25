package exceptions

type UsersToShelvesException struct {
	RepositoryException
}

func NewUsersToShelvesException() UsersToShelvesException {
	return UsersToShelvesException{RepositoryException: NewRepositoryException("UsersToShelves")}
}
