package email

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	emailcontract "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	emaileventscontract "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"
	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	crepositories "github.com/HiIamJeff67/notegic-backend/contracts/types/models/repositories"
)

type ClientInterface interface {
	SendWelcomeEmail(ctx context.Context, requestDto emaileventscontract.SendWelcomeEmailRequestDto) *exceptions.Exception
	SendValidationEmail(ctx context.Context, requestDto emaileventscontract.SendValidationEmailRequestDto) *exceptions.Exception
	SendSecurityAlertEmail(ctx context.Context, requestDto emaileventscontract.SendSecurityAlertEmailRequestDto) *exceptions.Exception
}

type Client struct {
	db *gorm.DB
}

func NewClient(db *gorm.DB) ClientInterface {
	return &Client{db: db}
}

func (c *Client) SendWelcomeEmail(
	ctx context.Context,
	requestDto emaileventscontract.SendWelcomeEmailRequestDto,
) *exceptions.Exception {
	requestDto.RequestId = uuid.New()
	requestDto.Operation = emailcontract.SendWelcomeEmailOperation
	requestDto.OccurredAt = time.Now().UTC()
	return enqueue(c, ctx, requestDto.RequestId, requestDto.OccurredAt, requestDto)
}

func (c *Client) SendValidationEmail(
	ctx context.Context,
	requestDto emaileventscontract.SendValidationEmailRequestDto,
) *exceptions.Exception {
	requestDto.RequestId = uuid.New()
	requestDto.Operation = emailcontract.SendValidationEmailOperation
	requestDto.OccurredAt = time.Now().UTC()
	return enqueue(c, ctx, requestDto.RequestId, requestDto.OccurredAt, requestDto)
}

func (c *Client) SendSecurityAlertEmail(
	ctx context.Context,
	requestDto emaileventscontract.SendSecurityAlertEmailRequestDto,
) *exceptions.Exception {
	requestDto.RequestId = uuid.New()
	requestDto.Operation = emailcontract.SendSecurityAlertEmailOperation
	requestDto.OccurredAt = time.Now().UTC()
	return enqueue(c, ctx, requestDto.RequestId, requestDto.OccurredAt, requestDto)
}

func enqueue[D any](
	c *Client,
	ctx context.Context,
	requestID uuid.UUID,
	occurredAt time.Time,
	requestDto D,
) *exceptions.Exception {
	if c == nil || c.db == nil {
		return exceptions.New(
			"EmailServiceUnavailable",
			"Email",
			"Publish",
			"The email service producer is unavailable",
			http.StatusServiceUnavailable,
			true,
		)
	}

	envelope := eventcontract.EventEnvelope[D]{
		SchemaVersion: eventcontract.Version,
		EventId:       uuid.New(),
		EventType:     emaileventscontract.EventType_EmailRequested,
		AggregateType: emaileventscontract.AggregateType_EmailRequest,
		AggregateId:   requestID,
		KafkaKey:      requestID.String(),
		OccurredAt:    occurredAt,
		CorrelationId: requestID.String(),
		Data:          requestDto,
	}
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return exceptions.New(
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
		emaileventscontract.CoreEmailRequestTopic,
		[]eventcontract.EventEnvelope[D]{envelope},
	); err != nil {
		tx.Rollback()
		return exceptions.New(
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
		return exceptions.New(
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
