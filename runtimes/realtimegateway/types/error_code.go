package realtimetypes

type ErrorCode string

const (
	ErrorCode_AuthenticationManagedByUpgrade ErrorCode = "authentication_managed_by_upgrade"
	ErrorCode_BinaryChannelNotReady          ErrorCode = "binary_channel_not_ready"
	ErrorCode_BlockPackQuotaExceeded         ErrorCode = "block_pack_quota_exceeded"
	ErrorCode_ChannelLimitExceeded           ErrorCode = "channel_limit_exceeded"
	ErrorCode_ChannelNotFound                ErrorCode = "channel_not_found"
	ErrorCode_ChannelPermissionDenied        ErrorCode = "channel_permission_denied"
	ErrorCode_ChannelBackpressure            ErrorCode = "channel_backpressure"
	ErrorCode_InvalidAcknowledgement         ErrorCode = "invalid_acknowledgement"
	ErrorCode_InvalidBinaryFrame             ErrorCode = "invalid_binary_frame"
	ErrorCode_InvalidChannelId               ErrorCode = "invalid_channel_id"
	ErrorCode_InvalidChannelTicket           ErrorCode = "invalid_channel_ticket"
	ErrorCode_InvalidChannelType             ErrorCode = "invalid_channel_type"
	ErrorCode_InvalidConnectorChannelId      ErrorCode = "invalid_connector_channel_id"
	ErrorCode_InvalidControlFrame            ErrorCode = "invalid_control_frame"
	ErrorCode_PermissionRevoked              ErrorCode = "permission_revoked"
	ErrorCode_ResourceUnavailable            ErrorCode = "resource_unavailable"
	ErrorCode_RoomAdmissionUnavailable       ErrorCode = "room_admission_unavailable"
	ErrorCode_RoomConnectionLimitExceeded    ErrorCode = "room_connection_limit_exceeded"
	ErrorCode_ResubscribeRequired            ErrorCode = "resubscribe_required"
	ErrorCode_TicketAlreadyUsed              ErrorCode = "ticket_already_used"
	ErrorCode_UnsupportedBinaryType          ErrorCode = "unsupported_binary_type"
	ErrorCode_UnsupportedChannelType         ErrorCode = "unsupported_channel_type"
	ErrorCode_UnsupportedControlType         ErrorCode = "unsupported_control_type"
	ErrorCode_UnsupportedMessageType         ErrorCode = "unsupported_message_type"
	ErrorCode_UnsupportedProtocolVersion     ErrorCode = "unsupported_protocol_version"
	ErrorCode_WorkerUnavailable              ErrorCode = "worker_unavailable"
)
