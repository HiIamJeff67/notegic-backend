package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/attribute"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cyjsworker "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	smetrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	straces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	realtimeconfig "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/configs"
	realtimeleasecache "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/data/redis/realtimelease"
	realtimetypes "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/types"
	workers "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/workers"
)

type WebSocketAdapter struct {
	upgrader                             websocket.Upgrader
	workerManager                        workers.WorkerManagerInterface
	leaseStore                           *realtimeleasecache.RealtimeLeaseCacheClient
	realtimeDisabled                     bool
	realtimeBetaUserPublicIdSet          map[uuid.UUID]bool
	connectorMutex                       sync.RWMutex
	connectors                           map[uuid.UUID]*Connector
	pendingConnectorCount                int
	maximumConnectors                    int
	maximumConnectionsPerUser            int
	shutdownRevocationListener           func()
	shutdownSessionListener              func()
	shutdownPresenceListener             func()
	shutdownResourceListener             func()
	shutdownNotificationListener         func()
	shutdownRoutineTaskLifecycleListener func()
}

func NewWebSocketAdapter(
	config realtimeconfig.Config,
	leaseStore *realtimeleasecache.RealtimeLeaseCacheClient,
) *WebSocketAdapter {
	workerManager := workers.NewWorkerManager(config.YjsWorkerUrls)
	var realtimeBetaUserPublicIdSet map[uuid.UUID]bool
	if len(config.BetaUserPublicIds) > 0 {
		realtimeBetaUserPublicIdSet = make(map[uuid.UUID]bool)
		for _, rawUserPublicId := range config.BetaUserPublicIds {
			userPublicId, err := uuid.Parse(rawUserPublicId)
			if err == nil {
				realtimeBetaUserPublicIdSet[userPublicId] = true
			}
		}
	}

	application := &WebSocketAdapter{
		workerManager:               workerManager,
		leaseStore:                  leaseStore,
		realtimeDisabled:            !config.RealtimeEnabled,
		realtimeBetaUserPublicIdSet: realtimeBetaUserPublicIdSet,
		connectors:                  make(map[uuid.UUID]*Connector),
		maximumConnectors:           sconstants.RealtimeMaxConnectorsPerGateway,
		maximumConnectionsPerUser:   sconstants.RealtimeMaxConnectionsPerUser,
	}
	application.upgrader = websocket.Upgrader{
		CheckOrigin: func(req *http.Request) bool {
			return req.Header.Get("Origin") != ""
		},
	}
	workerManager.SetFrameHandler(application.handleInternalFrame)
	application.subscribeBlockPackChannelRevocations()
	application.subscribeUserSessionRevocations()
	application.subscribeBlockPackPresenceEvents()
	application.subscribeResourceEvents()
	application.subscribeNotifications()
	application.subscribeRoutineTaskLifecycleEvents()

	return application
}

func (g *WebSocketAdapter) subscribeBlockPackChannelRevocations() {
	shutdown, err := g.leaseStore.SubscribeBlockPackChannelRevocations(g.revokeBlockPackChannels)
	if err != nil {
		slogs.NotegicLogger.Error(context.Background(), err, "Failed to subscribe to realtime BlockPack channel revocations")
		return
	}

	g.shutdownRevocationListener = shutdown
}

func (g *WebSocketAdapter) subscribeBlockPackPresenceEvents() {
	shutdown, err := g.leaseStore.SubscribeBlockPackPresenceEvents(g.broadcastBlockPackPresenceEvent)
	if err != nil {
		slogs.NotegicLogger.Error(context.Background(), err, "Failed to subscribe to realtime BlockPack presence events")
		return
	}

	g.shutdownPresenceListener = shutdown
}

func (g *WebSocketAdapter) subscribeUserSessionRevocations() {
	shutdown, err := g.leaseStore.SubscribeUserSessionRevocations(g.revokeUserSessions)
	if err != nil {
		slogs.NotegicLogger.Error(context.Background(), err, "Failed to subscribe to realtime user session revocations")
		return
	}

	g.shutdownSessionListener = shutdown
}

func (g *WebSocketAdapter) subscribeResourceEvents() {
	shutdown, err := g.leaseStore.SubscribeResourceEvents(g.broadcastResourceEvent)
	if err != nil {
		slogs.NotegicLogger.Error(context.Background(), err, "Failed to subscribe to realtime resource events")
		return
	}

	g.shutdownResourceListener = shutdown
}

func (g *WebSocketAdapter) subscribeNotifications() {
	shutdown, err := g.leaseStore.SubscribeNotifications(g.broadcastNotification)
	if err != nil {
		slogs.NotegicLogger.Error(context.Background(), err, "Failed to subscribe to realtime notifications")
		return
	}

	g.shutdownNotificationListener = shutdown
}

func (g *WebSocketAdapter) subscribeRoutineTaskLifecycleEvents() {
	shutdown, err := g.leaseStore.SubscribeRoutineTaskLifecycleEvents(
		g.broadcastRoutineTaskLifecycleEvent,
	)
	if err != nil {
		slogs.NotegicLogger.Error(
			context.Background(),
			err,
			"Failed to subscribe to realtime RoutineTask lifecycle events",
		)
		return
	}

	g.shutdownRoutineTaskLifecycleListener = shutdown
}

func (g *WebSocketAdapter) broadcastNotification(event realtimeleasecache.NotificationEvent) {
	g.connectorMutex.RLock()
	connectors := make([]*Connector, 0)
	for _, connector := range g.connectors {
		if connector.UserPublicId == event.RecipientUserPublicId {
			connectors = append(connectors, connector)
		}
	}
	g.connectorMutex.RUnlock()

	for _, connector := range connectors {
		if err := connector.writeJSON(realtimetypes.NotificationFrame{
			Version:          sconstants.RealtimeProtocolVersion,
			Type:             realtimetypes.FrameType_Notification,
			EventId:          event.EventId,
			NotificationId:   event.NotificationId,
			NotificationType: event.Type,
			Priority:         event.Priority,
			TemplateKey:      event.TemplateKey,
			TemplateVersion:  event.TemplateVersion,
			Payload:          event.Payload,
			CreatedAt:        event.CreatedAt,
			ExpiresAt:        event.ExpiresAt,
		}); err != nil {
			slogs.NotegicLogger.Error(context.Background(), err, "Failed to enqueue realtime notification")
		}
	}
}

func (g *WebSocketAdapter) broadcastRoutineTaskLifecycleEvent(
	event realtimeleasecache.RoutineTaskLifecycleEvent,
) {
	g.connectorMutex.RLock()
	connectors := make([]*Connector, 0)
	for _, connector := range g.connectors {
		if connector.UserPublicId == event.ActorUserPublicId {
			connectors = append(connectors, connector)
		}
	}
	g.connectorMutex.RUnlock()

	for _, connector := range connectors {
		if err := connector.writeJSON(realtimetypes.RoutineTaskLifecycleFrame{
			Version:             sconstants.RealtimeProtocolVersion,
			Type:                realtimetypes.FrameType_RoutineTaskLifecycle,
			EventId:             event.EventId,
			RoutineTaskId:       event.RoutineTaskId,
			RoutineTaskRecordId: event.RoutineTaskRecordId,
			RoutineId:           event.RoutineId,
			Purpose:             event.Purpose,
			Status:              event.Status,
			Attempt:             event.Attempt,
			OccurredAt:          event.OccurredAt,
		}); err != nil {
			slogs.NotegicLogger.Error(
				context.Background(),
				err,
				"Failed to enqueue realtime RoutineTask lifecycle event",
			)
		}
	}
}

func (g *WebSocketAdapter) broadcastResourceEvent(event realtimeleasecache.ResourceEvent) {
	g.connectorMutex.RLock()
	connectors := make([]*Connector, 0, len(g.connectors))
	for _, connector := range g.connectors {
		if event.TargetUserPublicId != nil && connector.UserPublicId != *event.TargetUserPublicId {
			continue
		}
		connectors = append(connectors, connector)
	}
	g.connectorMutex.RUnlock()

	for _, connector := range connectors {
		if event.TargetUserPublicId == nil {
			connector.channelMutex.RLock()
			hasResourceChannel := false
			for _, channel := range connector.channels {
				if channel.Type == realtimetypes.ChannelType_BlockPack && channel.Id == event.ResourceId {
					hasResourceChannel = true
					break
				}
			}
			connector.channelMutex.RUnlock()
			if !hasResourceChannel {
				continue
			}
		}

		if err := connector.writeJSON(realtimetypes.ResourceEventFrame{
			Version:            sconstants.RealtimeProtocolVersion,
			Type:               realtimetypes.FrameType_ResourceEvent,
			EventId:            event.EventId,
			EventType:          event.EventType,
			ResourceId:         event.ResourceId,
			TargetUserPublicId: event.TargetUserPublicId,
			Change:             event.Change,
			Permission:         event.Permission,
		}); err != nil {
			slogs.NotegicLogger.Error(context.Background(), err, "Failed to enqueue realtime resource event")
		}
	}
}

