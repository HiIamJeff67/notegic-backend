package senders

import (
	"context"

	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailrenderers "github.com/HiIamJeff67/notegic-backend/runtimes/email/renderers"
	emailtypes "github.com/HiIamJeff67/notegic-backend/runtimes/email/types"
)

const welcomeEmailSubject = "Welcome to Notegic - Thanks for the Registration"

type WelcomeEmailSenderInterface interface {
	Send(context.Context, cemailevents.SendWelcomeEmailRequestDto) error
	SendAsync(context.Context, cemailevents.SendWelcomeEmailRequestDto) error
}

type WelcomeEmailSender struct {
	renderer    emailrenderers.RendererInterface
	enqueueFunc emailtypes.EnqueueFunc
}

func NewWelcomeEmailSender(renderer emailrenderers.RendererInterface, enqueueFunc emailtypes.EnqueueFunc) WelcomeEmailSenderInterface {
	return &WelcomeEmailSender{renderer: renderer, enqueueFunc: enqueueFunc}
}

func (s *WelcomeEmailSender) Send(
	_ context.Context,
	request cemailevents.SendWelcomeEmailRequestDto,
) error {
	body, err := s.renderer.Render(map[string]any{
		"UserName": request.UserName,
		"Email":    request.To,
		"Status":   request.Status,
	})
	if err != nil {
		return err
	}

	return s.enqueueFunc(
		emailtypes.EmailObject{
			To:               request.To,
			Subject:          welcomeEmailSubject,
			Body:             body,
			EmailContentType: s.renderer.ContentType(),
		},
		emailtypes.EmailTaskType_Welcome,
		1,
		3,
	)
}

func (s *WelcomeEmailSender) SendAsync(
	ctx context.Context,
	request cemailevents.SendWelcomeEmailRequestDto,
) error {
	return s.Send(ctx, request)
}

var _ WelcomeEmailSenderInterface = (*WelcomeEmailSender)(nil)
