package websocket

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	crealtimegateway "github.com/HiIamJeff67/notegic-backend/contracts/realtime-gateway/v1"
	cyjsworker "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	realtimeleasecache "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/data/redis/realtimelease"
	realtimetypes "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/types"
)

type fakeWorkerManager struct {
	frameHandler           func(realtimetypes.InternalFrame)
	frames                 []realtimetypes.InternalFrame
	disableAcknowledgement bool
	mutex                  sync.Mutex
}

func (m *fakeWorkerManager) Attach(frame realtimetypes.InternalFrame) bool {
	m.mutex.Lock()
	m.frames = append(m.frames, frame)
	frameHandler := m.frameHandler
	disableAcknowledgement := m.disableAcknowledgement
	m.mutex.Unlock()

	if frameHandler != nil && !disableAcknowledgement {
		frame.Type = realtimetypes.InternalFrameType_Attached
		frame.Payload = nil
		frameHandler(frame)
	}

	return true
}

func generateRealtimeConnectionTicket(userPublicId uuid.UUID, userAgent string) (*string, time.Time, error) {
	userAgentHash := sha256.Sum256([]byte(userAgent))
	claims := sharedtokens.RealtimeConnectionTicketClaims{
		UserAgentHash:           fmt.Sprintf("%x", userAgentHash),
		RealtimeProtocolVersion: sconstants.RealtimeProtocolVersion,
	}
	claims.Subject = userPublicId.String()

	return sharedtokens.GenerateRealtimeConnectionTicket(
		claims,
	)
}

func generateRealtimeBlockPackTicket(
	userPublicId uuid.UUID,
	userAgent string,
	blockPackId uuid.UUID,
	permission realtimetypes.ChannelPermission,
) (*string, time.Time, error) {
	return generateRealtimeBlockPackTicketWithMaximumSubscribers(
		userPublicId,
		userAgent,
		blockPackId,
		permission,
		10,
	)
}

func generateRealtimeBlockPackTicketWithMaximumSubscribers(
	userPublicId uuid.UUID,
	userAgent string,
	blockPackId uuid.UUID,
	permission realtimetypes.ChannelPermission,
	maximumSubscribers int32,
) (*string, time.Time, error) {
	userAgentHash := sha256.Sum256([]byte(userAgent))
	claims := sharedtokens.RealtimeBlockPackTicketClaims{
		UserAgentHash:                    fmt.Sprintf("%x", userAgentHash),
		ChannelType:                      string(realtimetypes.ChannelType_BlockPack),
		ChannelId:                        blockPackId.String(),
		Permission:                       string(permission),
		RealtimeProtocolVersion:          sconstants.RealtimeProtocolVersion,
		SchemaVersion:                    cyjsworker.YjsBlockPackSchemaVersion,
		RoomAdmissionPolicyVersion:       crealtimegateway.BlockPackRoomAdmissionPolicyVersion,
		RoomAdmissionEnforcementStrategy: string(crealtimegateway.RoomAdmissionEnforcementStrategy_RejectNewSubscriber),
		MaximumSubscribers:               maximumSubscribers,
		DocumentQuotaPolicyVersion:       cyjsworker.BlockPackDocumentQuotaPolicyVersion,
		MaximumBlockCount:                1000,
	}
	claims.Subject = userPublicId.String()

	return sharedtokens.GenerateRealtimeBlockPackTicket(
		claims,
	)
}

func (m *fakeWorkerManager) Detach(frame realtimetypes.InternalFrame) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.frames = append(m.frames, frame)
}

func (m *fakeWorkerManager) Forward(frame realtimetypes.InternalFrame) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.frames = append(m.frames, frame)
	if m.frameHandler != nil {
		m.frameHandler(frame)
	}

	return true
}

func (m *fakeWorkerManager) SetFrameHandler(handler func(realtimetypes.InternalFrame)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.frameHandler = handler
}

func newTestRealtimeLeaseStore(t *testing.T) *realtimeleasecache.RealtimeLeaseCacheClient {
	t.Helper()

	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start test redis server: %v", err)
	}
	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		server.Close()
		_ = redisClient.Close()
	})

	clientSet := sredis.NewClientSetFromClients(redisClient)
	cacheStore := realtimeleasecache.NewRealtimeLeaseCacheStore(clientSet)
	return realtimeleasecache.NewRealtimeLeaseCacheClient(cacheStore)
}

