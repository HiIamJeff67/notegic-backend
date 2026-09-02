package exceptions

import "testing"

func TestNewCoreExceptionKeepsRuntimeDomain(t *testing.T) {
	exception := NewCoreException()
	if exception.Domain != "Core" {
		t.Fatalf("domain = %q, want Core", exception.Domain)
	}
}

func TestNewAuthExceptionKeepsDomain(t *testing.T) {
	exception := NewAuthException()
	if exception.Domain != "Auth" {
		t.Fatalf("domain = %q, want Auth", exception.Domain)
	}
	wrongPassword := exception.WrongPassword()
	if wrongPassword.Domain != "Auth" {
		t.Fatalf("wrong-password domain = %q, want Auth", wrongPassword.Domain)
	}
}
