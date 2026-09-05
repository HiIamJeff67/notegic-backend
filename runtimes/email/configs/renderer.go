package config

import cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"

type RendererConfig struct {
	TemplatePath string
	ContentType  cemail.EmailContentType
	Links        EmailLinksConfig
}

type EmailLinksConfig struct {
	WebBaseUrl   string
	TermsUrl     string
	ContactEmail string
}

type RendererConfigs struct {
	Welcome       RendererConfig
	Validation    RendererConfig
	SecurityAlert RendererConfig
}

func loadRendererConfig(links EmailLinksConfig) RendererConfigs {
	return RendererConfigs{
		Welcome: RendererConfig{
			TemplatePath: "templates/welcome_email_template.html",
			ContentType:  cemail.EmailContentType_HTML,
			Links:        links,
		},
		Validation: RendererConfig{
			TemplatePath: "templates/validation_email_template.html",
			ContentType:  cemail.EmailContentType_HTML,
			Links:        links,
		},
		SecurityAlert: RendererConfig{
			TemplatePath: "templates/security_alert_email_template.html",
			ContentType:  cemail.EmailContentType_HTML,
			Links:        links,
		},
	}
}