func (g *WebSocketAdapter) Shutdown() {
	if g.shutdownRevocationListener != nil {
		g.shutdownRevocationListener()
	}
	if g.shutdownSessionListener != nil {
		g.shutdownSessionListener()
	}
	if g.shutdownPresenceListener != nil {
		g.shutdownPresenceListener()
	}
	if g.shutdownResourceListener != nil {
		g.shutdownResourceListener()
	}
	if g.shutdownNotificationListener != nil {
		g.shutdownNotificationListener()
	}
	if g.shutdownRoutineTaskLifecycleListener != nil {
		g.shutdownRoutineTaskLifecycleListener()
	}

	g.connectorMutex.RLock()
	connections := make([]*websocket.Conn, 0, len(g.connectors))
	for _, connector := range g.connectors {
		connections = append(connections, connector.connection)
	}
	g.connectorMutex.RUnlock()

	for _, connection := range connections {
		_ = connection.Close()
	}
	if workerManager, ok := g.workerManager.(*workers.WorkerManager); ok {
		workerManager.Shutdown()
	}
}

func (g *WebSocketAdapter) revokeUserSessions(revocation realtimeleasecache.UserSessionRevocation) {
	g.connectorMutex.RLock()
	connections := make([]*websocket.Conn, 0)
	for _, connector := range g.connectors {
		if connector.UserPublicId == revocation.UserPublicId {
			connections = append(connections, connector.connection)
		}
	}
	g.connectorMutex.RUnlock()

	for _, connection := range connections {
		_ = connection.Close()
	}
}

func (g *WebSocketAdapter) revokeBlockPackChannels(revocation realtimeleasecache.BlockPackChannelRevocation) {
	g.connectorMutex.RLock()
	connectors := make([]*Connector, 0, len(g.connectors))
	for _, connector := range g.connectors {
		if revocation.TargetUserPublicId == nil || connector.UserPublicId == *revocation.TargetUserPublicId {
			connectors = append(connectors, connector)
		}
	}
	g.connectorMutex.RUnlock()

	for _, connector := range connectors {
		connector.channelMutex.RLock()
		blockPackIdByConnectorChannelId := make(map[uint32]uuid.UUID, len(connector.channels))
		for connectorChannelId, channel := range connector.channels {
			if channel.Type != realtimetypes.ChannelType_BlockPack {
				continue
			}
			if channel.Id == revocation.BlockPackId {
				blockPackIdByConnectorChannelId[connectorChannelId] = channel.Id
			}
		}
		connector.channelMutex.RUnlock()

		for connectorChannelId, blockPackId := range blockPackIdByConnectorChannelId {
			code := realtimetypes.ErrorCode_PermissionRevoked
			message := "permission for this channel has been revoked"
			outcome := "permission_revoked"
			if revocation.Reason == coreevents.BlockPackAccessRevocationReason_ResourceUnavailable {
				code = realtimetypes.ErrorCode_ResourceUnavailable
				message = "the block pack is no longer available"
				outcome = "resource_unavailable"
			}
			g.detachBlockPackChannel(
				connector,
				connectorChannelId,
				blockPackId,
				code,
				message,
				outcome,
			)
		}
	}
}

func (g *WebSocketAdapter) detachBlockPackChannel(
	connector *Connector,
	connectorChannelId uint32,
	blockPackId uuid.UUID,
	code realtimetypes.ErrorCode,
	message string,
	outcome string,
) {
	channel, exists := connector.unsubscribe(connectorChannelId)
	if !exists || channel.Type != realtimetypes.ChannelType_BlockPack || channel.Id != blockPackId {
		return
	}

	if err := g.releaseBlockPackSubscriber(
		channel.Id,
		fmt.Sprintf("%s:%d", connector.Id, connectorChannelId),
	); err != nil {
		slogs.NotegicLogger.Error(context.Background(), err, "Failed to release realtime BlockPack subscriber lease")
	}

	g.workerManager.Detach(realtimetypes.InternalFrame{
		Version:            byte(sconstants.RealtimeWorkerProtocolVersion),
		Type:               realtimetypes.InternalFrameType_Detach,
		ChannelType:        channel.Type,
		ConnectionId:       connector.Id,
		ConnectorChannelId: connectorChannelId,
		ChannelId:          channel.Id,
	})

	smetrics.NotegicMeter.Count(context.Background(), "realtime.channel.subscription.count", 1,
		attribute.String("action", "detach"),
		attribute.String("channelType", string(channel.Type)),
		attribute.String("outcome", outcome),
	)
	smetrics.NotegicMeter.UpDown(context.Background(), "realtime.channel.count", -1,
		attribute.String("channelType", string(channel.Type)),
		attribute.String("permission", string(channel.Permission)),
	)

	connector.writeError(realtimetypes.ErrorFrame{
		Version:            sconstants.RealtimeProtocolVersion,
		Type:               realtimetypes.FrameType_Error,
		ChannelType:        channel.Type,
		ChannelId:          &channel.Id,
		ConnectorChannelId: connectorChannelId,
		Code:               code,
		Message:            message,
	})
}

func (g *WebSocketAdapter) broadcastBlockPackPresenceEvent(event realtimeleasecache.RealtimeBlockPackPresenceEvent) {
	frameType := realtimetypes.FrameType_PresenceUpdated
	switch event.Type {
	case realtimeleasecache.RealtimeBlockPackPresenceEventType_Joined:
		frameType = realtimetypes.FrameType_PresenceJoined
	case realtimeleasecache.RealtimeBlockPackPresenceEventType_Left:
		frameType = realtimetypes.FrameType_PresenceLeft
	}

	g.connectorMutex.RLock()
	connectors := make([]*Connector, 0, len(g.connectors))
	for _, connector := range g.connectors {
		connectors = append(connectors, connector)
	}
	g.connectorMutex.RUnlock()

	for _, connector := range connectors {
		if connector.Id == event.OriginConnectionId {
			continue
		}

		connector.channelMutex.RLock()
		hasBlockPackChannel := false
		for _, channel := range connector.channels {
			if channel.Type == realtimetypes.ChannelType_BlockPack && channel.Id == event.BlockPackId {
				hasBlockPackChannel = true
				break
			}
		}
		connector.channelMutex.RUnlock()
		if !hasBlockPackChannel {
			continue
		}

		_ = connector.writeJSON(realtimetypes.BlockPackPresenceDeltaFrame{
			Version:     sconstants.RealtimeProtocolVersion,
			Type:        frameType,
			ChannelType: realtimetypes.ChannelType_BlockPack,
			ChannelId:   event.BlockPackId,
			Participant: realtimetypes.BlockPackPresenceParticipant{
				UserPublicId:      event.Participant.UserPublicId,
				ChannelPermission: event.Participant.ChannelPermission,
				ConnectionCount:   event.Participant.ConnectionCount,
			},
		})
	}
}

func (g *WebSocketAdapter) publishBlockPackPresenceEvent(
	blockPackId uuid.UUID,
	originConnectionId uuid.UUID,
	userPublicId uuid.UUID,
	previousParticipants []realtimeleasecache.RealtimeBlockPackParticipant,
	currentParticipants []realtimeleasecache.RealtimeBlockPackParticipant,
) {
	var previousParticipant realtimeleasecache.RealtimeBlockPackParticipant
	var currentParticipant realtimeleasecache.RealtimeBlockPackParticipant
	previousExists := false
	currentExists := false
	for _, participant := range previousParticipants {
		if participant.UserPublicId == userPublicId {
			previousParticipant = participant
			previousExists = true
			break
		}
	}
	for _, participant := range currentParticipants {
		if participant.UserPublicId == userPublicId {
			currentParticipant = participant
			currentExists = true
			break
		}
	}
	if !previousExists && !currentExists {
		return
	}

	event := realtimeleasecache.RealtimeBlockPackPresenceEvent{
		BlockPackId:        blockPackId,
		OriginConnectionId: originConnectionId,
	}
	switch {
	case !previousExists:
		event.Type = realtimeleasecache.RealtimeBlockPackPresenceEventType_Joined
		event.Participant = currentParticipant
	case !currentExists:
		event.Type = realtimeleasecache.RealtimeBlockPackPresenceEventType_Left
		event.Participant = previousParticipant
		event.Participant.ConnectionCount = 0
	case previousParticipant.ChannelPermission != currentParticipant.ChannelPermission ||
		previousParticipant.ConnectionCount != currentParticipant.ConnectionCount:
		event.Type = realtimeleasecache.RealtimeBlockPackPresenceEventType_Updated
		event.Participant = currentParticipant
	default:
		return
	}

	if err := g.leaseStore.PublishBlockPackPresenceEvent(event); err != nil {
		slogs.NotegicLogger.Error(context.Background(), err, "Failed to publish realtime BlockPack presence event")
	}
}

