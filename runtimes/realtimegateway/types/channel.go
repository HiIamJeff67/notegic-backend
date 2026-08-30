package realtimetypes

import "github.com/google/uuid"

type Channel struct {
	Type                       ChannelType
	Id                         uuid.UUID
	Permission                 ChannelPermission
	SubscribeRequestId         string
	Ready                      bool
	DocumentQuotaPolicyVersion int
	MaximumBlockCount          int32
	AcknowledgedSequence       int64
}