func TestGatewayRevokesMatchingBlockPackChannels(t *testing.T) {
	workerManager := &fakeWorkerManager{}
	leaseStore := newTestRealtimeLeaseStore(t)
	userPublicId := uuid.New()
	blockPackId := uuid.New()
	connector := &Connector{
		Id:           uuid.New(),
		UserPublicId: userPublicId,
		channels: map[uint32]realtimetypes.Channel{
			1: {
				Type:       realtimetypes.ChannelType_BlockPack,
				Id:         blockPackId,
				Permission: realtimetypes.ChannelPermission_Read,
			},
		},
		outbound: newOutboundQueue(nil),
	}
	gateway := &WebSocketAdapter{
		workerManager: workerManager,
		leaseStore:    leaseStore,
		connectors: map[uuid.UUID]*Connector{
			connector.Id: connector,
		},
	}

	gateway.revokeBlockPackChannels(realtimeleasecache.BlockPackChannelRevocation{
		EventId:            uuid.New(),
		BlockPackId:        blockPackId,
		TargetUserPublicId: &userPublicId,
		Reason:             coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
	})

	if _, exists := connector.get(1); exists {
		t.Fatal("expected matching BlockPack channel to be detached")
	}

	workerManager.mutex.Lock()
	defer workerManager.mutex.Unlock()
	if len(workerManager.frames) != 1 || workerManager.frames[0].Type != realtimetypes.InternalFrameType_Detach {
		t.Fatalf("expected one detach frame, got %#v", workerManager.frames)
	}
}

func TestGatewaySendsReadyAndPong(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userAgent := "notegic-realtime-test"
	userPublicId := uuid.New()
	configureRealtimeTicketKeys(t)
	connectionTicket, _, exception := generateRealtimeConnectionTicket(userPublicId, userAgent)
	if exception != nil {
		t.Fatalf("failed to generate connection ticket: %v", exception)
	}

	workerManager := &fakeWorkerManager{}
	gateway := &WebSocketAdapter{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(req *http.Request) bool {
				return req.Header.Get("Origin") != ""
			},
		},
		workerManager: workerManager,
		leaseStore:    newTestRealtimeLeaseStore(t),
		connectors:    make(map[uuid.UUID]*Connector),
	}
	workerManager.SetFrameHandler(gateway.handleInternalFrame)

	router := gin.New()
	router.GET("/"+crealtimegateway.RealtimeDevelopmentBaseURL, gateway.Handle)

	server := httptest.NewServer(router)
	defer server.Close()

	connection := dialGateway(t, server.URL, userAgent, *connectionTicket)
	defer connection.Close()

	var ready realtimetypes.ReadyFrame
	if err := connection.ReadJSON(&ready); err != nil {
		t.Fatalf("failed to read ready frame: %v", err)
	}
	if ready.Version != sconstants.RealtimeProtocolVersion ||
		ready.Type != realtimetypes.FrameType_Ready ||
		ready.ConnectionId == "" {
		t.Fatalf("unexpected ready frame: %#v", ready)
	}

	if err := connection.WriteJSON(realtimetypes.ControlFrame{
		Version:   sconstants.RealtimeProtocolVersion,
		Type:      realtimetypes.FrameType_Ping,
		RequestId: "request-1",
	}); err != nil {
		t.Fatalf("failed to write ping frame: %v", err)
	}

	var pong realtimetypes.ControlFrame
	if err := connection.ReadJSON(&pong); err != nil {
		t.Fatalf("failed to read pong frame: %v", err)
	}
	if pong.Version != sconstants.RealtimeProtocolVersion ||
		pong.Type != realtimetypes.FrameType_Pong ||
		pong.RequestId != "request-1" {
		t.Fatalf("unexpected pong frame: %#v", pong)
	}
}