func (g *WebSocketAdapter) releaseBlockPackSubscriber(
	blockPackId uuid.UUID,
	member string,
) error {
	participant, err := g.leaseStore.GetBlockPackParticipantByMember(blockPackId, member)
	if err != nil {
		return err
	}
	if participant == nil {
		return g.leaseStore.ReleaseBlockPackSubscriber(blockPackId, member)
	}

	previousParticipants, err := g.leaseStore.GetBlockPackParticipants(blockPackId)
	if err != nil {
		return err
	}
	if err := g.leaseStore.ReleaseBlockPackSubscriber(blockPackId, member); err != nil {
		return err
	}

	currentParticipants, err := g.leaseStore.GetBlockPackParticipants(blockPackId)
	if err != nil {
		return err
	}
	originConnectionId := uuid.Nil
	memberParts := strings.SplitN(member, ":", 2)
	if len(memberParts) > 0 {
		parsedConnectionId, err := uuid.Parse(memberParts[0])
		if err == nil {
			originConnectionId = parsedConnectionId
		}
	}
	g.publishBlockPackPresenceEvent(
		blockPackId,
		originConnectionId,
		participant.UserPublicId,
		previousParticipants,
		currentParticipants,
	)

	return nil
}

func (g *WebSocketAdapter) Handle(ctx *gin.Context) {
	// extract and validate the ticket which is in Sec-WebSocket-Protocol header
	connectionTicket := websocket.Subprotocols(ctx.Request)
	if len(connectionTicket) != 1 {
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "missing_connection_ticket"),
		)
		slogs.NotegicLogger.Warn(ctx.Request.Context(), fmt.Sprintf("Rejected realtime connection: expected one connection ticket, got %d subprotocols", len(connectionTicket)))
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	connectionClaims, err := sharedtokens.ParseRealtimeConnectionTicket(
		connectionTicket[0],
		ctx.GetHeader("User-Agent"),
	)
	if err != nil {
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "invalid_connection_ticket"),
		)
		slogs.NotegicLogger.Warn(ctx.Request.Context(), fmt.Sprintf("Rejected realtime connection: invalid connection ticket: %v", err))
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userPublicId, err := uuid.Parse(connectionClaims.Subject)
	if err != nil {
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "invalid_user_public_id"),
		)
		slogs.NotegicLogger.Warn(ctx.Request.Context(), fmt.Sprintf("Rejected realtime connection: connection ticket subject is not a user public id: %v", err))
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if g.realtimeDisabled ||
		len(g.realtimeBetaUserPublicIdSet) > 0 && !g.realtimeBetaUserPublicIdSet[userPublicId] {
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "rollout_not_enabled"),
		)
		slogs.NotegicLogger.Info(ctx.Request.Context(), "Rejected realtime connection because rollout is not enabled")
		ctx.AbortWithStatus(http.StatusServiceUnavailable)

		return
	}

	connectorId := uuid.New()

	maximumConnectors := g.maximumConnectors
	if maximumConnectors <= 0 {
		maximumConnectors = sconstants.RealtimeMaxConnectorsPerGateway
	}

	g.connectorMutex.Lock()
	if len(g.connectors)+g.pendingConnectorCount >= maximumConnectors {
		g.connectorMutex.Unlock()
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "gateway_capacity_exceeded"),
		)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)

		return
	}
	g.pendingConnectorCount++
	g.connectorMutex.Unlock()

	pendingConnectorAdmission := true
	defer func() {
		if !pendingConnectorAdmission {
			return
		}

		g.connectorMutex.Lock()
		g.pendingConnectorCount--
		g.connectorMutex.Unlock()
	}()

	maximumConnectionsPerUser := g.maximumConnectionsPerUser
	if maximumConnectionsPerUser <= 0 {
		maximumConnectionsPerUser = sconstants.RealtimeMaxConnectionsPerUser
	}

	acquired, _, err := g.leaseStore.AcquireUserConnection(
		userPublicId,
		connectorId,
		maximumConnectionsPerUser,
	)
	if err != nil {
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "admission_unavailable"),
		)
		slogs.NotegicLogger.Error(ctx.Request.Context(), err, "Failed to acquire realtime user connection lease")
		ctx.AbortWithStatus(http.StatusServiceUnavailable)

		return
	}
	if !acquired {
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "user_connection_limit_exceeded"),
		)
		ctx.Header("Retry-After", strconv.Itoa(int(sconstants.RealtimeLeaseTTL.Seconds())))
		ctx.AbortWithStatus(http.StatusTooManyRequests)

		return
	}
	defer func() {
		if err := g.leaseStore.ReleaseUserConnection(userPublicId, connectorId); err != nil {
			slogs.NotegicLogger.Error(ctx.Request.Context(), err, "Failed to release realtime user connection lease")
		}
	}()
	consumed, err := g.leaseStore.ConsumeTicket(
		connectionClaims.ID,
		connectionClaims.ExpiresAt.Time,
	)
	if err != nil {
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "ticket_consumption_unavailable"),
		)
		slogs.NotegicLogger.Error(ctx.Request.Context(), err, "Failed to consume realtime connection ticket")
		ctx.AbortWithStatus(http.StatusServiceUnavailable)

		return
	}
	if !consumed {
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "connection_ticket_already_used"),
		)
		ctx.AbortWithStatus(http.StatusConflict)

		return
	}
	websocketConnection, err := g.upgrader.Upgrade(
		ctx.Writer,
		ctx.Request,
		http.Header{"Sec-WebSocket-Protocol": []string{connectionTicket[0]}},
	)
	if err != nil {
		smetrics.NotegicMeter.Count(ctx.Request.Context(), "realtime.connection.rejected.count", 1,
			attribute.String("reason", "websocket_upgrade_failed"),
		)
		straces.NotegicTracer.RecordError(ctx.Request.Context(), err)

		return
	}
	defer websocketConnection.Close()

	connectionStart := time.Now()
	connectionContext, connectionSpan := straces.NotegicTracer.Start(
		ctx.Request.Context(), "realtime.connection",
	)
	defer func() { straces.NotegicTracer.End(connectionSpan, nil) }()

	connector := Connector{
		Id:           connectorId,
		UserPublicId: userPublicId,
		UserAgent:    ctx.GetHeader("User-Agent"),
		connection:   websocketConnection,
		channels:     make(map[uint32]realtimetypes.Channel),
		outbound:     newOutboundQueue(websocketConnection),
	}
	connectionSpan.SetAttributes(attribute.String("realtime.connection.id", connector.Id.String()))
	connector.startWriter()
	defer connector.stopWriter()

	g.connectorMutex.Lock()
	g.pendingConnectorCount--
	g.connectors[connector.Id] = &connector
	g.connectorMutex.Unlock()
	pendingConnectorAdmission = false
	smetrics.NotegicMeter.Count(connectionContext, "realtime.connection.accepted.count", 1)
	smetrics.NotegicMeter.UpDown(connectionContext, "realtime.connector.count", 1)

	defer func() {
		g.connectorMutex.Lock()
		delete(g.connectors, connector.Id)
		g.connectorMutex.Unlock()
		smetrics.NotegicMeter.Count(connectionContext, "realtime.connection.closed.count", 1)
		smetrics.NotegicMeter.Duration(connectionContext, "realtime.connection.duration", time.Since(connectionStart))
		smetrics.NotegicMeter.UpDown(connectionContext, "realtime.connector.count", -1)
	}()
	defer func() {
		connector.channelMutex.Lock()
		channels := connector.channels
		connector.channels = make(map[uint32]realtimetypes.Channel)
		connector.channelMutex.Unlock()

		for connectorChannelId, channel := range channels {
			if err := g.releaseBlockPackSubscriber(
				channel.Id,
				fmt.Sprintf("%s:%d", connector.Id, connectorChannelId),
			); err != nil {
				slogs.NotegicLogger.Error(connectionContext, err, "Failed to release realtime BlockPack subscriber lease")
			}

			g.workerManager.Detach(realtimetypes.InternalFrame{
				Version:            byte(sconstants.RealtimeWorkerProtocolVersion),
				Type:               realtimetypes.InternalFrameType_Detach,
				ChannelType:        channel.Type,
				ConnectionId:       connector.Id,
				ConnectorChannelId: connectorChannelId,
				ChannelId:          channel.Id,
			})
			smetrics.NotegicMeter.Count(connectionContext, "realtime.channel.subscription.count", 1,
				attribute.String("action", "detach"),
				attribute.String("channelType", string(channel.Type)),
				attribute.String("outcome", "connection_closed"),
			)
			smetrics.NotegicMeter.UpDown(connectionContext, "realtime.channel.count", -1,
				attribute.String("channelType", string(channel.Type)),
				attribute.String("permission", string(channel.Permission)),
			)
		}
	}()

	websocketConnection.SetReadLimit(sconstants.RealtimeMaxMessageSize)
	websocketConnection.SetReadDeadline(time.Now().Add(sconstants.RealtimePongWait))
	websocketConnection.SetPongHandler(func(string) error {
		return websocketConnection.SetReadDeadline(time.Now().Add(sconstants.RealtimePongWait))
	})

	if err := connector.writeJSON(realtimetypes.ReadyFrame{
		Version:             sconstants.RealtimeProtocolVersion,
		Type:                realtimetypes.FrameType_Ready,
		ConnectionId:        connector.Id.String(),
		ResubscribeRequired: true,
	}); err != nil {
		return
	}

	pingDone := make(chan struct{})
	defer close(pingDone)

	go func() {
		ticker := time.NewTicker(sconstants.RealtimePingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-pingDone:
				return
			case <-ticker.C:
				refreshed, err := g.leaseStore.RefreshUserConnection(connector.UserPublicId, connector.Id)
				if err != nil || !refreshed {
					if err != nil {
						slogs.NotegicLogger.Error(connectionContext, err, "Failed to refresh realtime user connection lease")
					}
					_ = websocketConnection.Close()

					return
				}

				connector.channelMutex.RLock()
				channels := make(map[uint32]realtimetypes.Channel, len(connector.channels))
				for connectorChannelId, channel := range connector.channels {
					channels[connectorChannelId] = channel
				}
				connector.channelMutex.RUnlock()

				for connectorChannelId, channel := range channels {
					refreshed, err := g.leaseStore.RefreshBlockPackSubscriber(
						channel.Id,
						fmt.Sprintf("%s:%d", connector.Id, connectorChannelId),
					)
					if err != nil || !refreshed {
						if err != nil {
							slogs.NotegicLogger.Error(connectionContext, err, "Failed to refresh realtime BlockPack subscriber lease")
						}
						_ = websocketConnection.Close()

						return
					}
				}

				if err := connector.writeControl(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	for {
		messageType, payload, err := websocketConnection.ReadMessage()
		if err != nil {
			return
		}

		switch messageType {
		case websocket.BinaryMessage:
			if !g.handleBinaryFrame(connectionContext, &connector, payload) {
				return
			}
		case websocket.TextMessage:
			if !g.handleControlFrame(connectionContext, &connector, payload) {
				return
			}
		default:
			if !connector.writeError(realtimetypes.ErrorFrame{
				Version: sconstants.RealtimeProtocolVersion,
				Type:    realtimetypes.FrameType_Error,
				Code:    realtimetypes.ErrorCode_UnsupportedMessageType,
				Message: "only text control frames and binary channel frames are supported",
			}) {
				return
			}
		}
	}
}

func (g *WebSocketAdapter) handleBinaryFrame(ctx context.Context, connector *Connector, payload []byte) bool {
	var frame realtimetypes.BinaryFrame
	if err := frame.UnmarshalBytes(payload); err != nil {
		smetrics.NotegicMeter.Count(ctx, "realtime.frame.rejected.count", 1,
			attribute.String("direction", "inbound"),
			attribute.String("reason", "invalid_binary_frame"),
		)
		return connector.writeError(realtimetypes.ErrorFrame{
			Version: sconstants.RealtimeProtocolVersion,
			Type:    realtimetypes.FrameType_Error,
			Code:    realtimetypes.ErrorCode_InvalidBinaryFrame,
			Message: "binary frames must contain a version, type, channelId, and payload",
		})
	}
	if int(frame.Version) != sconstants.RealtimeProtocolVersion {
		smetrics.NotegicMeter.Count(ctx, "realtime.frame.rejected.count", 1,
			attribute.String("direction", "inbound"),
			attribute.String("reason", "unsupported_protocol_version"),
		)
		return connector.writeError(realtimetypes.ErrorFrame{
			Version:            sconstants.RealtimeProtocolVersion,
			Type:               realtimetypes.FrameType_Error,
			ConnectorChannelId: frame.ConnectorChannelId,
			Code:               realtimetypes.ErrorCode_UnsupportedProtocolVersion,
			Message:            "unsupported realtime protocol version",
		})
	}

	channel, exists := connector.get(frame.ConnectorChannelId)

	if !exists {
		smetrics.NotegicMeter.Count(ctx, "realtime.frame.rejected.count", 1,
			attribute.String("direction", "inbound"),
			attribute.String("reason", "channel_not_found"),
		)
		return connector.writeError(realtimetypes.ErrorFrame{
			Version:            sconstants.RealtimeProtocolVersion,
			Type:               realtimetypes.FrameType_Error,
			ConnectorChannelId: frame.ConnectorChannelId,
			Code:               realtimetypes.ErrorCode_ChannelNotFound,
			Message:            "connectorChannelId is not subscribed on this connection",
		})
	}
	if !channel.Ready {
		smetrics.NotegicMeter.Count(ctx, "realtime.frame.rejected.count", 1,
			attribute.String("direction", "inbound"),
			attribute.String("reason", "channel_not_ready"),
			attribute.String("channelType", string(channel.Type)),
		)
		return connector.writeError(realtimetypes.ErrorFrame{
			Version:            sconstants.RealtimeProtocolVersion,
			Type:               realtimetypes.FrameType_Error,
			ChannelType:        channel.Type,
			ChannelId:          &channel.Id,
			ConnectorChannelId: frame.ConnectorChannelId,
			Code:               realtimetypes.ErrorCode_BinaryChannelNotReady,
			Message:            "the yjs worker has not finished attaching this channel",
		})
	}
	if frame.Type != realtimetypes.BinaryFrameType_YjsDocument &&
		frame.Type != realtimetypes.BinaryFrameType_Awareness {
		smetrics.NotegicMeter.Count(ctx, "realtime.frame.rejected.count", 1,
			attribute.String("direction", "inbound"),
			attribute.String("reason", "unsupported_binary_type"),
			attribute.String("channelType", string(channel.Type)),
		)
		return connector.writeError(realtimetypes.ErrorFrame{
			Version:            sconstants.RealtimeProtocolVersion,
			Type:               realtimetypes.FrameType_Error,
			ChannelType:        channel.Type,
			ChannelId:          &channel.Id,
			ConnectorChannelId: frame.ConnectorChannelId,
			Code:               realtimetypes.ErrorCode_UnsupportedBinaryType,
			Message:            "binary frame type is not enabled",
		})
	}
	if frame.Type == realtimetypes.BinaryFrameType_YjsDocument &&
		channel.Permission != realtimetypes.ChannelPermission_Write {
		smetrics.NotegicMeter.Count(ctx, "realtime.frame.rejected.count", 1,
			attribute.String("direction", "inbound"),
			attribute.String("reason", "permission_denied"),
			attribute.String("channelType", string(channel.Type)),
		)
		return connector.writeError(realtimetypes.ErrorFrame{
			Version:            sconstants.RealtimeProtocolVersion,
			Type:               realtimetypes.FrameType_Error,
			ChannelType:        channel.Type,
			ChannelId:          &channel.Id,
			ConnectorChannelId: frame.ConnectorChannelId,
			Code:               realtimetypes.ErrorCode_ChannelPermissionDenied,
			Message:            "channel permission does not allow yjs document updates",
		})
	}
	internalFrameType := realtimetypes.InternalFrameType_YjsDocument
	if frame.Type == realtimetypes.BinaryFrameType_Awareness {
		internalFrameType = realtimetypes.InternalFrameType_Awareness
	}
	smetrics.NotegicMeter.Count(ctx, "realtime.frame.count", 1,
		attribute.String("direction", "inbound"),
		attribute.String("channelType", string(channel.Type)),
		attribute.String("frameType", string(frame.Type)),
	)
	smetrics.NotegicMeter.Bytes(ctx, "realtime.payload.bytes", int64(len(frame.Payload)),
		attribute.String("direction", "inbound"),
		attribute.String("channelType", string(channel.Type)),
		attribute.String("frameType", string(frame.Type)),
	)

	if !g.workerManager.Forward(realtimetypes.InternalFrame{
		Version:            byte(sconstants.RealtimeWorkerProtocolVersion),
		Type:               internalFrameType,
		ChannelType:        channel.Type,
		ConnectionId:       connector.Id,
		ConnectorChannelId: frame.ConnectorChannelId,
		ChannelId:          channel.Id,
		Payload:            frame.Payload,
	}) {
		smetrics.NotegicMeter.Count(ctx, "realtime.frame.rejected.count", 1,
			attribute.String("direction", "inbound"),
			attribute.String("reason", "worker_unavailable"),
			attribute.String("channelType", string(channel.Type)),
		)
		g.detachBlockPackChannel(
			connector,
			frame.ConnectorChannelId,
			channel.Id,
			realtimetypes.ErrorCode_WorkerUnavailable,
			"the yjs worker is unavailable",
			"worker_unavailable",
		)

		return true
	}

	return true
}

func (g *WebSocketAdapter) handleControlFrame(ctx context.Context, connector *Connector, payload []byte) bool {
	var controlFrame realtimetypes.ControlFrame
	if err := json.Unmarshal(payload, &controlFrame); err != nil {
		smetrics.NotegicMeter.Count(ctx, "realtime.frame.rejected.count", 1,
			attribute.String("direction", "inbound"),
			attribute.String("reason", "invalid_control_frame"),
		)
		return connector.writeError(realtimetypes.ErrorFrame{
			Version: sconstants.RealtimeProtocolVersion,
			Type:    realtimetypes.FrameType_Error,
			Code:    realtimetypes.ErrorCode_InvalidControlFrame,
			Message: "control frames must be valid JSON",
		})
	}
	if controlFrame.Version != sconstants.RealtimeProtocolVersion {
		smetrics.NotegicMeter.Count(ctx, "realtime.frame.rejected.count", 1,
			attribute.String("direction", "inbound"),
			attribute.String("reason", "unsupported_protocol_version"),
		)
		return connector.writeError(realtimetypes.ErrorFrame{
			Version:   sconstants.RealtimeProtocolVersion,
			Type:      realtimetypes.FrameType_Error,
			RequestId: controlFrame.RequestId,
			Code:      realtimetypes.ErrorCode_UnsupportedProtocolVersion,
			Message:   "unsupported realtime protocol version",
		})
	}
	smetrics.NotegicMeter.Count(ctx, "realtime.frame.count", 1,
		attribute.String("direction", "inbound"),
		attribute.String("frameType", string(controlFrame.Type)),
	)
	smetrics.NotegicMeter.Bytes(ctx, "realtime.payload.bytes", int64(len(payload)),
		attribute.String("direction", "inbound"),
		attribute.String("frameType", string(controlFrame.Type)),
	)

	switch controlFrame.Type {
	case realtimetypes.FrameType_Ping:
		return connector.writeJSON(realtimetypes.ControlFrame{
			Version:   sconstants.RealtimeProtocolVersion,
			Type:      realtimetypes.FrameType_Pong,
			RequestId: controlFrame.RequestId,
		}) == nil
	case realtimetypes.FrameType_Heartbeat:
		return connector.writeJSON(realtimetypes.HeartbeatFrame{
			Version:      sconstants.RealtimeProtocolVersion,
			Type:         realtimetypes.FrameType_Heartbeat,
			RequestId:    controlFrame.RequestId,
			UnixMilliNow: time.Now().UnixMilli(),
		}) == nil
	case realtimetypes.FrameType_Subscribe:
		var subscribeFrame realtimetypes.SubscribeFrame
		if err := json.Unmarshal(payload, &subscribeFrame); err != nil || subscribeFrame.ChannelId == uuid.Nil {
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("outcome", "invalid_channel_id"),
			)
			return connector.writeError(realtimetypes.ErrorFrame{
				Version:   sconstants.RealtimeProtocolVersion,
				Type:      realtimetypes.FrameType_Error,
				RequestId: controlFrame.RequestId,
				Code:      realtimetypes.ErrorCode_InvalidChannelId,
				Message:   "subscribe requires a valid channelId",
			})
		}

		switch subscribeFrame.ChannelType {
		case realtimetypes.ChannelType_BlockPack:
		default:
			if subscribeFrame.ChannelType == "" {
				smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
					attribute.String("action", "subscribe"),
					attribute.String("outcome", "invalid_channel_type"),
				)
				return connector.writeError(realtimetypes.ErrorFrame{
					Version:   sconstants.RealtimeProtocolVersion,
					Type:      realtimetypes.FrameType_Error,
					RequestId: controlFrame.RequestId,
					ChannelId: &subscribeFrame.ChannelId,
					Code:      realtimetypes.ErrorCode_InvalidChannelType,
					Message:   "subscribe requires a channelType",
				})
			}
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("channelType", string(subscribeFrame.ChannelType)),
				attribute.String("outcome", "unsupported_channel_type"),
			)

			return connector.writeError(realtimetypes.ErrorFrame{
				Version:     sconstants.RealtimeProtocolVersion,
				Type:        realtimetypes.FrameType_Error,
				RequestId:   controlFrame.RequestId,
				ChannelType: subscribeFrame.ChannelType,
				ChannelId:   &subscribeFrame.ChannelId,
				Code:        realtimetypes.ErrorCode_UnsupportedChannelType,
				Message:     "channelType is not enabled",
			})
		}

		channelClaims, err := sharedtokens.ParseRealtimeBlockPackTicket(
			subscribeFrame.ChannelTicket,
			connector.UserAgent,
		)
		if err != nil || channelClaims.Subject != connector.UserPublicId.String() ||
			channelClaims.ChannelType != string(subscribeFrame.ChannelType) ||
			channelClaims.ChannelId != subscribeFrame.ChannelId.String() {
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("channelType", string(subscribeFrame.ChannelType)),
				attribute.String("outcome", "invalid_channel_ticket"),
			)
			return connector.writeError(realtimetypes.ErrorFrame{
				Version:     sconstants.RealtimeProtocolVersion,
				Type:        realtimetypes.FrameType_Error,
				RequestId:   subscribeFrame.RequestId,
				ChannelType: subscribeFrame.ChannelType,
				ChannelId:   &subscribeFrame.ChannelId,
				Code:        realtimetypes.ErrorCode_InvalidChannelTicket,
				Message:     "channel ticket is invalid",
			})
		}
		consumed, err := g.leaseStore.ConsumeTicket(
			channelClaims.ID,
			channelClaims.ExpiresAt.Time,
		)
		if err != nil {
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("channelType", string(subscribeFrame.ChannelType)),
				attribute.String("outcome", "ticket_consumption_unavailable"),
			)
			slogs.NotegicLogger.Error(ctx, err, "Failed to consume realtime BlockPack channel ticket")

			return connector.writeError(realtimetypes.ErrorFrame{
				Version:     sconstants.RealtimeProtocolVersion,
				Type:        realtimetypes.FrameType_Error,
				RequestId:   subscribeFrame.RequestId,
				ChannelType: subscribeFrame.ChannelType,
				ChannelId:   &subscribeFrame.ChannelId,
				Code:        realtimetypes.ErrorCode_RoomAdmissionUnavailable,
				Message:     "realtime ticket validation is temporarily unavailable",
			})
		}
		if !consumed {
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("channelType", string(subscribeFrame.ChannelType)),
				attribute.String("outcome", "channel_ticket_already_used"),
			)

			return connector.writeError(realtimetypes.ErrorFrame{
				Version:     sconstants.RealtimeProtocolVersion,
				Type:        realtimetypes.FrameType_Error,
				RequestId:   subscribeFrame.RequestId,
				ChannelType: subscribeFrame.ChannelType,
				ChannelId:   &subscribeFrame.ChannelId,
				Code:        realtimetypes.ErrorCode_TicketAlreadyUsed,
				Message:     "channel ticket has already been used",
			})
		}

		// create the channel here, so if handleControlFrame of subscribe does not fire first
		// the channel just will not be found by g.connectors.get() methods, and error will be thrown
		channel := realtimetypes.Channel{
			Type:                       subscribeFrame.ChannelType,
			Id:                         subscribeFrame.ChannelId,
			Permission:                 realtimetypes.ChannelPermission(channelClaims.Permission),
			SubscribeRequestId:         subscribeFrame.RequestId,
			DocumentQuotaPolicyVersion: channelClaims.DocumentQuotaPolicyVersion,
			MaximumBlockCount:          channelClaims.MaximumBlockCount,
		}
		connectorChannelId, existing := connector.subscribe(channel)
		if connectorChannelId == 0 {
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("channelType", string(channel.Type)),
				attribute.String("outcome", "channel_limit_exceeded"),
			)
			return connector.writeError(realtimetypes.ErrorFrame{
				Version:     sconstants.RealtimeProtocolVersion,
				Type:        realtimetypes.FrameType_Error,
				RequestId:   subscribeFrame.RequestId,
				ChannelType: subscribeFrame.ChannelType,
				ChannelId:   &subscribeFrame.ChannelId,
				Code:        realtimetypes.ErrorCode_ChannelLimitExceeded,
				Message:     "the connection cannot subscribe to more channels",
			})
		}

		if existing {
			existingChannel, exists := connector.get(connectorChannelId)
			if !exists || !existingChannel.Ready {
				return connector.writeError(realtimetypes.ErrorFrame{
					Version:            sconstants.RealtimeProtocolVersion,
					Type:               realtimetypes.FrameType_Error,
					RequestId:          subscribeFrame.RequestId,
					ChannelType:        channel.Type,
					ChannelId:          &channel.Id,
					ConnectorChannelId: connectorChannelId,
					Code:               realtimetypes.ErrorCode_BinaryChannelNotReady,
					Message:            "the yjs worker has not finished attaching this channel",
				})
			}

			participants, err := g.leaseStore.GetBlockPackParticipants(channel.Id)
			if err != nil {
				return connector.writeError(realtimetypes.ErrorFrame{
					Version:            sconstants.RealtimeProtocolVersion,
					Type:               realtimetypes.FrameType_Error,
					RequestId:          subscribeFrame.RequestId,
					ChannelType:        channel.Type,
					ChannelId:          &channel.Id,
					ConnectorChannelId: connectorChannelId,
					Code:               realtimetypes.ErrorCode_RoomAdmissionUnavailable,
					Message:            "realtime participant presence is temporarily unavailable",
				})
			}
			presenceParticipants := make([]realtimetypes.BlockPackPresenceParticipant, len(participants))
			for index, participant := range participants {
				presenceParticipants[index] = realtimetypes.BlockPackPresenceParticipant{
					UserPublicId:      participant.UserPublicId,
					ChannelPermission: participant.ChannelPermission,
					ConnectionCount:   participant.ConnectionCount,
				}
			}
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("channelType", string(channel.Type)),
				attribute.String("outcome", "existing"),
			)
			return connector.writeJSON(realtimetypes.SubscribedFrame{
				Version:                    sconstants.RealtimeProtocolVersion,
				Type:                       realtimetypes.FrameType_Subscribed,
				RequestId:                  subscribeFrame.RequestId,
				ChannelType:                subscribeFrame.ChannelType,
				ChannelId:                  subscribeFrame.ChannelId,
				ConnectorChannelId:         connectorChannelId,
				Existing:                   true,
				DocumentQuotaPolicyVersion: existingChannel.DocumentQuotaPolicyVersion,
				MaximumBlockCount:          existingChannel.MaximumBlockCount,
				Participants:               presenceParticipants,
			}) == nil
		}

		leaseMember := fmt.Sprintf("%s:%d", connector.Id, connectorChannelId)
		acquired, activeSubscribers, err := g.leaseStore.AcquireBlockPackSubscriber(
			channel.Id,
			leaseMember,
			channelClaims.MaximumSubscribers,
		)
		if err != nil {
			connector.unsubscribe(connectorChannelId)
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("channelType", string(channel.Type)),
				attribute.String("outcome", "admission_unavailable"),
			)
			slogs.NotegicLogger.Error(ctx, err, "Failed to acquire realtime BlockPack subscriber lease")

			return connector.writeError(realtimetypes.ErrorFrame{
				Version:            sconstants.RealtimeProtocolVersion,
				Type:               realtimetypes.FrameType_Error,
				RequestId:          subscribeFrame.RequestId,
				ChannelType:        channel.Type,
				ChannelId:          &channel.Id,
				ConnectorChannelId: connectorChannelId,
				Code:               realtimetypes.ErrorCode_RoomAdmissionUnavailable,
				Message:            "room admission is temporarily unavailable",
			})
		}
		if !acquired {
			connector.unsubscribe(connectorChannelId)

			leaseMembers := make([]string, 0)
			leases, err := g.leaseStore.GetBlockPackSubscriberLeases(channel.Id)
			if err != nil {
				slogs.NotegicLogger.Error(ctx, err, "Failed to inspect realtime BlockPack subscriber leases")
			} else {
				for _, lease := range leases {
					leaseMembers = append(leaseMembers, fmt.Sprintf("%s expiresAt=%s", lease.Member, lease.ExpiresAt.UTC().Format(time.RFC3339Nano)))
				}
			}
			slogs.NotegicLogger.Warn(ctx, "Rejected realtime BlockPack subscription because subscriber limit was reached",
				attribute.String("realtime.block_pack.id", channel.Id.String()),
				attribute.Int("realtime.room.maximum_subscribers", int(channelClaims.MaximumSubscribers)),
				attribute.Int64("realtime.room.active_subscribers", activeSubscribers),
				attribute.StringSlice("realtime.room.subscriber_leases", leaseMembers),
			)

			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("channelType", string(channel.Type)),
				attribute.String("outcome", "room_connection_limit_exceeded"),
			)

			return connector.writeError(realtimetypes.ErrorFrame{
				Version:            sconstants.RealtimeProtocolVersion,
				Type:               realtimetypes.FrameType_Error,
				RequestId:          subscribeFrame.RequestId,
				ChannelType:        channel.Type,
				ChannelId:          &channel.Id,
				ConnectorChannelId: connectorChannelId,
				Code:               realtimetypes.ErrorCode_RoomConnectionLimitExceeded,
				Message:            "the room has reached the active subscriber limit for its plan",
			})
		}
		previousParticipants, err := g.leaseStore.GetBlockPackParticipants(channel.Id)
		if err != nil {
			connector.unsubscribe(connectorChannelId)
			if releaseErr := g.releaseBlockPackSubscriber(channel.Id, leaseMember); releaseErr != nil {
				slogs.NotegicLogger.Error(ctx, releaseErr, "Failed to release realtime BlockPack subscriber lease")
			}

			return connector.writeError(realtimetypes.ErrorFrame{
				Version:            sconstants.RealtimeProtocolVersion,
				Type:               realtimetypes.FrameType_Error,
				RequestId:          subscribeFrame.RequestId,
				ChannelType:        channel.Type,
				ChannelId:          &channel.Id,
				ConnectorChannelId: connectorChannelId,
				Code:               realtimetypes.ErrorCode_RoomAdmissionUnavailable,
				Message:            "realtime participant presence is temporarily unavailable",
			})
		}

		if err := g.leaseStore.SetBlockPackParticipant(
			channel.Id,
			leaseMember,
			connector.UserPublicId,
			string(channel.Permission),
		); err != nil {
			if releaseErr := g.releaseBlockPackSubscriber(channel.Id, leaseMember); releaseErr != nil {
				slogs.NotegicLogger.Error(ctx, releaseErr, "Failed to release realtime BlockPack subscriber lease")
			}
			connector.unsubscribe(connectorChannelId)

			return connector.writeError(realtimetypes.ErrorFrame{
				Version:            sconstants.RealtimeProtocolVersion,
				Type:               realtimetypes.FrameType_Error,
				RequestId:          subscribeFrame.RequestId,
				ChannelType:        channel.Type,
				ChannelId:          &channel.Id,
				ConnectorChannelId: connectorChannelId,
				Code:               realtimetypes.ErrorCode_RoomAdmissionUnavailable,
				Message:            "realtime participant presence is temporarily unavailable",
			})
		}

		quotaPolicyPayload, err := json.Marshal(cyjsworker.BlockPackQuotaPolicy{
			Version:           channel.DocumentQuotaPolicyVersion,
			MaximumBlockCount: channel.MaximumBlockCount,
		})
		if err != nil || !g.workerManager.Attach(realtimetypes.InternalFrame{
			Version:            byte(sconstants.RealtimeWorkerProtocolVersion),
			Type:               realtimetypes.InternalFrameType_Attach,
			ChannelType:        channel.Type,
			ConnectionId:       connector.Id,
			ConnectorChannelId: connectorChannelId,
			ChannelId:          channel.Id,
			Payload:            quotaPolicyPayload,
		}) {
			if releaseErr := g.releaseBlockPackSubscriber(channel.Id, leaseMember); releaseErr != nil {
				slogs.NotegicLogger.Error(ctx, releaseErr, "Failed to release realtime BlockPack subscriber lease")
			}

			connector.unsubscribe(connectorChannelId)
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "subscribe"),
				attribute.String("channelType", string(channel.Type)),
				attribute.String("outcome", "worker_unavailable"),
			)

			return connector.writeError(realtimetypes.ErrorFrame{
				Version:            sconstants.RealtimeProtocolVersion,
				Type:               realtimetypes.FrameType_Error,
				RequestId:          subscribeFrame.RequestId,
				ChannelType:        channel.Type,
				ChannelId:          &channel.Id,
				ConnectorChannelId: connectorChannelId,
				Code:               realtimetypes.ErrorCode_WorkerUnavailable,
				Message:            "the yjs worker is unavailable",
			})
		}

		time.AfterFunc(sconstants.RealtimeWorkerAttachTimeout, func() {
			pendingChannel, exists := connector.get(connectorChannelId)
			if exists && !pendingChannel.Ready {
				g.detachBlockPackChannel(
					connector,
					connectorChannelId,
					pendingChannel.Id,
					realtimetypes.ErrorCode_WorkerUnavailable,
					"the yjs worker did not finish attaching this channel",
					"worker_attach_timeout",
				)
			}
		})

		currentParticipants, err := g.leaseStore.GetBlockPackParticipants(channel.Id)
		if err != nil {
			if releaseErr := g.releaseBlockPackSubscriber(channel.Id, leaseMember); releaseErr != nil {
				slogs.NotegicLogger.Error(ctx, releaseErr, "Failed to release realtime BlockPack subscriber lease")
			}
			connector.unsubscribe(connectorChannelId)
			g.workerManager.Detach(realtimetypes.InternalFrame{
				Version:            byte(sconstants.RealtimeWorkerProtocolVersion),
				Type:               realtimetypes.InternalFrameType_Detach,
				ChannelType:        channel.Type,
				ConnectionId:       connector.Id,
				ConnectorChannelId: connectorChannelId,
				ChannelId:          channel.Id,
			})

			return connector.writeError(realtimetypes.ErrorFrame{
				Version:            sconstants.RealtimeProtocolVersion,
				Type:               realtimetypes.FrameType_Error,
				RequestId:          subscribeFrame.RequestId,
				ChannelType:        channel.Type,
				ChannelId:          &channel.Id,
				ConnectorChannelId: connectorChannelId,
				Code:               realtimetypes.ErrorCode_RoomAdmissionUnavailable,
				Message:            "realtime participant presence is temporarily unavailable",
			})
		}
		smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
			attribute.String("action", "subscribe"),
			attribute.String("channelType", string(channel.Type)),
			attribute.String("outcome", "success"),
		)
		smetrics.NotegicMeter.UpDown(ctx, "realtime.channel.count", 1,
			attribute.String("channelType", string(channel.Type)),
			attribute.String("permission", string(channel.Permission)),
		)

		g.publishBlockPackPresenceEvent(
			channel.Id,
			connector.Id,
			connector.UserPublicId,
			previousParticipants,
			currentParticipants,
		)

		return true
	case realtimetypes.FrameType_Unsubscribe:
		var unsubscribeFrame realtimetypes.UnsubscribeFrame
		if err := json.Unmarshal(payload, &unsubscribeFrame); err != nil || unsubscribeFrame.ConnectorChannelId == 0 {
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "unsubscribe"),
				attribute.String("outcome", "invalid_connector_channel_id"),
			)
			return connector.writeError(realtimetypes.ErrorFrame{
				Version:   sconstants.RealtimeProtocolVersion,
				Type:      realtimetypes.FrameType_Error,
				RequestId: controlFrame.RequestId,
				Code:      realtimetypes.ErrorCode_InvalidConnectorChannelId,
				Message:   "unsubscribe requires a valid connectorChannelId",
			})
		}

		channel, exists := connector.unsubscribe(unsubscribeFrame.ConnectorChannelId)

		if !exists {
			smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
				attribute.String("action", "unsubscribe"),
				attribute.String("outcome", "channel_not_found"),
			)
			return connector.writeError(realtimetypes.ErrorFrame{
				Version:            sconstants.RealtimeProtocolVersion,
				Type:               realtimetypes.FrameType_Error,
				RequestId:          unsubscribeFrame.RequestId,
				ConnectorChannelId: unsubscribeFrame.ConnectorChannelId,
				Code:               realtimetypes.ErrorCode_ChannelNotFound,
				Message:            "connectorChannelId is not subscribed on this connection",
			})
		}
		if err := g.releaseBlockPackSubscriber(
			channel.Id,
			fmt.Sprintf("%s:%d", connector.Id, unsubscribeFrame.ConnectorChannelId),
		); err != nil {
			slogs.NotegicLogger.Error(ctx, err, "Failed to release realtime BlockPack subscriber lease")
		}

		g.workerManager.Detach(realtimetypes.InternalFrame{
			Version:            byte(sconstants.RealtimeWorkerProtocolVersion),
			Type:               realtimetypes.InternalFrameType_Detach,
			ChannelType:        channel.Type,
			ConnectionId:       connector.Id,
			ConnectorChannelId: unsubscribeFrame.ConnectorChannelId,
			ChannelId:          channel.Id,
		})
		smetrics.NotegicMeter.Count(ctx, "realtime.channel.subscription.count", 1,
			attribute.String("action", "unsubscribe"),
			attribute.String("channelType", string(channel.Type)),
			attribute.String("outcome", "success"),
		)
		smetrics.NotegicMeter.UpDown(ctx, "realtime.channel.count", -1,
			attribute.String("channelType", string(channel.Type)),
			attribute.String("permission", string(channel.Permission)),
		)

		return connector.writeJSON(realtimetypes.UnsubscribedFrame{
			Version:            sconstants.RealtimeProtocolVersion,
			Type:               realtimetypes.FrameType_Unsubscribed,
			RequestId:          unsubscribeFrame.RequestId,
			ChannelType:        channel.Type,
			ChannelId:          channel.Id,
			ConnectorChannelId: unsubscribeFrame.ConnectorChannelId,
		}) == nil
	case realtimetypes.FrameType_Acknowledge:
		var acknowledgeFrame realtimetypes.AcknowledgeFrame
		if err := json.Unmarshal(payload, &acknowledgeFrame); err != nil || acknowledgeFrame.ConnectorChannelId == 0 {
			return connector.writeError(realtimetypes.ErrorFrame{
				Version:   sconstants.RealtimeProtocolVersion,
				Type:      realtimetypes.FrameType_Error,
				RequestId: controlFrame.RequestId,
				Code:      realtimetypes.ErrorCode_InvalidConnectorChannelId,
				Message:   "ack requires a valid connectorChannelId",
			})
		}

		exists, acknowledged := connector.acknowledge(
			acknowledgeFrame.ConnectorChannelId,
			acknowledgeFrame.Sequence,
		)

		if !exists {
			return connector.writeError(realtimetypes.ErrorFrame{
				Version:            sconstants.RealtimeProtocolVersion,
				Type:               realtimetypes.FrameType_Error,
				RequestId:          acknowledgeFrame.RequestId,
				ConnectorChannelId: acknowledgeFrame.ConnectorChannelId,
				Code:               realtimetypes.ErrorCode_ChannelNotFound,
				Message:            "connectorChannelId is not subscribed on this connection",
			})
		}
		if !acknowledged {
			return connector.writeError(realtimetypes.ErrorFrame{
				Version:            sconstants.RealtimeProtocolVersion,
				Type:               realtimetypes.FrameType_Error,
				RequestId:          acknowledgeFrame.RequestId,
				ConnectorChannelId: acknowledgeFrame.ConnectorChannelId,
				Code:               realtimetypes.ErrorCode_InvalidAcknowledgement,
				Message:            "ack sequence cannot move backwards",
			})
		}

		return connector.writeJSON(realtimetypes.AcknowledgedFrame{
			Version:            sconstants.RealtimeProtocolVersion,
			Type:               realtimetypes.FrameType_Acknowledged,
			RequestId:          acknowledgeFrame.RequestId,
			ConnectorChannelId: acknowledgeFrame.ConnectorChannelId,
			Sequence:           acknowledgeFrame.Sequence,
		}) == nil
	case realtimetypes.FrameType_Reconnect:
		return connector.writeError(realtimetypes.ErrorFrame{
			Version:   sconstants.RealtimeProtocolVersion,
			Type:      realtimetypes.FrameType_Error,
			RequestId: controlFrame.RequestId,
			Code:      realtimetypes.ErrorCode_ResubscribeRequired,
			Message:   "new connections must resubscribe their channels",
		})
	case realtimetypes.FrameType_Authenticate:
		return connector.writeError(realtimetypes.ErrorFrame{
			Version:   sconstants.RealtimeProtocolVersion,
			Type:      realtimetypes.FrameType_Error,
			RequestId: controlFrame.RequestId,
			Code:      realtimetypes.ErrorCode_AuthenticationManagedByUpgrade,
			Message:   "authenticate with the WebSocket upgrade request",
		})
	default:
		return connector.writeError(realtimetypes.ErrorFrame{
			Version:   sconstants.RealtimeProtocolVersion,
			Type:      realtimetypes.FrameType_Error,
			RequestId: controlFrame.RequestId,
			Code:      realtimetypes.ErrorCode_UnsupportedControlType,
			Message:   "control frame type is not enabled",
		})
	}
}

