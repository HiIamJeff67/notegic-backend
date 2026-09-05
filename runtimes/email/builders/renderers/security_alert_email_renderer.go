package renderers

import (
	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailconfig "github.com/HiIamJeff67/notegic-backend/runtimes/email/configs"
	emailexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/email/exceptions"
)

type SecurityAlertEmailRenderer struct {
	renderer Renderer
}

func NewSecurityAlertEmailRenderer(config emailconfig.RendererConfig) (*SecurityAlertEmailRenderer, error) {
	if config.ContentType != cemail.EmailContentType_HTML {
		return nil, emailexceptions.
			NewRendererException("SecurityAlertEmail").
			InvalidContentType()
	}

	return &SecurityAlertEmailRenderer{
		renderer: Renderer{
			config:            config,
			expectedExtension: "html",
			footerLinks:       buildEmailFooterLinks(config.Links),
		},
	}, nil
}

func (r *SecurityAlertEmailRenderer) Render(pattern cemailevents.SecurityAlertEmailPattern) (string, error) {
	return r.renderer.render(pattern)
}

func (r *SecurityAlertEmailRenderer) ContentType() cemail.EmailContentType {
	return r.renderer.ContentType()
}