func TestGatewayRejectsConnectionsOutsideRealtimeBetaAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userAgent := "notegic-realtime-test"
	userPublicId := uuid.New()
	configureRealtimeTicketKeys(t)
	connectionTicket, _, exception := generateRealtimeConnectionTicket(userPublicId, userAgent)
	if exception != nil {
		t.Fatalf("failed to generate connection ticket: %v", exception)
	}

	gateway := &WebSocketAdapter{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(req *http.Request) bool {
				return req.Header.Get("Origin") != ""
			},
		},
		realtimeBetaUserPublicIdSet: map[uuid.UUID]bool{uuid.New(): true},
		leaseStore:                  newTestRealtimeLeaseStore(t),
		connectors:                  make(map[uuid.UUID]*Connector),
	}

	router := gin.New()
	router.GET("/"+crealtimegateway.RealtimeDevelopmentBaseURL, gateway.Handle)

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/" + crealtimegateway.RealtimeDevelopmentBaseURL
	connection, response, err := (&websocket.Dialer{
		Subprotocols: []string{*connectionTicket},
	}).Dial(wsURL, http.Header{
		"Origin":     []string{server.URL},
		"User-Agent": []string{userAgent},
	})
	if connection != nil {
		connection.Close()
	}
	if err == nil {
		t.Fatal("expected realtime beta allowlist to reject the connection")
	}
	if response == nil || response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %#v", http.StatusServiceUnavailable, response)
	}
}

func TestGatewayRejectsConnectionsWhenGatewayCapacityIsReached(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userAgent := "notegic-realtime-test"
	configureRealtimeTicketKeys(t)

	firstTicket, _, exception := generateRealtimeConnectionTicket(uuid.New(), userAgent)
	if exception != nil {
		t.Fatalf("failed to generate first connection ticket: %v", exception)
	}
	secondTicket, _, exception := generateRealtimeConnectionTicket(uuid.New(), userAgent)
	if exception != nil {
		t.Fatalf("failed to generate second connection ticket: %v", exception)
	}

	workerManager := &fakeWorkerManager{}
	gateway := &WebSocketAdapter{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(req *http.Request) bool {
				return req.Header.Get("Origin") != ""
			},
		},
		workerManager:     workerManager,
		leaseStore:        newTestRealtimeLeaseStore(t),
		connectors:        make(map[uuid.UUID]*Connector),
		maximumConnectors: 1,
	}
	workerManager.SetFrameHandler(gateway.handleInternalFrame)

	router := gin.New()
	router.GET("/"+crealtimegateway.RealtimeDevelopmentBaseURL, gateway.Handle)

	server := httptest.NewServer(router)
	defer server.Close()

	firstConnection := dialGateway(t, server.URL, userAgent, *firstTicket)
	defer firstConnection.Close()

	assertGatewayConnectionRejected(t, server.URL, userAgent, *secondTicket, http.StatusServiceUnavailable)
}

func TestGatewayRejectsConnectionsWhenUserCapacityIsReached(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userAgent := "notegic-realtime-test"
	userPublicId := uuid.New()
	configureRealtimeTicketKeys(t)
	connectionTicket, _, exception := generateRealtimeConnectionTicket(userPublicId, userAgent)
	if exception != nil {
		t.Fatalf("failed to generate connection ticket: %v", exception)
	}

	workerManager := &fakeWorkerManager{}
	gateway := &WebSocketAdapter{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(req *http.Request) bool {
				return req.Header.Get("Origin") != ""
			},
		},
		workerManager:             workerManager,
		leaseStore:                newTestRealtimeLeaseStore(t),
		connectors:                make(map[uuid.UUID]*Connector),
		maximumConnectionsPerUser: 1,
	}
	workerManager.SetFrameHandler(gateway.handleInternalFrame)

	router := gin.New()
	router.GET("/"+crealtimegateway.RealtimeDevelopmentBaseURL, gateway.Handle)

	server := httptest.NewServer(router)
	defer server.Close()

	firstConnection := dialGateway(t, server.URL, userAgent, *connectionTicket)
	defer firstConnection.Close()

	secondConnectionTicket, _, exception := generateRealtimeConnectionTicket(userPublicId, userAgent)
	if exception != nil {
		t.Fatalf("failed to generate second connection ticket: %v", exception)
	}

	assertGatewayConnectionRejected(t, server.URL, userAgent, *secondConnectionTicket, http.StatusTooManyRequests)
}

func TestGatewayRejectsReplayedConnectionTicket(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userAgent := "notegic-realtime-test"
	configureRealtimeTicketKeys(t)
	connectionTicket, _, exception := generateRealtimeConnectionTicket(uuid.New(), userAgent)
	if exception != nil {
		t.Fatalf("failed to generate connection ticket: %v", exception)
	}

	gateway := &WebSocketAdapter{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(req *http.Request) bool {
				return req.Header.Get("Origin") != ""
			},
		},
		workerManager: &fakeWorkerManager{},
		leaseStore:    newTestRealtimeLeaseStore(t),
		connectors:    make(map[uuid.UUID]*Connector),
	}

	router := gin.New()
	router.GET("/"+crealtimegateway.RealtimeDevelopmentBaseURL, gateway.Handle)

	server := httptest.NewServer(router)
	defer server.Close()

	connection := dialGateway(t, server.URL, userAgent, *connectionTicket)
	defer connection.Close()

	assertGatewayConnectionRejected(t, server.URL, userAgent, *connectionTicket, http.StatusConflict)
}

