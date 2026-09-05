package renderers

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	emailconfig "github.com/HiIamJeff67/notegic-backend/runtimes/email/configs"
)

var testEmailLinks = emailconfig.EmailLinksConfig{
	WebBaseUrl:   "https://notegic.example.com",
	TermsUrl:     "https://notegic.example.com/terms",
	ContactEmail: "support@notegic.example.com",
}

func TestWelcomeEmailRendererRendersPattern(t *testing.T) {
	renderer, exception := NewWelcomeEmailRenderer(emailconfig.RendererConfig{
		TemplatePath: filepath.Join("..", "..", "templates", "welcome_email_template.html"),
		ContentType:  cemail.EmailContentType_HTML,
		Links:        testEmailLinks,
	})
	if exception != nil {
		t.Fatalf("NewWelcomeEmailRenderer() exception = %v", exception)
	}

	body, exception := renderer.Render(cemailevents.WelcomeEmailPattern{
		UserName: "Avery Lin",
		Email:    "avery.lin@example.com",
		Status:   "Active",
		RoutineItems: []cemailevents.WelcomeEmailRoutineItemPattern{
			{
				Name:   "Review the launch brief",
				Status: "Ready",
			},
			{
				Name:   "Publish the weekly summary",
				Status: "Waiting",
			},
		},
	})
	if exception != nil {
		t.Fatalf("Render() exception = %v", exception)
	}

	for _, wantFragment := range []string{
		"src=\"https://api.notegic.app/api/development/v1/static/logo\"",
		"Hello, Avery Lin.",
		"Avery Lin's workspace",
		"Station&nbsp;&nbsp;·&nbsp;&nbsp;Avery Lin",
		"2 items",
		"Review the launch brief",
		"Ready",
		"Publish the weekly summary",
		"Waiting",
		"Your account&nbsp;&nbsp;·&nbsp;&nbsp;avery.lin@example.com&nbsp;&nbsp;·&nbsp;&nbsp;Active",
		"Sent to avery.lin@example.com because a Notegic account was created.",
		"href=\"https://notegic.example.com/privacy-policy\">Privacy",
		"href=\"https://notegic.example.com/terms\">Terms",
		"href=\"mailto:support@notegic.example.com\">Contact",
	} {
		if !strings.Contains(body, wantFragment) {
			t.Fatalf("Render() body does not contain %q", wantFragment)
		}
	}
	for _, unwantedFragment := range []string{
		"Avery's workspace",
		"07:42",
		"12 items",
		"Morning ignition",
		"Clear the landing shelf",
	} {
		if strings.Contains(body, unwantedFragment) {
			t.Fatalf("Render() body contains fake fragment %q", unwantedFragment)
		}
	}

	body, exception = renderer.Render(cemailevents.WelcomeEmailPattern{
		UserName: "Avery Lin",
		Email:    "avery.lin@example.com",
		Status:   "Active",
	})
	if exception != nil {
		t.Fatalf("Render() empty routine items exception = %v", exception)
	}
	for _, wantFragment := range []string{
		"No routines yet",
		"Create one when your work needs a rhythm.",
	} {
		if !strings.Contains(body, wantFragment) {
			t.Fatalf("Render() body does not contain %q", wantFragment)
		}
	}
}

func TestValidationEmailRendererRendersPattern(t *testing.T) {
	renderer, exception := NewValidationEmailRenderer(emailconfig.RendererConfig{
		TemplatePath: filepath.Join("..", "..", "templates", "validation_email_template.html"),
		ContentType:  cemail.EmailContentType_HTML,
		Links:        testEmailLinks,
	})
	if exception != nil {
		t.Fatalf("NewValidationEmailRenderer() exception = %v", exception)
	}

	body, exception := renderer.Render(cemailevents.ValidationEmailPattern{
		UserName:      "Avery Lin",
		Email:         "avery.lin@example.com",
		AuthCode:      "481729",
		UserAgent:     "Chrome 140 on macOS 15",
		ExpiryMinutes: 10,
		RequestTime:   "2026-09-04 09:42 UTC",
	})
	if exception != nil {
		t.Fatalf("Render() exception = %v", exception)
	}

	for _, wantFragment := range []string{
		"src=\"https://api.notegic.app/api/development/v1/static/logo\"",
		"Your Notegic verification code: 481729",
		"Hello, Avery Lin.",
		"class=\"code\" style=\"margin:0;color:#f1efe7;font-size:32px;line-height:38px;letter-spacing:8px;font-weight:800\">481729</p>",
		"10 minutes remaining",
		"Requested for avery.lin@example.com&nbsp;&nbsp;·&nbsp;&nbsp;2026-09-04 09:42 UTC",
		"Source&nbsp;&nbsp;·&nbsp;&nbsp;Chrome 140 on macOS 15",
		"href=\"https://notegic.example.com/privacy-policy\">Privacy",
		"href=\"https://notegic.example.com/terms\">Terms",
		"href=\"mailto:support@notegic.example.com\">Contact",
	} {
		if !strings.Contains(body, wantFragment) {
			t.Fatalf("Render() body does not contain %q", wantFragment)
		}
	}
}

