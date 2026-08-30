package core

import (
	"context"

	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailsenders "github.com/HiIamJeff67/notegic-backend/runtimes/email/senders"
)

type SenderInterface interface {
	SendWelcomeEmail(context.Context, cemailevents.SendWelcomeEmailRequestDto) error
	SendValidationEmail(context.Context, cemailevents.SendValidationEmailRequestDto) error
	SendSecurityAlertEmail(context.Context, cemailevents.SendSecurityAlertEmailRequestDto) error
}

type Sender struct {
	welcome       emailsenders.WelcomeEmailSenderInterface
	validation    emailsenders.ValidationEmailSenderInterface
	securityAlert emailsenders.SecurityAlertEmailSenderInterface
}

func NewSender(
	welcome emailsenders.WelcomeEmailSenderInterface,
	validation emailsenders.ValidationEmailSenderInterface,
	securityAlert emailsenders.SecurityAlertEmailSenderInterface,
) SenderInterface {
	return &Sender{
		welcome:       welcome,
		validation:    validation,
		securityAlert: securityAlert,
	}
}

func (s *Sender) SendWelcomeEmail(
	ctx context.Context,
	request cemailevents.SendWelcomeEmailRequestDto,
) error {
	return s.welcome.Send(ctx, request)
}

func (s *Sender) SendValidationEmail(
	ctx context.Context,
	request cemailevents.SendValidationEmailRequestDto,
) error {
	return s.validation.Send(ctx, request)
}

func (s *Sender) SendSecurityAlertEmail(
	ctx context.Context,
	request cemailevents.SendSecurityAlertEmailRequestDto,
) error {
	return s.securityAlert.Send(ctx, request)
}

var _ SenderInterface = (*Sender)(nil)