func (g *WebSocketAdapter) handleInternalFrame(frame realtimetypes.InternalFrame) {
	g.connectorMutex.RLock()
	connector, exists := g.connectors[frame.ConnectionId]
	g.connectorMutex.RUnlock()

	if !exists {
		return
	}

	channel, exists := connector.get(frame.ConnectorChannelId)
	if !exists || channel.Type != frame.ChannelType || channel.Id != frame.ChannelId {
		return
	}
	if frame.Type == realtimetypes.InternalFrameType_Attached {
		var alreadyReady bool
		channel, alreadyReady, exists = connector.markReady(frame.ConnectorChannelId)
		if !exists {
			return
		}
		if alreadyReady {
			return
		}

		participants, err := g.leaseStore.GetBlockPackParticipants(channel.Id)
		if err != nil {
			g.detachBlockPackChannel(
				connector,
				frame.ConnectorChannelId,
				channel.Id,
				realtimetypes.ErrorCode_RoomAdmissionUnavailable,
				"realtime participant presence is temporarily unavailable",
				"participant_presence_unavailable",
			)

			return
		}

		presenceParticipants := make([]realtimetypes.BlockPackPresenceParticipant, len(participants))
		for index, participant := range participants {
			presenceParticipants[index] = realtimetypes.BlockPackPresenceParticipant{
				UserPublicId:      participant.UserPublicId,
				ChannelPermission: participant.ChannelPermission,
				ConnectionCount:   participant.ConnectionCount,
			}
		}
		if err := connector.writeJSON(realtimetypes.SubscribedFrame{
			Version:                    sconstants.RealtimeProtocolVersion,
			Type:                       realtimetypes.FrameType_Subscribed,
			RequestId:                  channel.SubscribeRequestId,
			ChannelType:                channel.Type,
			ChannelId:                  channel.Id,
			ConnectorChannelId:         frame.ConnectorChannelId,
			DocumentQuotaPolicyVersion: channel.DocumentQuotaPolicyVersion,
			MaximumBlockCount:          channel.MaximumBlockCount,
			Participants:               presenceParticipants,
		}); err != nil {
			g.handleChannelBackpressure(connector, channel)
		}

		return
	}

	if frame.Type == realtimetypes.InternalFrameType_ResyncRequired ||
		frame.Type == realtimetypes.InternalFrameType_PermissionRevoked ||
		frame.Type == realtimetypes.InternalFrameType_BlockPackQuotaExceeded {
		slogs.NotegicLogger.Error(
			context.Background(),
			nil,
			"Yjs worker requested realtime channel resync",
			attribute.String("blockPackId", channel.Id.String()),
			attribute.String("connectionId", connector.Id.String()),
			attribute.Int("connectorChannelId", int(frame.ConnectorChannelId)),
			attribute.Int("internalFrameType", int(frame.Type)),
		)
		code := realtimetypes.ErrorCode_ResubscribeRequired
		message := "the yjs worker requires this channel to resubscribe"
		outcome := "resync_required"
		if frame.Type == realtimetypes.InternalFrameType_PermissionRevoked {
			code = realtimetypes.ErrorCode_PermissionRevoked
			message = "permission for this channel has been revoked"
			outcome = "permission_revoked"
		} else if frame.Type == realtimetypes.InternalFrameType_BlockPackQuotaExceeded {
			code = realtimetypes.ErrorCode_BlockPackQuotaExceeded
			message = "the block pack has reached the block limit for its plan"
			outcome = "block_pack_quota_exceeded"
		}
		g.detachBlockPackChannel(
			connector,
			frame.ConnectorChannelId,
			channel.Id,
			code,
			message,
			outcome,
		)

		return
	}

	binaryFrameType := realtimetypes.BinaryFrameType_YjsDocument
	if frame.Type == realtimetypes.InternalFrameType_Awareness {
		binaryFrameType = realtimetypes.BinaryFrameType_Awareness
	} else if frame.Type != realtimetypes.InternalFrameType_YjsDocument {
		return
	}

	if err := connector.writeBinary(realtimetypes.BinaryFrame{
		Version:            byte(sconstants.RealtimeProtocolVersion),
		Type:               binaryFrameType,
		ConnectorChannelId: frame.ConnectorChannelId,
		Payload:            frame.Payload,
	}); err != nil {
		g.handleChannelBackpressure(connector, channel)

		return
	}
	smetrics.NotegicMeter.Count(context.Background(), "realtime.frame.count", 1,
		attribute.String("direction", "outbound"),
		attribute.String("channelType", string(channel.Type)),
		attribute.String("frameType", string(binaryFrameType)),
	)
}