func TestSecurityAlertEmailRendererRendersPattern(t *testing.T) {
	renderer, exception := NewSecurityAlertEmailRenderer(emailconfig.RendererConfig{
		TemplatePath: filepath.Join("..", "..", "templates", "security_alert_email_template.html"),
		ContentType:  cemail.EmailContentType_HTML,
		Links:        testEmailLinks,
	})
	if exception != nil {
		t.Fatalf("NewSecurityAlertEmailRenderer() exception = %v", exception)
	}

	timeOfOccurrence := time.Date(2026, time.September, 4, 9, 37, 0, 0, time.UTC)
	body, exception := renderer.Render(cemailevents.SecurityAlertEmailPattern{
		UserName:         "Avery Lin",
		Status:           "Active",
		AlertType:        "New sign-in detected",
		Reason:           "A sign-in was completed from a device we have not seen on your account before.",
		TimeOfOccurrence: timeOfOccurrence,
		OtherDetails:     "Taipei, TW · Chrome 140 · macOS 15",
	})
	if exception != nil {
		t.Fatalf("Render() exception = %v", exception)
	}

	for _, wantFragment := range []string{
		"src=\"https://api.notegic.app/api/development/v1/static/logo\"",
		"Hello, Avery Lin.",
		"Detected event</p><p style=\"margin:0 0 7px;color:#f1efe7;font-size:15px;line-height:22px;font-weight:700\">New sign-in detected</p>",
		"A sign-in was completed from a device we have not seen on your account before.",
		"Account status</td><td class=\"detail-value\" valign=\"top\" style=\"padding:12px 16px;background:#d0cec7;color:#22221f;font-size:12px;font-weight:700;border-top:1px solid #aaa8a0\">Active</td>",
		"2026-09-04",
		"Details</td><td class=\"detail-value\" valign=\"top\" style=\"padding:12px 16px;background:#d0cec7;color:#22221f;font-size:12px;font-weight:700;border-top:1px solid #aaa8a0\">Taipei, TW · Chrome 140 · macOS 15</td>",
		"Notegic security notice for Avery Lin.",
		"href=\"https://notegic.example.com/privacy-policy\">Privacy",
		"href=\"https://notegic.example.com/terms\">Terms",
		"href=\"mailto:support@notegic.example.com\">Contact",
	} {
		if !strings.Contains(body, wantFragment) {
			t.Fatalf("Render() body does not contain %q", wantFragment)
		}
	}
}

func TestEmailRenderersRejectInvalidContentType(t *testing.T) {
	tests := []struct {
		name string
		new  func(emailconfig.RendererConfig) (any, error)
	}{
		{
			name: "welcome",
			new: func(config emailconfig.RendererConfig) (any, error) {
				return NewWelcomeEmailRenderer(config)
			},
		},
		{
			name: "validation",
			new: func(config emailconfig.RendererConfig) (any, error) {
				return NewValidationEmailRenderer(config)
			},
		},
		{
			name: "security alert",
			new: func(config emailconfig.RendererConfig) (any, error) {
				return NewSecurityAlertEmailRenderer(config)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, exception := test.new(emailconfig.RendererConfig{
				ContentType: cemail.EmailContentType("application/octet-stream"),
			})
			if exception == nil {
				t.Fatal("exception = nil, want an exception")
			}
			var emailException *cexceptions.Exception
			if !errors.As(exception, &emailException) {
				t.Fatalf("exception type = %T, want *exceptions.Exception", exception)
			}
			if emailException.Reason != "InvalidContentType" {
				t.Fatalf("exception.Reason = %q, want %q", emailException.Reason, "InvalidContentType")
			}
		})
	}
}

func TestWelcomeEmailRendererRejectsInvalidTemplates(t *testing.T) {
	tests := []struct {
		name       string
		config     emailconfig.RendererConfig
		wantReason string
	}{
		{
			name: "wrong extension",
			config: emailconfig.RendererConfig{
				TemplatePath: filepath.Join(t.TempDir(), "message.txt"),
				ContentType:  cemail.EmailContentType_HTML,
			},
			wantReason: "InvalidTemplate",
		},
		{
			name: "missing file",
			config: emailconfig.RendererConfig{
				TemplatePath: filepath.Join(t.TempDir(), "missing.html"),
				ContentType:  cemail.EmailContentType_HTML,
			},
			wantReason: "TemplateReadFailed",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer, exception := NewWelcomeEmailRenderer(test.config)
			if exception != nil {
				t.Fatalf("NewWelcomeEmailRenderer() exception = %v", exception)
			}

			_, exception = renderer.Render(cemailevents.WelcomeEmailPattern{})
			if exception == nil {
				t.Fatal("Render() exception = nil, want an exception")
			}
			var emailException *cexceptions.Exception
			if !errors.As(exception, &emailException) {
				t.Fatalf("exception type = %T, want *exceptions.Exception", exception)
			}
			if emailException.Reason != test.wantReason {
				t.Fatalf("exception.Reason = %q, want %q", emailException.Reason, test.wantReason)
			}
		})
	}
}

func TestWelcomeEmailRendererRejectsMalformedTemplate(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "message.html")
	if err := os.WriteFile(templatePath, []byte("{{"), 0o600); err != nil {
		t.Fatal(err)
	}

	renderer, exception := NewWelcomeEmailRenderer(emailconfig.RendererConfig{
		TemplatePath: templatePath,
		ContentType:  cemail.EmailContentType_HTML,
	})
	if exception != nil {
		t.Fatalf("NewWelcomeEmailRenderer() exception = %v", exception)
	}

	_, exception = renderer.Render(cemailevents.WelcomeEmailPattern{})
	if exception == nil {
		t.Fatal("Render() exception = nil, want an exception")
	}
	var emailException *cexceptions.Exception
	if !errors.As(exception, &emailException) {
		t.Fatalf("exception type = %T, want *exceptions.Exception", exception)
	}
	if emailException.Reason != "TemplateParseFailed" {
		t.Fatalf("exception.Reason = %q, want %q", emailException.Reason, "TemplateParseFailed")
	}
}
