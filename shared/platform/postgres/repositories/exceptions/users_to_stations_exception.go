package exceptions

type UsersToStationsException struct {
	RepositoryException
}

func NewUsersToStationsException() UsersToStationsException {
	return UsersToStationsException{RepositoryException: NewRepositoryException("UsersToStations")}
}
