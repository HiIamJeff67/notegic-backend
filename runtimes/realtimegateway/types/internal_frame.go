package realtimetypes

import (
	"encoding/binary"
	"errors"

	"github.com/google/uuid"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
)

type InternalFrame struct {
	Version            byte
	Type               InternalFrameType
	ChannelType        ChannelType
	ConnectionId       uuid.UUID
	ConnectorChannelId uint32
	ChannelId          uuid.UUID
	Payload            []byte
}

func (f *InternalFrame) UnmarshalBytes(payload []byte) error {
	*f = InternalFrame{}

	// [version:1][type:1][channelType:1][connectionId:16][connectorChannelId:4][channelId:16][raw payload:n]
	if len(payload) < sconstants.RealtimeInternalFrameHeaderSize {
		return errors.New("invalid realtime internal frame")
	}

	*f = InternalFrame{
		Version:            payload[0],
		Type:               InternalFrameType(payload[1]),
		ConnectorChannelId: binary.BigEndian.Uint32(payload[19:23]),
		Payload:            payload[39:],
	}
	switch InternalChannelType(payload[2]) {
	case InternalChannelType_BlockPack:
		f.ChannelType = ChannelType_BlockPack
	default:
		*f = InternalFrame{}

		return errors.New("invalid realtime internal frame")
	}

	copy(f.ConnectionId[:], payload[3:19])
	copy(f.ChannelId[:], payload[23:39])
	if f.ConnectionId == uuid.Nil || f.ConnectorChannelId == 0 || f.ChannelId == uuid.Nil {
		*f = InternalFrame{}

		return errors.New("invalid realtime internal frame")
	}

	return nil
}

func (f InternalFrame) MarshalBytes() ([]byte, error) {
	var internalChannelType InternalChannelType
	switch f.ChannelType {
	case ChannelType_BlockPack:
		internalChannelType = InternalChannelType_BlockPack
	default:
		return nil, errors.New("invalid realtime internal frame")
	}
	if f.ConnectionId == uuid.Nil || f.ConnectorChannelId == 0 || f.ChannelId == uuid.Nil {
		return nil, errors.New("invalid realtime internal frame")
	}

	payload := make([]byte, sconstants.RealtimeInternalFrameHeaderSize+len(f.Payload))
	payload[0] = f.Version
	payload[1] = byte(f.Type)
	payload[2] = byte(internalChannelType)
	copy(payload[3:19], f.ConnectionId[:])
	binary.BigEndian.PutUint32(payload[19:23], f.ConnectorChannelId)
	copy(payload[23:39], f.ChannelId[:])
	copy(payload[39:], f.Payload)

	return payload, nil
}
