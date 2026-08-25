package apiexceptions

type StationException struct {
	Exception
}

func NewStationException() StationException {
	return StationException{
		Exception: NewException("Station"),
	}
}
