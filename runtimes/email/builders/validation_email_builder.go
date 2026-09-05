package builders

import (
	"time"

	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailrenderers "github.com/HiIamJeff67/notegic-backend/runtimes/email/builders/renderers"
)

const validationEmailSubject = "Verify Your Identity - Notegic Authentication Code"

type ValidationEmailBuilderInterface interface {
	Build(cemailevents.SendValidationEmailRequestDto) (*cemailevents.SendValidationEmailResponseDto, error)
}

type ValidationEmailBuilder struct {
	renderer *emailrenderers.ValidationEmailRenderer
}

func NewValidationEmailBuilder(renderer *emailrenderers.ValidationEmailRenderer) ValidationEmailBuilderInterface {
	return &ValidationEmailBuilder{renderer: renderer}
}

func (b *ValidationEmailBuilder) Build(
	request cemailevents.SendValidationEmailRequestDto,
) (*cemailevents.SendValidationEmailResponseDto, error) {
	request.Pattern.Email = request.To
	request.Pattern.ExpiryMinutes = int(time.Until(request.Pattern.ExpiredAt).Minutes())
	request.Pattern.RequestTime = time.Now().Format("2006-01-02 15:04:05 MST")
	body, err := b.renderer.Render(request.Pattern)
	if err != nil {
		return nil, err
	}

	return &cemailevents.SendValidationEmailResponseDto{
		To:               request.To,
		Subject:          validationEmailSubject,
		Body:             body,
		EmailContentType: b.renderer.ContentType(),
		MaxRetries:       2,
		Priority:         3,
	}, nil
}

var _ ValidationEmailBuilderInterface = (*ValidationEmailBuilder)(nil)
