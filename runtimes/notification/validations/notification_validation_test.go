package validations

import (
	"testing"

	validator "github.com/go-playground/validator/v10"

	cnotificationtypes "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/types"

	svalidations "github.com/HiIamJeff67/notegic-backend/shared/validations"
)

func TestRegisterNotificationValidations(t *testing.T) {
	validate := validator.New()
	svalidations.RegisterStringsValidation(validate)
	svalidations.RegisterTimesValidation(validate)
	RegisterNotificationValidation(validate)
	RegisterNewsValidation(validate)
	RegisterWarningValidation(validate)
	RegisterImportantValidation(validate)

	if err := validate.Struct(cnotificationtypes.NotificationMetadata{
		Type:            "news",
		Priority:        "normal",
		TemplateVersion: 1,
	}); err != nil {
		t.Fatalf("expected valid notification metadata, got %v", err)
	}
	if err := validate.Struct(cnotificationtypes.NotificationMetadata{
		Type:            "unknown",
		Priority:        "normal",
		TemplateVersion: 1,
	}); err == nil {
		t.Fatal("expected notification metadata validation error")
	}

	if err := validate.Struct(cnotificationtypes.NewsPayload{
		Title:   "Release update",
		Summary: "A new release is available.",
		Body:    "Read the release notes for more details.",
	}); err != nil {
		t.Fatalf("expected valid news payload, got %v", err)
	}
	if err := validate.Struct(cnotificationtypes.WarningPayload{Title: "Security warning"}); err == nil {
		t.Fatal("expected warning payload validation error")
	}
}
