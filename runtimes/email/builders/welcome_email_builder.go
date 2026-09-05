package builders

import (
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailrenderers "github.com/HiIamJeff67/notegic-backend/runtimes/email/builders/renderers"
)

const welcomeEmailSubject = "Welcome to Notegic - Thanks for the Registration"

type WelcomeEmailBuilderInterface interface {
	Build(cemailevents.SendWelcomeEmailRequestDto) (*cemailevents.SendWelcomeEmailResponseDto, error)
}

type WelcomeEmailBuilder struct {
	renderer *emailrenderers.WelcomeEmailRenderer
}

func NewWelcomeEmailBuilder(renderer *emailrenderers.WelcomeEmailRenderer) WelcomeEmailBuilderInterface {
	return &WelcomeEmailBuilder{renderer: renderer}
}

func (b *WelcomeEmailBuilder) Build(
	request cemailevents.SendWelcomeEmailRequestDto,
) (*cemailevents.SendWelcomeEmailResponseDto, error) {
	request.Pattern.Email = request.To
	body, err := b.renderer.Render(request.Pattern)
	if err != nil {
		return nil, err
	}

	return &cemailevents.SendWelcomeEmailResponseDto{
		To:               request.To,
		Subject:          welcomeEmailSubject,
		Body:             body,
		EmailContentType: b.renderer.ContentType(),
		MaxRetries:       1,
		Priority:         3,
	}, nil
}

var _ WelcomeEmailBuilderInterface = (*WelcomeEmailBuilder)(nil)
