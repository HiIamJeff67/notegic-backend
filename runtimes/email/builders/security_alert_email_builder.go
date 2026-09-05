package builders

import (
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailrenderers "github.com/HiIamJeff67/notegic-backend/runtimes/email/builders/renderers"
)

const securityAlertEmailSubject = "Security Alert - Some Suspicious Actions Detected on Your Account"

type SecurityAlertEmailBuilderInterface interface {
	Build(cemailevents.SendSecurityAlertEmailRequestDto) (*cemailevents.SendSecurityAlertEmailResponseDto, error)
}

type SecurityAlertEmailBuilder struct {
	renderer *emailrenderers.SecurityAlertEmailRenderer
}

func NewSecurityAlertEmailBuilder(renderer *emailrenderers.SecurityAlertEmailRenderer) SecurityAlertEmailBuilderInterface {
	return &SecurityAlertEmailBuilder{renderer: renderer}
}

func (b *SecurityAlertEmailBuilder) Build(
	request cemailevents.SendSecurityAlertEmailRequestDto,
) (*cemailevents.SendSecurityAlertEmailResponseDto, error) {
	body, err := b.renderer.Render(request.Pattern)
	if err != nil {
		return nil, err
	}

	return &cemailevents.SendSecurityAlertEmailResponseDto{
		To:               request.To,
		Subject:          securityAlertEmailSubject,
		Body:             body,
		EmailContentType: b.renderer.ContentType(),
		MaxRetries:       5,
		Priority:         3,
	}, nil
}

var _ SecurityAlertEmailBuilderInterface = (*SecurityAlertEmailBuilder)(nil)
