package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	validatorpkg "github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
)

type EmailRequestConsumer struct {
	sender      SenderInterface
	validator   *validatorpkg.Validate
	kafkaConfig skafka.ConsumerConfig
}

func NewEmailRequestConsumer(
	sender SenderInterface,
	validator *validatorpkg.Validate,
	kafkaConfig skafka.ConsumerConfig,
) *EmailRequestConsumer {
	if validator == nil {
		validator = validatorpkg.New()
	}
	return &EmailRequestConsumer{sender: sender, validator: validator, kafkaConfig: kafkaConfig}
}

func (c *EmailRequestConsumer) Start(ctx context.Context) func() {
	consumer, err := skafka.NewConsumer(
		c.kafkaConfig,
		cemailevents.CoreEmailRequestTopic.String(),
	)
	if err != nil {
		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, err, "Failed to create Core email request consumer")
		}
		return func() {}
	}

	workerCtx, cancel := context.WithCancel(ctx)
	go func() {
		if err := consumer.Run(workerCtx, c.consume); err != nil && workerCtx.Err() == nil && slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(workerCtx, err, "Core email request consumer stopped")
		}
	}()

	return func() {
		cancel()
		consumer.Close()
	}
}

func (c *EmailRequestConsumer) consume(
	ctx context.Context,
	_ skafka.ConsumerRecord,
	event cevent.EventEnvelope[json.RawMessage],
) error {
	if event.EventType != cemailevents.EventType_EmailRequested ||
		event.AggregateType != cemailevents.AggregateType_EmailRequest {
		return nil
	}

	var metadata struct {
		RequestId uuid.UUID `json:"requestId"`
		Operation string    `json:"operation"`
	}
	if err := json.Unmarshal(event.Data, &metadata); err != nil {
		return &skafka.ConsumerError{
			Classification: skafka.ErrorClassification_SchemaIncompatible,
			Origin:         fmt.Errorf("decode Core email request: %w", err),
		}
	}
	if metadata.RequestId == uuid.Nil || metadata.RequestId != event.AggregateId {
		return &skafka.ConsumerError{
			Classification: skafka.ErrorClassification_SchemaIncompatible,
			Origin:         fmt.Errorf("Core email request ID does not match the aggregate ID"),
		}
	}

	var err error
	switch metadata.Operation {
	case cemail.SendWelcomeEmailOperation:
		var request cemailevents.SendWelcomeEmailRequestDto
		if err := json.Unmarshal(event.Data, &request); err != nil {
			return invalidEmailRequest(err.Error())
		}
		if request.RequestId != event.AggregateId || request.Operation != metadata.Operation {
			return invalidEmailRequest("welcome request metadata is invalid")
		}
		if err := c.validator.Struct(&request); err != nil {
			return invalidEmailRequest(err.Error())
		}
		err = c.sender.SendWelcomeEmail(ctx, request)
	case cemail.SendValidationEmailOperation:
		var request cemailevents.SendValidationEmailRequestDto
		if err := json.Unmarshal(event.Data, &request); err != nil {
			return invalidEmailRequest(err.Error())
		}
		if request.RequestId != event.AggregateId || request.Operation != metadata.Operation {
			return invalidEmailRequest("validation request metadata is invalid")
		}
		if err := c.validator.Struct(&request); err != nil {
			return invalidEmailRequest(err.Error())
		}
		err = c.sender.SendValidationEmail(ctx, request)
	case cemail.SendSecurityAlertEmailOperation:
		var request cemailevents.SendSecurityAlertEmailRequestDto
		if err := json.Unmarshal(event.Data, &request); err != nil {
			return invalidEmailRequest(err.Error())
		}
		if request.RequestId != event.AggregateId || request.Operation != metadata.Operation {
			return invalidEmailRequest("security alert request metadata is invalid")
		}
		if err := c.validator.Struct(&request); err != nil {
			return invalidEmailRequest(err.Error())
		}
		err = c.sender.SendSecurityAlertEmail(ctx, request)
	default:
		return invalidEmailRequest("unsupported email operation")
	}
	if err != nil {
		classification := skafka.ErrorClassification_Transient
		var emailException *cexceptions.Exception
		if errors.As(err, &emailException) && !emailException.Retryable {
			classification = skafka.ErrorClassification_PoisonMessage
		}
		return &skafka.ConsumerError{
			Classification: classification,
			Origin:         err,
		}
	}

	return nil
}

func invalidEmailRequest(message string) error {
	return &skafka.ConsumerError{
		Classification: skafka.ErrorClassification_SchemaIncompatible,
		Origin:         fmt.Errorf("invalid Core email request: %s", message),
	}
}
