package exceptions

import "testing"

func TestNewEmailException(t *testing.T) {
	exception := NewEmailException("Email")
	if exception.Domain != "Email" {
		t.Fatalf("domain = %q, want Email", exception.Domain)
	}
}
