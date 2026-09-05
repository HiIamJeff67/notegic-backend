package core

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	validatorpkg "github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	cemail "github.com/HiIamJeff67/notegic-backend/contracts/email/v1"
	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
)

type welcomeBuilderStub struct{}

func (welcomeBuilderStub) Build(cemailevents.SendWelcomeEmailRequestDto) (*cemailevents.SendWelcomeEmailResponseDto, error) {
	return nil, nil
}

type validationBuilderStub struct{}

func (validationBuilderStub) Build(cemailevents.SendValidationEmailRequestDto) (*cemailevents.SendValidationEmailResponseDto, error) {
	return nil, nil
}

type securityAlertBuilderStub struct{}

func (securityAlertBuilderStub) Build(cemailevents.SendSecurityAlertEmailRequestDto) (*cemailevents.SendSecurityAlertEmailResponseDto, error) {
	return nil, nil
}

type queueStub struct {
	err error
}

func (s queueStub) EnqueueWelcomeEmail(*cemailevents.SendWelcomeEmailResponseDto) error {
	return s.err
}

func (s queueStub) EnqueueValidationEmail(*cemailevents.SendValidationEmailResponseDto) error {
	return s.err
}

func (s queueStub) EnqueueSecurityAlertEmail(*cemailevents.SendSecurityAlertEmailResponseDto) error {
	return s.err
}

func TestEmailRequestConsumerMapsLocalErrorClassification(t *testing.T) {
	cases := []struct {
		name      string
		retryable bool
		wantClass skafka.ErrorClassification
	}{
		{
			name:      "retryable delivery error",
			retryable: true,
			wantClass: skafka.ErrorClassification_Transient,
		},
		{
			name:      "non retryable configuration error",
			retryable: false,
			wantClass: skafka.ErrorClassification_PoisonMessage,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			requestId := uuid.New()
			request := cemailevents.SendWelcomeEmailRequestDto{
				RequestId:  requestId,
				Operation:  cemail.SendWelcomeEmailOperation,
				OccurredAt: time.Now().UTC(),
				To:         "user@example.com",
				Pattern: cemailevents.WelcomeEmailPattern{
					UserName: "Notegic User",
					Status:   "active",
				},
			}
			data, err := json.Marshal(request)
			if err != nil {
				t.Fatalf("marshal request: %v", err)
			}

			stubException := cexceptions.New("DeliveryFailed", "Email", "SendEmail", "Failed to deliver the email", 502)
			stubException.Retryable = test.retryable
			consumer := &EmailRequestConsumer{
				welcomeBuilder:       welcomeBuilderStub{},
				validationBuilder:    validationBuilderStub{},
				securityAlertBuilder: securityAlertBuilderStub{},
				queue:                queueStub{err: stubException},
			}
			consumer.validator = validatorpkg.New()
			resultErr := consumer.consume(
				context.Background(),
				skafka.ConsumerRecord{},
				cevent.EventEnvelope[json.RawMessage]{
					SchemaVersion: cevent.Version,
					EventType:     cemailevents.EventType_EmailRequested,
					AggregateType: cemailevents.AggregateType_EmailRequest,
					AggregateId:   requestId,
					Data:          data,
				},
			)

			consumerError, ok := resultErr.(*skafka.ConsumerError)
			if !ok {
				t.Fatalf("error type = %T, want *skafka.ConsumerError", resultErr)
			}
			if consumerError.Classification != test.wantClass {
				t.Fatalf("classification = %q, want %q", consumerError.Classification, test.wantClass)
			}
		})
	}
}
