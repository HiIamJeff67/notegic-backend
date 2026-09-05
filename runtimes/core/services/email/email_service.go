package email

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	general "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/general"
)

type EmailServiceInterface interface {
	SendWelcomeEmail(ctx context.Context, requestDto cemailevents.SendWelcomeEmailRequestDto) *cexceptions.Exception
	SendValidationEmail(ctx context.Context, requestDto cemailevents.SendValidationEmailRequestDto) *cexceptions.Exception
	SendSecurityAlertEmail(ctx context.Context, requestDto cemailevents.SendSecurityAlertEmailRequestDto) *cexceptions.Exception
}

type EmailService struct {
	db                              *gorm.DB
	welcomeOutboxEventRepository    general.OutboxEventRepositoryInterface[cemailevents.SendWelcomeEmailRequestDto]
	validationOutboxEventRepository general.OutboxEventRepositoryInterface[cemailevents.SendValidationEmailRequestDto]
	securityOutboxEventRepository   general.OutboxEventRepositoryInterface[cemailevents.SendSecurityAlertEmailRequestDto]
}

func NewEmailService(
	db *gorm.DB,
	welcomeOutboxEventRepository general.OutboxEventRepositoryInterface[cemailevents.SendWelcomeEmailRequestDto],
	validationOutboxEventRepository general.OutboxEventRepositoryInterface[cemailevents.SendValidationEmailRequestDto],
	securityOutboxEventRepository general.OutboxEventRepositoryInterface[cemailevents.SendSecurityAlertEmailRequestDto],
) EmailServiceInterface {
	return &EmailService{
		db:                              db,
		welcomeOutboxEventRepository:    welcomeOutboxEventRepository,
		validationOutboxEventRepository: validationOutboxEventRepository,
		securityOutboxEventRepository:   securityOutboxEventRepository,
	}
}

func (s *EmailService) SendWelcomeEmail(
	ctx context.Context,
	requestDto cemailevents.SendWelcomeEmailRequestDto,
) *cexceptions.Exception {
	requestDto.RequestId = uuid.New()
	requestDto.Operation = cemail.SendWelcomeEmailOperation
	requestDto.OccurredAt = time.Now().UTC()
	if s == nil || s.db == nil || s.welcomeOutboxEventRepository == nil {
		return cexceptions.New("EmailServiceUnavailable", "Email", "SendWelcomeEmail", "The email service is unavailable", http.StatusServiceUnavailable, true)
	}
	tx := s.db.WithContext(ctx).Begin()
	if err := s.welcomeOutboxEventRepository.EnqueueOutboxEvents(
		tx,
		cemailevents.CoreEmailRequestTopic,
		[]cevent.EventEnvelope[cemailevents.SendWelcomeEmailRequestDto]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     cemailevents.EventType_EmailRequested,
				AggregateType: cemailevents.AggregateType_EmailRequest,
				AggregateId:   requestDto.RequestId,
				KafkaKey:      requestDto.RequestId.String(),
				OccurredAt:    requestDto.OccurredAt,
				CorrelationId: requestDto.RequestId.String(),
				Data:          requestDto,
			},
		},
	); err != nil {
		tx.Rollback()
		return cexceptions.New("EmailServiceUnavailable", "Email", "SendWelcomeEmail", "Failed to persist the email event", http.StatusServiceUnavailable, true).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return cexceptions.New("EmailServiceUnavailable", "Email", "SendWelcomeEmail", "Failed to commit the email event", http.StatusServiceUnavailable, true).WithOrigin(err)
	}
	return nil
}

func (s *EmailService) SendValidationEmail(
	ctx context.Context,
	requestDto cemailevents.SendValidationEmailRequestDto,
) *cexceptions.Exception {
	requestDto.RequestId = uuid.New()
	requestDto.Operation = cemail.SendValidationEmailOperation
	requestDto.OccurredAt = time.Now().UTC()
	if s == nil || s.db == nil || s.validationOutboxEventRepository == nil {
		return cexceptions.New("EmailServiceUnavailable", "Email", "SendValidationEmail", "The email service is unavailable", http.StatusServiceUnavailable, true)
	}
	tx := s.db.WithContext(ctx).Begin()
	if err := s.validationOutboxEventRepository.EnqueueOutboxEvents(
		tx,
		cemailevents.CoreEmailRequestTopic,
		[]cevent.EventEnvelope[cemailevents.SendValidationEmailRequestDto]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     cemailevents.EventType_EmailRequested,
				AggregateType: cemailevents.AggregateType_EmailRequest,
				AggregateId:   requestDto.RequestId,
				KafkaKey:      requestDto.RequestId.String(),
				OccurredAt:    requestDto.OccurredAt,
				CorrelationId: requestDto.RequestId.String(),
				Data:          requestDto,
			},
		},
	); err != nil {
		tx.Rollback()
		return cexceptions.New("EmailServiceUnavailable", "Email", "SendValidationEmail", "Failed to persist the email event", http.StatusServiceUnavailable, true).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return cexceptions.New("EmailServiceUnavailable", "Email", "SendValidationEmail", "Failed to commit the email event", http.StatusServiceUnavailable, true).WithOrigin(err)
	}
	return nil
}

func (s *EmailService) SendSecurityAlertEmail(
	ctx context.Context,
	requestDto cemailevents.SendSecurityAlertEmailRequestDto,
) *cexceptions.Exception {
	requestDto.RequestId = uuid.New()
	requestDto.Operation = cemail.SendSecurityAlertEmailOperation
	requestDto.OccurredAt = time.Now().UTC()
	if s == nil || s.db == nil || s.securityOutboxEventRepository == nil {
		return cexceptions.New("EmailServiceUnavailable", "Email", "SendSecurityAlertEmail", "The email service is unavailable", http.StatusServiceUnavailable, true)
	}
	tx := s.db.WithContext(ctx).Begin()
	if err := s.securityOutboxEventRepository.EnqueueOutboxEvents(
		tx,
		cemailevents.CoreEmailRequestTopic,
		[]cevent.EventEnvelope[cemailevents.SendSecurityAlertEmailRequestDto]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     cemailevents.EventType_EmailRequested,
				AggregateType: cemailevents.AggregateType_EmailRequest,
				AggregateId:   requestDto.RequestId,
				KafkaKey:      requestDto.RequestId.String(),
				OccurredAt:    requestDto.OccurredAt,
				CorrelationId: requestDto.RequestId.String(),
				Data:          requestDto,
			},
		},
	); err != nil {
		tx.Rollback()
		return cexceptions.New("EmailServiceUnavailable", "Email", "SendSecurityAlertEmail", "Failed to persist the email event", http.StatusServiceUnavailable, true).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return cexceptions.New("EmailServiceUnavailable", "Email", "SendSecurityAlertEmail", "Failed to commit the email event", http.StatusServiceUnavailable, true).WithOrigin(err)
	}
	return nil
}

var _ EmailServiceInterface = (*EmailService)(nil)