func TestGatewayRejectsBlockPackSubscriptionWhenRoomCapacityIsReached(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userAgent := "notegic-realtime-test"
	firstUserPublicId := uuid.New()
	secondUserPublicId := uuid.New()
	blockPackId := uuid.New()
	configureRealtimeTicketKeys(t)

	firstConnectionTicket, _, exception := generateRealtimeConnectionTicket(firstUserPublicId, userAgent)
	if exception != nil {
		t.Fatalf("failed to generate first connection ticket: %v", exception)
	}
	secondConnectionTicket, _, exception := generateRealtimeConnectionTicket(secondUserPublicId, userAgent)
	if exception != nil {
		t.Fatalf("failed to generate second connection ticket: %v", exception)
	}
	firstChannelTicket, _, exception := generateRealtimeBlockPackTicketWithMaximumSubscribers(
		firstUserPublicId,
		userAgent,
		blockPackId,
		realtimetypes.ChannelPermission_Write,
		1,
	)
	if exception != nil {
		t.Fatalf("failed to generate first channel ticket: %v", exception)
	}
	secondChannelTicket, _, exception := generateRealtimeBlockPackTicketWithMaximumSubscribers(
		secondUserPublicId,
		userAgent,
		blockPackId,
		realtimetypes.ChannelPermission_Read,
		1,
	)
	if exception != nil {
		t.Fatalf("failed to generate second channel ticket: %v", exception)
	}

	workerManager := &fakeWorkerManager{}
	gateway := &WebSocketAdapter{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(req *http.Request) bool {
				return req.Header.Get("Origin") != ""
			},
		},
		workerManager: workerManager,
		leaseStore:    newTestRealtimeLeaseStore(t),
		connectors:    make(map[uuid.UUID]*Connector),
	}
	workerManager.SetFrameHandler(gateway.handleInternalFrame)

	router := gin.New()
	router.GET("/"+crealtimegateway.RealtimeDevelopmentBaseURL, gateway.Handle)

	server := httptest.NewServer(router)
	defer server.Close()

	firstConnection := dialGateway(t, server.URL, userAgent, *firstConnectionTicket)
	defer firstConnection.Close()
	secondConnection := dialGateway(t, server.URL, userAgent, *secondConnectionTicket)
	defer secondConnection.Close()

	for _, connection := range []*websocket.Conn{firstConnection, secondConnection} {
		var ready realtimetypes.ReadyFrame
		if err := connection.ReadJSON(&ready); err != nil {
			t.Fatalf("failed to read ready frame: %v", err)
		}
	}

	if err := firstConnection.WriteJSON(realtimetypes.SubscribeFrame{
		Version:       sconstants.RealtimeProtocolVersion,
		Type:          realtimetypes.FrameType_Subscribe,
		RequestId:     "subscribe-first",
		ChannelType:   realtimetypes.ChannelType_BlockPack,
		ChannelId:     blockPackId,
		ChannelTicket: *firstChannelTicket,
	}); err != nil {
		t.Fatalf("failed to subscribe first connection: %v", err)
	}

	var subscribed realtimetypes.SubscribedFrame
	if err := firstConnection.ReadJSON(&subscribed); err != nil {
		t.Fatalf("failed to read subscribed frame: %v", err)
	}
	if len(subscribed.Participants) != 1 ||
		subscribed.Participants[0].UserPublicId != firstUserPublicId ||
		subscribed.Participants[0].ConnectionCount != 1 {
		t.Fatalf("unexpected BlockPack presence snapshot: %#v", subscribed.Participants)
	}

	if err := secondConnection.WriteJSON(realtimetypes.SubscribeFrame{
		Version:       sconstants.RealtimeProtocolVersion,
		Type:          realtimetypes.FrameType_Subscribe,
		RequestId:     "subscribe-second",
		ChannelType:   realtimetypes.ChannelType_BlockPack,
		ChannelId:     blockPackId,
		ChannelTicket: *secondChannelTicket,
	}); err != nil {
		t.Fatalf("failed to subscribe second connection: %v", err)
	}

	var errorFrame realtimetypes.ErrorFrame
	if err := secondConnection.ReadJSON(&errorFrame); err != nil {
		t.Fatalf("failed to read room capacity error frame: %v", err)
	}
	if errorFrame.Code != realtimetypes.ErrorCode_RoomConnectionLimitExceeded {
		t.Fatalf("unexpected room capacity error frame: %#v", errorFrame)
	}
	secondChannelTicket, _, exception = generateRealtimeBlockPackTicketWithMaximumSubscribers(
		secondUserPublicId,
		userAgent,
		blockPackId,
		realtimetypes.ChannelPermission_Read,
		1,
	)
	if exception != nil {
		t.Fatalf("failed to generate replacement channel ticket: %v", exception)
	}

	if err := firstConnection.WriteJSON(realtimetypes.UnsubscribeFrame{
		Version:            sconstants.RealtimeProtocolVersion,
		Type:               realtimetypes.FrameType_Unsubscribe,
		RequestId:          "unsubscribe-first",
		ConnectorChannelId: subscribed.ConnectorChannelId,
	}); err != nil {
		t.Fatalf("failed to unsubscribe first connection: %v", err)
	}

	var unsubscribed realtimetypes.UnsubscribedFrame
	if err := firstConnection.ReadJSON(&unsubscribed); err != nil {
		t.Fatalf("failed to read unsubscribed frame: %v", err)
	}

	if err := secondConnection.WriteJSON(realtimetypes.SubscribeFrame{
		Version:       sconstants.RealtimeProtocolVersion,
		Type:          realtimetypes.FrameType_Subscribe,
		RequestId:     "subscribe-second-after-release",
		ChannelType:   realtimetypes.ChannelType_BlockPack,
		ChannelId:     blockPackId,
		ChannelTicket: *secondChannelTicket,
	}); err != nil {
		t.Fatalf("failed to resubscribe second connection: %v", err)
	}

	if err := secondConnection.ReadJSON(&subscribed); err != nil {
		t.Fatalf("expected released room capacity to admit the second connection: %v", err)
	}
}

