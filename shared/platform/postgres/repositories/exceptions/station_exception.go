package exceptions

type StationException struct {
	RepositoryException
}

func NewStationException() StationException {
	return StationException{RepositoryException: NewRepositoryException("Station")}
}
