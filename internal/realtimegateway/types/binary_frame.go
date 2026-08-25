package realtimetypes

import (
	"encoding/binary"
	"errors"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
)

type BinaryFrame struct {
	Version            byte
	Type               BinaryFrameType
	ConnectorChannelId uint32
	Payload            []byte
}

func (f *BinaryFrame) UnmarshalBytes(payload []byte) error {
	*f = BinaryFrame{}

	// [version:1][type:1][connectorChannelId:4 big-endian][raw payload:n]
	if len(payload) < sconstants.RealtimeBinaryFrameHeaderSize {
		return errors.New("invalid realtime binary frame")
	}

	*f = BinaryFrame{
		Version:            payload[0],
		Type:               BinaryFrameType(payload[1]),
		ConnectorChannelId: binary.BigEndian.Uint32(payload[2:6]),
		Payload:            payload[6:],
	}
	if f.ConnectorChannelId == 0 {
		*f = BinaryFrame{}

		return errors.New("invalid realtime binary frame")
	}

	return nil
}

func (f BinaryFrame) MarshalBytes() ([]byte, error) {
	if f.ConnectorChannelId == 0 {
		return nil, errors.New("invalid realtime binary frame")
	}

	payload := make([]byte, sconstants.RealtimeBinaryFrameHeaderSize+len(f.Payload))
	payload[0] = f.Version
	payload[1] = byte(f.Type)
	binary.BigEndian.PutUint32(payload[2:6], f.ConnectorChannelId)
	copy(payload[6:], f.Payload)

	return payload, nil
}