func TestGatewayMultiplexesAndRelaysBlockPackChannels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userAgent := "notegic-realtime-test"
	userPublicId := uuid.New()
	configureRealtimeTicketKeys(t)
	connectionTicket, _, exception := generateRealtimeConnectionTicket(userPublicId, userAgent)
	if exception != nil {
		t.Fatalf("failed to generate connection ticket: %v", exception)
	}

	workerManager := &fakeWorkerManager{}
	gateway := &WebSocketAdapter{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(req *http.Request) bool {
				return req.Header.Get("Origin") != ""
			},
		},
		workerManager: workerManager,
		leaseStore:    newTestRealtimeLeaseStore(t),
		connectors:    make(map[uuid.UUID]*Connector),
	}
	workerManager.SetFrameHandler(gateway.handleInternalFrame)

	router := gin.New()
	router.GET("/"+crealtimegateway.RealtimeDevelopmentBaseURL, gateway.Handle)

	server := httptest.NewServer(router)
	defer server.Close()

	connection := dialGateway(t, server.URL, userAgent, *connectionTicket)
	defer connection.Close()

	var ready realtimetypes.ReadyFrame
	if err := connection.ReadJSON(&ready); err != nil {
		t.Fatalf("failed to read ready frame: %v", err)
	}
	if !ready.ResubscribeRequired {
		t.Fatalf("expected ready frame to require resubscription: %#v", ready)
	}

	unsupportedChannelId := uuid.New()
	if err := connection.WriteJSON(realtimetypes.SubscribeFrame{
		Version:     sconstants.RealtimeProtocolVersion,
		Type:        realtimetypes.FrameType_Subscribe,
		RequestId:   "subscribe-unsupported",
		ChannelType: realtimetypes.ChannelType("Unsupported"),
		ChannelId:   unsupportedChannelId,
	}); err != nil {
		t.Fatalf("failed to subscribe to unsupported channel type: %v", err)
	}

	var unsupportedChannelTypeError realtimetypes.ErrorFrame
	if err := connection.ReadJSON(&unsupportedChannelTypeError); err != nil {
		t.Fatalf("failed to read unsupported channel type error: %v", err)
	}
	if unsupportedChannelTypeError.Code != realtimetypes.ErrorCode_UnsupportedChannelType ||
		unsupportedChannelTypeError.ChannelId == nil ||
		*unsupportedChannelTypeError.ChannelId != unsupportedChannelId {
		t.Fatalf("unexpected unsupported channel type error: %#v", unsupportedChannelTypeError)
	}

	firstBlockPackId := uuid.New()
	secondBlockPackId := uuid.New()
	channelTickets := make(map[uuid.UUID]string, 2)
	for _, blockPackId := range []uuid.UUID{firstBlockPackId, secondBlockPackId} {
		channelTicket, _, exception := generateRealtimeBlockPackTicket(
			userPublicId,
			userAgent,
			blockPackId,
			realtimetypes.ChannelPermission_Write,
		)
		if exception != nil {
			t.Fatalf("failed to generate channel ticket: %v", exception)
		}

		channelTickets[blockPackId] = *channelTicket
	}

	for _, subscribe := range []realtimetypes.SubscribeFrame{
		{
			Version:       sconstants.RealtimeProtocolVersion,
			Type:          realtimetypes.FrameType_Subscribe,
			RequestId:     "subscribe-first",
			ChannelType:   realtimetypes.ChannelType_BlockPack,
			ChannelId:     firstBlockPackId,
			ChannelTicket: channelTickets[firstBlockPackId],
		},
		{
			Version:       sconstants.RealtimeProtocolVersion,
			Type:          realtimetypes.FrameType_Subscribe,
			RequestId:     "subscribe-second",
			ChannelType:   realtimetypes.ChannelType_BlockPack,
			ChannelId:     secondBlockPackId,
			ChannelTicket: channelTickets[secondBlockPackId],
		},
	} {
		if err := connection.WriteJSON(subscribe); err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}
	}

	var firstSubscribed realtimetypes.SubscribedFrame
	if err := connection.ReadJSON(&firstSubscribed); err != nil {
		t.Fatalf("failed to read first subscribed frame: %v", err)
	}
	var secondSubscribed realtimetypes.SubscribedFrame
	if err := connection.ReadJSON(&secondSubscribed); err != nil {
		t.Fatalf("failed to read second subscribed frame: %v", err)
	}
	if firstSubscribed.ConnectorChannelId == 0 || secondSubscribed.ConnectorChannelId == 0 ||
		firstSubscribed.ConnectorChannelId == secondSubscribed.ConnectorChannelId ||
		firstSubscribed.ChannelType != realtimetypes.ChannelType_BlockPack ||
		secondSubscribed.ChannelType != realtimetypes.ChannelType_BlockPack ||
		firstSubscribed.ChannelId != firstBlockPackId ||
		secondSubscribed.ChannelId != secondBlockPackId ||
		firstSubscribed.DocumentQuotaPolicyVersion != cyjsworker.BlockPackDocumentQuotaPolicyVersion ||
		secondSubscribed.DocumentQuotaPolicyVersion != cyjsworker.BlockPackDocumentQuotaPolicyVersion ||
		firstSubscribed.MaximumBlockCount != 1000 ||
		secondSubscribed.MaximumBlockCount != 1000 {
		t.Fatalf("unexpected subscribed frames: %#v %#v", firstSubscribed, secondSubscribed)
	}
	workerManager.mutex.Lock()
	attachFrames := append([]realtimetypes.InternalFrame(nil), workerManager.frames...)
	workerManager.mutex.Unlock()
	if len(attachFrames) < 2 {
		t.Fatalf("expected two worker attach frames, got %#v", attachFrames)
	}
	for _, attachFrame := range attachFrames[:2] {
		var quotaPolicy cyjsworker.BlockPackQuotaPolicy
		if err := json.Unmarshal(attachFrame.Payload, &quotaPolicy); err != nil ||
			quotaPolicy.Version != cyjsworker.BlockPackDocumentQuotaPolicyVersion ||
			quotaPolicy.MaximumBlockCount != 1000 {
			t.Fatalf("unexpected worker attach quota policy: %#v", attachFrame)
		}
	}

	if err := connection.WriteJSON(realtimetypes.SubscribeFrame{
		Version:       sconstants.RealtimeProtocolVersion,
		Type:          realtimetypes.FrameType_Subscribe,
		RequestId:     "subscribe-first-again",
		ChannelType:   realtimetypes.ChannelType_BlockPack,
		ChannelId:     firstBlockPackId,
		ChannelTicket: channelTickets[firstBlockPackId],
	}); err != nil {
		t.Fatalf("failed to repeat subscribe: %v", err)
	}

	var reusedTicketError realtimetypes.ErrorFrame
	if err := connection.ReadJSON(&reusedTicketError); err != nil {
		t.Fatalf("failed to read reused ticket error: %v", err)
	}
	if reusedTicketError.Code != realtimetypes.ErrorCode_TicketAlreadyUsed {
		t.Fatalf("expected a reused ticket error, got %#v", reusedTicketError)
	}

	binaryPayload, err := realtimetypes.BinaryFrame{
		Version:            byte(sconstants.RealtimeProtocolVersion),
		Type:               realtimetypes.BinaryFrameType_YjsDocument,
		ConnectorChannelId: firstSubscribed.ConnectorChannelId,
		Payload:            []byte{1, 2, 3},
	}.MarshalBytes()
	if err != nil {
		t.Fatalf("failed to marshal binary frame: %v", err)
	}

	if err := connection.WriteMessage(websocket.BinaryMessage, binaryPayload); err != nil {
		t.Fatalf("failed to write binary frame: %v", err)
	}

	messageType, relayedPayload, err := connection.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read relayed binary frame: %v", err)
	}
	if messageType != websocket.BinaryMessage {
		t.Fatalf("expected relayed binary frame, got message type %d", messageType)
	}

	var relayedFrame realtimetypes.BinaryFrame
	if err := relayedFrame.UnmarshalBytes(relayedPayload); err != nil {
		t.Fatalf("failed to unmarshal relayed binary frame: %v", err)
	}
	if relayedFrame.Type != realtimetypes.BinaryFrameType_YjsDocument ||
		relayedFrame.ConnectorChannelId != firstSubscribed.ConnectorChannelId ||
		string(relayedFrame.Payload) != string([]byte{1, 2, 3}) {
		t.Fatalf("unexpected relayed binary frame: %#v", relayedFrame)
	}

	if err := connection.WriteJSON(realtimetypes.UnsubscribeFrame{
		Version:            sconstants.RealtimeProtocolVersion,
		Type:               realtimetypes.FrameType_Unsubscribe,
		RequestId:          "unsubscribe-second",
		ConnectorChannelId: secondSubscribed.ConnectorChannelId,
	}); err != nil {
		t.Fatalf("failed to unsubscribe: %v", err)
	}

	var unsubscribed realtimetypes.UnsubscribedFrame
	if err := connection.ReadJSON(&unsubscribed); err != nil {
		t.Fatalf("failed to read unsubscribed frame: %v", err)
	}
	if unsubscribed.ConnectorChannelId != secondSubscribed.ConnectorChannelId ||
		unsubscribed.ChannelType != realtimetypes.ChannelType_BlockPack ||
		unsubscribed.ChannelId != secondBlockPackId {
		t.Fatalf("unexpected unsubscribed frame: %#v", unsubscribed)
	}
}

