package exceptions

import "testing"

func TestNewRealtimeGatewayException(t *testing.T) {
	exception := NewRealtimeGatewayException("RateLimitRecord")
	if exception.Domain != "RateLimitRecord" {
		t.Fatalf("domain = %q, want RateLimitRecord", exception.Domain)
	}
}
