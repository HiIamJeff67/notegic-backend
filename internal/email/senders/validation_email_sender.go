package senders

import (
	"context"
	"time"

	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailrenderers "github.com/HiIamJeff67/notegic-backend/internal/email/renderers"
	emailtypes "github.com/HiIamJeff67/notegic-backend/internal/email/types"
)

const validationEmailSubject = "Verify Your Identity - Notegic Authentication Code"

type ValidationEmailSenderInterface interface {
	Send(context.Context, cemailevents.SendValidationEmailRequestDto) error
	SendAsync(context.Context, cemailevents.SendValidationEmailRequestDto) error
}

type ValidationEmailSender struct {
	renderer    emailrenderers.RendererInterface
	enqueueFunc emailtypes.EnqueueFunc
}

func NewValidationEmailSender(renderer emailrenderers.RendererInterface, enqueueFunc emailtypes.EnqueueFunc) ValidationEmailSenderInterface {
	return &ValidationEmailSender{renderer: renderer, enqueueFunc: enqueueFunc}
}

func (s *ValidationEmailSender) Send(
	_ context.Context,
	request cemailevents.SendValidationEmailRequestDto,
) error {
	body, err := s.renderer.Render(map[string]any{
		"UserName":      request.UserName,
		"Email":         request.To,
		"AuthCode":      request.AuthCode,
		"UserAgent":     request.UserAgent,
		"ExpiryMinutes": int(time.Until(request.ExpiredAt).Minutes()),
		"RequestTime":   time.Now().Format("2006-01-02 15:04:05 MST"),
	})
	if err != nil {
		return err
	}

	return s.enqueueFunc(
		emailtypes.EmailObject{
			To:               request.To,
			Subject:          validationEmailSubject,
			Body:             body,
			EmailContentType: s.renderer.ContentType(),
		},
		emailtypes.EmailTaskType_Validation,
		2,
		3,
	)
}

func (s *ValidationEmailSender) SendAsync(
	ctx context.Context,
	request cemailevents.SendValidationEmailRequestDto,
) error {
	return s.Send(ctx, request)
}

var _ ValidationEmailSenderInterface = (*ValidationEmailSender)(nil)