func TestGatewayRejectsYjsDocumentUpdatesOnReadOnlyChannels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userAgent := "notegic-realtime-test"
	userPublicId := uuid.New()
	configureRealtimeTicketKeys(t)
	connectionTicket, _, exception := generateRealtimeConnectionTicket(userPublicId, userAgent)
	if exception != nil {
		t.Fatalf("failed to generate connection ticket: %v", exception)
	}

	workerManager := &fakeWorkerManager{}
	gateway := &WebSocketAdapter{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(req *http.Request) bool {
				return req.Header.Get("Origin") != ""
			},
		},
		workerManager: workerManager,
		leaseStore:    newTestRealtimeLeaseStore(t),
		connectors:    make(map[uuid.UUID]*Connector),
	}
	workerManager.SetFrameHandler(gateway.handleInternalFrame)

	router := gin.New()
	router.GET("/"+crealtimegateway.RealtimeDevelopmentBaseURL, gateway.Handle)

	server := httptest.NewServer(router)
	defer server.Close()

	connection := dialGateway(t, server.URL, userAgent, *connectionTicket)
	defer connection.Close()

	var ready realtimetypes.ReadyFrame
	if err := connection.ReadJSON(&ready); err != nil {
		t.Fatalf("failed to read ready frame: %v", err)
	}

	blockPackId := uuid.New()
	channelTicket, _, exception := generateRealtimeBlockPackTicket(
		userPublicId,
		userAgent,
		blockPackId,
		realtimetypes.ChannelPermission_Read,
	)
	if exception != nil {
		t.Fatalf("failed to generate read channel ticket: %v", exception)
	}

	if err := connection.WriteJSON(realtimetypes.SubscribeFrame{
		Version:       sconstants.RealtimeProtocolVersion,
		Type:          realtimetypes.FrameType_Subscribe,
		RequestId:     "subscribe-read",
		ChannelType:   realtimetypes.ChannelType_BlockPack,
		ChannelId:     blockPackId,
		ChannelTicket: *channelTicket,
	}); err != nil {
		t.Fatalf("failed to subscribe to read channel: %v", err)
	}

	var subscribed realtimetypes.SubscribedFrame
	if err := connection.ReadJSON(&subscribed); err != nil {
		t.Fatalf("failed to read subscribed frame: %v", err)
	}

	binaryPayload, err := realtimetypes.BinaryFrame{
		Version:            byte(sconstants.RealtimeProtocolVersion),
		Type:               realtimetypes.BinaryFrameType_YjsDocument,
		ConnectorChannelId: subscribed.ConnectorChannelId,
		Payload:            []byte{1, 2, 3},
	}.MarshalBytes()
	if err != nil {
		t.Fatalf("failed to marshal Yjs document frame: %v", err)
	}

	if err := connection.WriteMessage(websocket.BinaryMessage, binaryPayload); err != nil {
		t.Fatalf("failed to write Yjs document frame: %v", err)
	}

	var permissionError realtimetypes.ErrorFrame
	if err := connection.ReadJSON(&permissionError); err != nil {
		t.Fatalf("failed to read channel permission error: %v", err)
	}
	if permissionError.Code != realtimetypes.ErrorCode_ChannelPermissionDenied ||
		permissionError.ConnectorChannelId != subscribed.ConnectorChannelId ||
		permissionError.ChannelId == nil || *permissionError.ChannelId != blockPackId {
		t.Fatalf("unexpected channel permission error: %#v", permissionError)
	}

	workerManager.mutex.Lock()
	defer workerManager.mutex.Unlock()
	for _, frame := range workerManager.frames {
		if frame.Type == realtimetypes.InternalFrameType_YjsDocument {
			t.Fatalf("read-only Yjs document frame must not reach the worker: %#v", frame)
		}
	}
}

