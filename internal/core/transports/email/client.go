package email

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	crepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
)

type ClientInterface interface {
	SendWelcomeEmail(ctx context.Context, requestDto cemailevents.SendWelcomeEmailRequestDto) *cexceptions.Exception
	SendValidationEmail(ctx context.Context, requestDto cemailevents.SendValidationEmailRequestDto) *cexceptions.Exception
	SendSecurityAlertEmail(ctx context.Context, requestDto cemailevents.SendSecurityAlertEmailRequestDto) *cexceptions.Exception
}

type Client struct {
	db *gorm.DB
}

func NewClient(db *gorm.DB) ClientInterface {
	return &Client{db: db}
}

func (c *Client) SendWelcomeEmail(
	ctx context.Context,
	requestDto cemailevents.SendWelcomeEmailRequestDto,
) *cexceptions.Exception {
	requestDto.RequestId = uuid.New()
	requestDto.Operation = cemail.SendWelcomeEmailOperation
	requestDto.OccurredAt = time.Now().UTC()
	return enqueue(c, ctx, requestDto.RequestId, requestDto.OccurredAt, requestDto)
}

func (c *Client) SendValidationEmail(
	ctx context.Context,
	requestDto cemailevents.SendValidationEmailRequestDto,
) *cexceptions.Exception {
	requestDto.RequestId = uuid.New()
	requestDto.Operation = cemail.SendValidationEmailOperation
	requestDto.OccurredAt = time.Now().UTC()
	return enqueue(c, ctx, requestDto.RequestId, requestDto.OccurredAt, requestDto)
}

func (c *Client) SendSecurityAlertEmail(
	ctx context.Context,
	requestDto cemailevents.SendSecurityAlertEmailRequestDto,
) *cexceptions.Exception {
	requestDto.RequestId = uuid.New()
	requestDto.Operation = cemail.SendSecurityAlertEmailOperation
	requestDto.OccurredAt = time.Now().UTC()
	return enqueue(c, ctx, requestDto.RequestId, requestDto.OccurredAt, requestDto)
}

func enqueue[D any](
	c *Client,
	ctx context.Context,
	requestID uuid.UUID,
	occurredAt time.Time,
	requestDto D,
) *cexceptions.Exception {
	if c == nil || c.db == nil {
		return cexceptions.New(
			"EmailServiceUnavailable",
			"Email",
			"Publish",
			"The email service producer is unavailable",
			http.StatusServiceUnavailable,
			true,
		)
	}

	envelope := cevent.EventEnvelope[D]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     cemailevents.EventType_EmailRequested,
		AggregateType: cemailevents.AggregateType_EmailRequest,
		AggregateId:   requestID,
		KafkaKey:      requestID.String(),
		OccurredAt:    occurredAt,
		CorrelationId: requestID.String(),
		Data:          requestDto,
	}
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return cexceptions.New(
			"EmailServiceUnavailable",
			"Email",
			"Publish",
			"Failed to start the email event transaction",
			http.StatusServiceUnavailable,
			true,
		).WithOrigin(tx.Error)
	}
	if err := crepositories.EnqueueOutboxEvents(
		tx,
		cemailevents.CoreEmailRequestTopic,
		[]cevent.EventEnvelope[D]{envelope},
	); err != nil {
		tx.Rollback()
		return cexceptions.New(
			"EmailServiceUnavailable",
			"Email",
			"Enqueue",
			"Failed to enqueue the email event",
			http.StatusServiceUnavailable,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return cexceptions.New(
			"EmailServiceUnavailable",
			"Email",
			"Enqueue",
			"Failed to commit the email event",
			http.StatusServiceUnavailable,
			true,
		).WithOrigin(err)
	}

	return nil
}
