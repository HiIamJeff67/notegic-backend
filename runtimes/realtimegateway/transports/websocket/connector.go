package websocket

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	realtimetypes "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/types"
)

type Connector struct {
	Id            uuid.UUID
	UserPublicId  uuid.UUID
	UserAgent     string
	connection    *websocket.Conn
	channels      map[uint32]realtimetypes.Channel
	nextChannelId uint32
	channelMutex  sync.RWMutex
	outbound      *outboundQueue
}

func (c *Connector) get(connectorChannelId uint32) (realtimetypes.Channel, bool) {
	c.channelMutex.RLock()
	defer c.channelMutex.RUnlock()

	channel, exists := c.channels[connectorChannelId]

	return channel, exists
}

func (c *Connector) findChannel(
	channelType realtimetypes.ChannelType,
	channelId uuid.UUID,
) (uint32, bool) {
	c.channelMutex.RLock()
	defer c.channelMutex.RUnlock()

	for connectorChannelId, channel := range c.channels {
		if channel.Type == channelType && channel.Id == channelId {
			return connectorChannelId, true
		}
	}

	return 0, false
}

func (c *Connector) subscribe(channel realtimetypes.Channel) (uint32, bool) {
	c.channelMutex.Lock()
	defer c.channelMutex.Unlock()

	for connectorChannelId, existingChannel := range c.channels {
		if existingChannel.Type == channel.Type && existingChannel.Id == channel.Id {
			return connectorChannelId, true
		}
	}
	if len(c.channels) >= sconstants.RealtimeMaxChannelsPerConnection || c.nextChannelId == sconstants.MAX_UINT32 {
		return 0, false
	}

	c.nextChannelId++
	c.channels[c.nextChannelId] = channel

	return c.nextChannelId, false
}

func (c *Connector) unsubscribe(connectorChannelId uint32) (realtimetypes.Channel, bool) {
	c.channelMutex.Lock()
	channel, exists := c.channels[connectorChannelId]
	if exists {
		delete(c.channels, connectorChannelId)
	}
	c.channelMutex.Unlock()

	if exists {
		c.outbound.clearChannel(connectorChannelId)
	}

	return channel, exists
}

func (c *Connector) markReady(connectorChannelId uint32) (realtimetypes.Channel, bool, bool) {
	c.channelMutex.Lock()
	defer c.channelMutex.Unlock()

	channel, exists := c.channels[connectorChannelId]
	if !exists {
		return realtimetypes.Channel{}, false, false
	}

	alreadyReady := channel.Ready
	channel.Ready = true
	c.channels[connectorChannelId] = channel

	return channel, alreadyReady, true
}

func (c *Connector) acknowledge(connectorChannelId uint32, sequence int64) (bool, bool) {
	c.channelMutex.Lock()
	defer c.channelMutex.Unlock()

	channel, exists := c.channels[connectorChannelId]
	if !exists {
		return false, false
	}
	if sequence < channel.AcknowledgedSequence {
		return true, false
	}

	channel.AcknowledgedSequence = sequence
	c.channels[connectorChannelId] = channel

	return true, true
}

func (c *Connector) writeError(frame realtimetypes.ErrorFrame) bool {
	return c.writeJSON(frame) == nil
}

func (c *Connector) startWriter() {
	c.outbound.startWriter()
}

func (c *Connector) stopWriter() {
	c.outbound.stopWriter()
}

func (c *Connector) writeJSON(frame any) error {
	return c.outbound.writeJSON(frame)
}

func (c *Connector) writeControl(messageType int, payload []byte) error {
	return c.outbound.writeControl(messageType, payload)
}

func (c *Connector) writeBinary(frame realtimetypes.BinaryFrame) error {
	return c.outbound.writeBinary(frame)
}
