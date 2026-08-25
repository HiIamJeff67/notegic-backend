package senders

import (
	"context"

	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	emailrenderers "github.com/HiIamJeff67/notegic-backend/internal/email/renderers"
	emailtypes "github.com/HiIamJeff67/notegic-backend/internal/email/types"
)

const securityAlertEmailSubject = "Security Alert - Some Suspicious Actions Detected on Your Account"

type SecurityAlertEmailSenderInterface interface {
	Send(context.Context, cemailevents.SendSecurityAlertEmailRequestDto) error
	SendAsync(context.Context, cemailevents.SendSecurityAlertEmailRequestDto) error
}

type SecurityAlertEmailSender struct {
	renderer    emailrenderers.RendererInterface
	enqueueFunc emailtypes.EnqueueFunc
}

func NewSecurityAlertEmailSender(renderer emailrenderers.RendererInterface, enqueueFunc emailtypes.EnqueueFunc) SecurityAlertEmailSenderInterface {
	return &SecurityAlertEmailSender{renderer: renderer, enqueueFunc: enqueueFunc}
}

func (s *SecurityAlertEmailSender) Send(
	_ context.Context,
	request cemailevents.SendSecurityAlertEmailRequestDto,
) error {
	body, err := s.renderer.Render(map[string]any{
		"UserName":         request.UserName,
		"Status":           request.Status,
		"AlertType":        request.AlertType,
		"Reason":           request.Reason,
		"TimeOfOccurrence": request.TimeOfOccurrence,
		"OtherDetails":     request.OtherDetails,
	})
	if err != nil {
		return err
	}

	return s.enqueueFunc(
		emailtypes.EmailObject{
			To:               request.To,
			Subject:          securityAlertEmailSubject,
			Body:             body,
			EmailContentType: s.renderer.ContentType(),
		},
		emailtypes.EmailTaskType_Security,
		5,
		3,
	)
}

func (s *SecurityAlertEmailSender) SendAsync(
	ctx context.Context,
	request cemailevents.SendSecurityAlertEmailRequestDto,
) error {
	return s.Send(ctx, request)
}

var _ SecurityAlertEmailSenderInterface = (*SecurityAlertEmailSender)(nil)
