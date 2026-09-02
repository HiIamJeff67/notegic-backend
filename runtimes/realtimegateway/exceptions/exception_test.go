package exceptions

import "testing"

func TestNewRealtimeGatewayException(t *testing.T) {
	exception := NewRealtimeGatewayException()
	if exception.Domain != "RealtimeGateway" {
		t.Fatalf("domain = %q, want RealtimeGateway", exception.Domain)
	}
}
