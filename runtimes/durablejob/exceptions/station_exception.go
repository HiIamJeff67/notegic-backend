package exceptions

type StationException struct {
	DurableJobException
}

func NewStationException() StationException {
	return StationException{
		DurableJobException: DurableJobException{
			Domain: "Station",
		},
	}
}