// Backpressure means this connector cannot write this channel's queued frames to its client fast enough.
// Do not silently discard Yjs document updates: losing one would leave the client with an incomplete
// CRDT history. Instead, detach only the congested channel, clear its pending outbound queue, and stop
// worker fanout for that channel. The control error is then sent with priority so the client can
// resubscribe and receive a complete state without disrupting other channels on the same connection.
func (g *WebSocketAdapter) handleChannelBackpressure(
	connector *Connector,
	channel realtimetypes.Channel,
) {
	slogs.NotegicLogger.Error(
		context.Background(),
		nil,
		"Realtime channel outbound backpressure",
		attribute.String("blockPackId", channel.Id.String()),
		attribute.String("connectionId", connector.Id.String()),
		attribute.String("channelType", string(channel.Type)),
	)
	smetrics.NotegicMeter.Count(
		context.Background(),
		"realtime.channel.backpressure.count",
		1,
		attribute.String("channelType", string(channel.Type)),
	)

	connectorChannelId, exists := connector.findChannel(channel.Type, channel.Id)
	if !exists {
		return
	}

	connector.unsubscribe(connectorChannelId)
	if err := g.releaseBlockPackSubscriber(
		channel.Id,
		fmt.Sprintf("%s:%d", connector.Id, connectorChannelId),
	); err != nil {
		slogs.NotegicLogger.Error(context.Background(), err, "Failed to release realtime BlockPack subscriber lease")
	}

	g.workerManager.Detach(realtimetypes.InternalFrame{
		Version:            byte(sconstants.RealtimeWorkerProtocolVersion),
		Type:               realtimetypes.InternalFrameType_Detach,
		ChannelType:        channel.Type,
		ConnectionId:       connector.Id,
		ConnectorChannelId: connectorChannelId,
		ChannelId:          channel.Id,
	})
	smetrics.NotegicMeter.Count(context.Background(), "realtime.channel.subscription.count", 1,
		attribute.String("action", "detach"),
		attribute.String("channelType", string(channel.Type)),
		attribute.String("outcome", "backpressure"),
	)
	smetrics.NotegicMeter.UpDown(context.Background(), "realtime.channel.count", -1,
		attribute.String("channelType", string(channel.Type)),
		attribute.String("permission", string(channel.Permission)),
	)

	if !connector.writeError(realtimetypes.ErrorFrame{
		Version:            sconstants.RealtimeProtocolVersion,
		Type:               realtimetypes.FrameType_Error,
		ChannelType:        channel.Type,
		ChannelId:          &channel.Id,
		ConnectorChannelId: connectorChannelId,
		Code:               realtimetypes.ErrorCode_ChannelBackpressure,
		Message:            "channel outbound queue is full; resubscribe this channel to resync",
	}) {
		_ = connector.connection.Close()
	}
}