func dialGateway(t *testing.T, serverURL string, userAgent string, connectionTicket string) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/" + crealtimegateway.RealtimeDevelopmentBaseURL
	connection, response, err := (&websocket.Dialer{
		Subprotocols: []string{connectionTicket},
	}).Dial(wsURL, http.Header{
		"Origin":     []string{serverURL},
		"User-Agent": []string{userAgent},
	})
	if err != nil {
		t.Fatalf("failed to connect to gateway: %v", err)
	}
	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected status %d, got %d", http.StatusSwitchingProtocols, response.StatusCode)
	}

	return connection
}

func assertGatewayConnectionRejected(
	t *testing.T,
	serverURL string,
	userAgent string,
	connectionTicket string,
	expectedStatus int,
) {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/" + crealtimegateway.RealtimeDevelopmentBaseURL
	connection, response, err := (&websocket.Dialer{
		Subprotocols: []string{connectionTicket},
	}).Dial(wsURL, http.Header{
		"Origin":     []string{serverURL},
		"User-Agent": []string{userAgent},
	})
	if connection != nil {
		connection.Close()
	}
	if err == nil {
		t.Fatal("expected realtime connection to be rejected")
	}
	if response == nil || response.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %#v", expectedStatus, response)
	}
}

func configureRealtimeTicketKeys(t *testing.T) {
	t.Helper()

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate realtime ticket private key: %v", err)
	}
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("failed to marshal realtime ticket private key: %v", err)
	}
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("failed to marshal realtime ticket public key: %v", err)
	}

	t.Setenv("REALTIME_TICKET_PRIVATE_KEY_BASE64", base64.StdEncoding.EncodeToString(privateKeyBytes))
	t.Setenv("REALTIME_TICKET_PUBLIC_KEY_BASE64", base64.StdEncoding.EncodeToString(publicKeyBytes))
}
