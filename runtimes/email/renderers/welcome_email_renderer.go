package renderers

import (
	emailconfig "github.com/HiIamJeff67/notegic-backend/runtimes/email/configs"
)

type HTMLEmailRenderer struct {
	Renderer
}

func NewHTMLEmailRenderer(config emailconfig.RendererConfig) RendererInterface {
	return &HTMLEmailRenderer{Renderer: newRenderer(config, "html")}
}
