package exceptions

import (
	"testing"
)

func TestCacheUnavailable(t *testing.T) {
	cause := &testError{message: "Redis unavailable"}
	exception := NewCacheException("RateLimitRecord").Unavailable(cause)

	if exception.Reason != "CacheClientUnavailable" || !exception.Retryable {
		t.Fatalf("unexpected cache exception: %#v", exception)
	}
	if exception.Origin() != cause {
		t.Fatal("cache exception does not preserve its origin")
	}
}

type testError struct{ message string }

func (e *testError) Error() string { return e.message }
