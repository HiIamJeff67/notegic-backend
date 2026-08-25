package validations

import (
	"testing"

	validator "github.com/go-playground/validator/v10"
)

func TestIsGoogleAuthenticationCode(t *testing.T) {
	validate := validator.New()
	RegisterStringsValidation(validate)

	type request struct {
		AuthenticationCode string `validate:"isgoogleauthenticationcode"`
	}

	if err := validate.Struct(request{AuthenticationCode: "4/0AfJohx-example_code"}); err != nil {
		t.Fatalf("expected a valid Google authentication code: %v", err)
	}
	if err := validate.Struct(request{AuthenticationCode: " code"}); err == nil {
		t.Fatal("expected a whitespace-prefixed Google authentication code to be invalid")
	}
}
