package renderers

import (
	emailconfig "github.com/HiIamJeff67/notegic-backend/runtimes/email/configs"
)

type PlainTextEmailRenderer struct {
	Renderer
}

func NewPlainTextEmailRenderer(config emailconfig.RendererConfig) RendererInterface {
	return &PlainTextEmailRenderer{Renderer: newRenderer(config, "txt")}
}
