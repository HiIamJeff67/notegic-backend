package renderers

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"

	emailconfig "github.com/HiIamJeff67/notegic-backend/runtimes/email/configs"
	emailexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/email/exceptions"
)

type Renderer struct {
	config            emailconfig.RendererConfig
	expectedExtension string
	footerLinks       emailFooterLinks
}

type emailFooterLinks struct {
	PrivacyUrl string
	TermsUrl   string
	ContactUrl string
}

func buildEmailFooterLinks(config emailconfig.EmailLinksConfig) emailFooterLinks {
	return emailFooterLinks{
		PrivacyUrl: config.WebBaseUrl + "/privacy-policy",
		TermsUrl:   config.TermsUrl,
		ContactUrl: "mailto:" + config.ContactEmail,
	}
}

func (r *Renderer) render(pattern any) (string, error) {
	if filepath.Ext(r.config.TemplatePath) != "."+r.expectedExtension {
		return "", emailexceptions.
			NewRendererException("Email").
			InvalidTemplate()
	}

	templateBytes, err := os.ReadFile(r.config.TemplatePath)
	if err != nil {
		return "", emailexceptions.
			NewRendererException("Email").
			TemplateReadFailed(err)
	}

	extractedTemplate, err := template.New(
		strings.TrimSuffix(filepath.Base(r.config.TemplatePath), filepath.Ext(r.config.TemplatePath)),
	).Funcs(template.FuncMap{
		"privacyUrl": func() string {
			return r.footerLinks.PrivacyUrl
		},
		"termsUrl": func() string {
			return r.footerLinks.TermsUrl
		},
		"contactUrl": func() string {
			return r.footerLinks.ContactUrl
		},
	}).Parse(string(templateBytes))
	if err != nil {
		return "", emailexceptions.
			NewRendererException("Email").
			TemplateParseFailed(err)
	}

	var buffer bytes.Buffer
	if err := extractedTemplate.Execute(&buffer, pattern); err != nil {
		return "", emailexceptions.
			NewRendererException("Email").
			TemplateRenderFailed(err)
	}

	body := buffer.String()
	return body, nil
}

func (r *Renderer) ContentType() cemail.EmailContentType {
	return r.config.ContentType
}
