package emaileventscontract

import cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

const (
	CoreEmailRequestTopic  cevent.Topic = "notegic.core.email.request.v1"
	CoreEmailConsumerGroup              = "notegic-email-core-v1"
)

const (
	AggregateType_EmailRequest cevent.AggregateType = "EmailRequest"
	EventType_EmailRequested   cevent.EventType     = "EmailRequested"
)
