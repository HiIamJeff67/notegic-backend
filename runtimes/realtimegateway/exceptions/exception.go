package exceptions

type RealtimeGatewayException struct {
	Domain string
}

func NewRealtimeGatewayException(domain string) RealtimeGatewayException {
	return RealtimeGatewayException{Domain: domain}
}
