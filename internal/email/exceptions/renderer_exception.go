package exceptions

import (
	"net/http"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type RendererException struct {
	EmailException
}

func NewRendererException(domain string) RendererException {
	return RendererException{EmailException: NewEmailException(domain)}
}

func (e RendererException) InvalidContentType() *cexceptions.Exception {
	return cexceptions.New("InvalidContentType", e.Domain, "RenderEmail", "The email content type is invalid", http.StatusBadRequest)
}

func (e RendererException) InvalidTemplate() *cexceptions.Exception {
	return cexceptions.New("InvalidTemplate", e.Domain, "RenderEmail", "The email template is invalid", http.StatusBadRequest)
}

func (e RendererException) TemplateReadFailed(cause error) *cexceptions.Exception {
	return cexceptions.New("TemplateReadFailed", e.Domain, "RenderEmail", "Failed to read the email template", http.StatusInternalServerError, true).WithOrigin(cause)
}

func (e RendererException) TemplateParseFailed(cause error) *cexceptions.Exception {
	return cexceptions.New("TemplateParseFailed", e.Domain, "RenderEmail", "Failed to parse the email template", http.StatusInternalServerError, true).WithOrigin(cause)
}

func (e RendererException) TemplateRenderFailed(cause error) *cexceptions.Exception {
	return cexceptions.New("TemplateRenderFailed", e.Domain, "RenderEmail", "Failed to render the email template", http.StatusInternalServerError, true).WithOrigin(cause)
}
