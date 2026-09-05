package renderers

import (
	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailconfig "github.com/HiIamJeff67/notegic-backend/runtimes/email/configs"
	emailexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/email/exceptions"
)

type WelcomeEmailRenderer struct {
	renderer Renderer
}

func NewWelcomeEmailRenderer(config emailconfig.RendererConfig) (*WelcomeEmailRenderer, error) {
	if config.ContentType != cemail.EmailContentType_HTML {
		return nil, emailexceptions.
			NewRendererException("WelcomeEmail").
			InvalidContentType()
	}

	return &WelcomeEmailRenderer{
		renderer: Renderer{
			config:            config,
			expectedExtension: "html",
			footerLinks:       buildEmailFooterLinks(config.Links),
		},
	}, nil
}

func (r *WelcomeEmailRenderer) Render(pattern cemailevents.WelcomeEmailPattern) (string, error) {
	return r.renderer.render(pattern)
}

func (r *WelcomeEmailRenderer) ContentType() cemail.EmailContentType {
	return r.renderer.ContentType()
}
