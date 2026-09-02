package exceptions

type RealtimeGatewayException struct {
	Domain string
}

func NewRealtimeGatewayException() RealtimeGatewayException {
	return RealtimeGatewayException{
		Domain: "RealtimeGateway",
	}
}
