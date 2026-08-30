package senders

import (
	"context"

	"gopkg.in/gomail.v2"

	emailconfig "github.com/HiIamJeff67/notegic-backend/runtimes/email/configs"
	emailexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/email/exceptions"
	emailtypes "github.com/HiIamJeff67/notegic-backend/runtimes/email/types"
)

type EmailSenderInterface interface {
	Send(ctx context.Context, emailObject emailtypes.EmailObject) error
	SendAsync(ctx context.Context, emailObject emailtypes.EmailObject) error
}

type EmailSender struct {
	config emailconfig.SMTPConfig
}

func NewEmailSender(config emailconfig.SMTPConfig) EmailSenderInterface {
	return &EmailSender{config: config}
}

func (s *EmailSender) Send(_ context.Context, emailObject emailtypes.EmailObject) error {
	if !emailObject.EmailContentType.IsValidEnum() {
		return emailexceptions.
			NewRendererException("Email").
			InvalidContentType()
	}

	message := gomail.NewMessage()
	message.SetHeader("From", s.config.From)
	message.SetHeader("To", emailObject.To)
	message.SetHeader("Subject", emailObject.Subject)
	message.SetBody(emailObject.EmailContentType.String(), emailObject.Body)

	dialer := gomail.NewDialer(
		s.config.Host,
		s.config.Port,
		s.config.UserName,
		s.config.Password,
	)
	if err := dialer.DialAndSend(message); err != nil {
		return emailexceptions.
			NewDeliveryException("Email").
			DeliveryFailed(err)
	}

	return nil
}

func (s *EmailSender) SendAsync(ctx context.Context, emailObject emailtypes.EmailObject) error {
	return s.Send(ctx, emailObject)
}
