package renderers

import (
	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailconfig "github.com/HiIamJeff67/notegic-backend/runtimes/email/configs"
	emailexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/email/exceptions"
)

type ValidationEmailRenderer struct {
	renderer Renderer
}

func NewValidationEmailRenderer(config emailconfig.RendererConfig) (*ValidationEmailRenderer, error) {
	if config.ContentType != cemail.EmailContentType_HTML {
		return nil, emailexceptions.
			NewRendererException("ValidationEmail").
			InvalidContentType()
	}

	return &ValidationEmailRenderer{
		renderer: Renderer{
			config:            config,
			expectedExtension: "html",
			footerLinks:       buildEmailFooterLinks(config.Links),
		},
	}, nil
}

func (r *ValidationEmailRenderer) Render(pattern cemailevents.ValidationEmailPattern) (string, error) {
	return r.renderer.render(pattern)
}

func (r *ValidationEmailRenderer) ContentType() cemail.EmailContentType {
	return r.renderer.ContentType()
}
