package renderers

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"

	emailconfig "github.com/HiIamJeff67/notegic-backend/internal/email/configs"
	emailexceptions "github.com/HiIamJeff67/notegic-backend/internal/email/exceptions"
)

type RendererInterface interface {
	Render(data map[string]any) (string, error)
	ContentType() cemail.EmailContentType
}

type Renderer struct {
	config            emailconfig.RendererConfig
	expectedExtension string
}

func NewRenderer(config emailconfig.RendererConfig) (RendererInterface, error) {
	switch config.ContentType {
	case cemail.EmailContentType_HTML:
		return &HTMLEmailRenderer{Renderer: newRenderer(config, "html")}, nil
	case cemail.EmailContentType_PlainText:
		return &PlainTextEmailRenderer{Renderer: newRenderer(config, "txt")}, nil
	case cemail.EmailContentType_Markdown:
		return &MarkdownEmailRenderer{Renderer: newRenderer(config, "md")}, nil
	default:
		return nil, emailexceptions.
			NewRendererException("Email").
			InvalidContentType()
	}
}

func newRenderer(
	config emailconfig.RendererConfig,
	expectedExtension string,
) Renderer {
	return Renderer{
		config:            config,
		expectedExtension: expectedExtension,
	}
}

func (r *Renderer) Render(data map[string]any) (string, error) {
	return renderTemplate(r.config, r.expectedExtension, data)
}

func (r *Renderer) ContentType() cemail.EmailContentType {
	return r.config.ContentType
}

func renderTemplate(
	config emailconfig.RendererConfig,
	expectedExtension string,
	data map[string]any,
) (string, error) {
	if filepath.Ext(config.TemplatePath) != "."+expectedExtension {
		return "", emailexceptions.
			NewRendererException("Email").
			InvalidTemplate()
	}

	templateBytes, err := os.ReadFile(config.TemplatePath)
	if err != nil {
		return "", emailexceptions.
			NewRendererException("Email").
			TemplateReadFailed(err)
	}

	extractedTemplate, err := template.New(strings.TrimSuffix(filepath.Base(config.TemplatePath), filepath.Ext(config.TemplatePath))).Parse(string(templateBytes))
	if err != nil {
		return "", emailexceptions.
			NewRendererException("Email").
			TemplateParseFailed(err)
	}

	var buffer bytes.Buffer
	if err := extractedTemplate.Execute(&buffer, data); err != nil {
		return "", emailexceptions.
			NewRendererException("Email").
			TemplateRenderFailed(err)
	}

	body := buffer.String()
	return body, nil
}
