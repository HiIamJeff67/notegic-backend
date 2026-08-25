package renderers

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	emailconfig "github.com/HiIamJeff67/notegic-backend/internal/email/configs"
)

func TestRendererRenderAndContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType cemail.EmailContentType
		extension   string
		wantType    cemail.EmailContentType
	}{
		{
			name:        "html",
			contentType: cemail.EmailContentType_HTML,
			extension:   ".html",
			wantType:    cemail.EmailContentType_HTML,
		},
		{
			name:        "plain text",
			contentType: cemail.EmailContentType_PlainText,
			extension:   ".txt",
			wantType:    cemail.EmailContentType_PlainText,
		},
		{
			name:        "markdown",
			contentType: cemail.EmailContentType_Markdown,
			extension:   ".md",
			wantType:    cemail.EmailContentType_Markdown,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			templatePath := filepath.Join(t.TempDir(), "message"+test.extension)
			if err := os.WriteFile(templatePath, []byte("Hello, {{.Name}}!"), 0o600); err != nil {
				t.Fatal(err)
			}

			renderer, exception := NewRenderer(emailconfig.RendererConfig{
				TemplatePath: templatePath,
				ContentType:  test.contentType,
			})
			if exception != nil {
				t.Fatalf("NewRenderer() exception = %v", exception)
			}
			if renderer.ContentType() != test.wantType {
				t.Fatalf("ContentType() = %q, want %q", renderer.ContentType(), test.wantType)
			}

			body, exception := renderer.Render(map[string]any{"Name": "Notegic"})
			if exception != nil {
				t.Fatalf("Render() exception = %v", exception)
			}
			if body != "Hello, Notegic!" {
				t.Fatalf("Render() = %q, want %q", body, "Hello, Notegic!")
			}
		})
	}
}

func TestNewRendererRejectsUnsupportedContentType(t *testing.T) {
	renderer, exception := NewRenderer(emailconfig.RendererConfig{
		ContentType: cemail.EmailContentType("application/octet-stream"),
	})
	if renderer != nil {
		t.Fatalf("NewRenderer() renderer = %#v, want nil", renderer)
	}
	if exception == nil {
		t.Fatal("NewRenderer() exception = nil, want an exception")
	}
	var emailException *cexceptions.Exception
	if !errors.As(exception, &emailException) {
		t.Fatalf("exception type = %T, want *exceptions.Exception", exception)
	}
	if emailException.Reason != "InvalidContentType" {
		t.Fatalf("exception.Reason = %q, want %q", emailException.Reason, "InvalidContentType")
	}
}

func TestRendererRejectsInvalidTemplate(t *testing.T) {
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
			renderer, exception := NewRenderer(test.config)
			if exception != nil {
				t.Fatalf("NewRenderer() exception = %v", exception)
			}

			_, exception = renderer.Render(nil)
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

func TestRendererRejectsMalformedTemplate(t *testing.T) {
	templatePath := filepath.Join(t.TempDir(), "message.html")
	if err := os.WriteFile(templatePath, []byte("{{"), 0o600); err != nil {
		t.Fatal(err)
	}

	renderer, exception := NewRenderer(emailconfig.RendererConfig{
		TemplatePath: templatePath,
		ContentType:  cemail.EmailContentType_HTML,
	})
	if exception != nil {
		t.Fatalf("NewRenderer() exception = %v", exception)
	}

	_, exception = renderer.Render(nil)
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
