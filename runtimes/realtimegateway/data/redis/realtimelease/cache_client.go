package realtimelease

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	redisscripts "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/data/redis/realtimelease/scripts"
)

type RealtimeBlockPackSubscriberLease struct {
	Member    string
	ExpiresAt time.Time
}

type RealtimeBlockPackParticipant struct {
	UserPublicId      uuid.UUID `json:"userPublicId"`
	ChannelPermission string    `json:"channelPermission"`
	ConnectionCount   int       `json:"connectionCount"`
}

type RealtimeBlockPackPresenceEventType string

const (
	RealtimeBlockPackPresenceEventType_Joined  RealtimeBlockPackPresenceEventType = "joined"
	RealtimeBlockPackPresenceEventType_Left    RealtimeBlockPackPresenceEventType = "left"
	RealtimeBlockPackPresenceEventType_Updated RealtimeBlockPackPresenceEventType = "updated"
)

type RealtimeBlockPackPresenceEvent struct {
	Type               RealtimeBlockPackPresenceEventType `json:"type"`
	BlockPackId        uuid.UUID                          `json:"blockPackId"`
	OriginConnectionId uuid.UUID                          `json:"originConnectionId"`
	Participant        RealtimeBlockPackParticipant       `json:"participant"`
}

type BlockPackChannelRevocation struct {
	EventId            uuid.UUID                                  `json:"eventId"`
	BlockPackId        uuid.UUID                                  `json:"blockPackId"`
	TargetUserPublicId *uuid.UUID                                 `json:"targetUserPublicId,omitempty"`
	Reason             coreevents.BlockPackAccessRevocationReason `json:"reason"`
}

type UserSessionRevocation struct {
	EventId      uuid.UUID `json:"eventId"`
	UserPublicId uuid.UUID `json:"userPublicId"`
}

type ResourceEvent struct {
	EventId            uuid.UUID  `json:"eventId"`
	EventType          string     `json:"eventType"`
	ResourceId         uuid.UUID  `json:"resourceId"`
	TargetUserPublicId *uuid.UUID `json:"targetUserPublicId,omitempty"`
	Change             string     `json:"change"`
	Permission         string     `json:"permission,omitempty"`
}

type NotificationEvent struct {
	EventId               uuid.UUID       `json:"eventId"`
	NotificationId        uuid.UUID       `json:"notificationId"`
	RecipientUserPublicId uuid.UUID       `json:"recipientUserPublicId"`
	Type                  string          `json:"type"`
	Priority              string          `json:"priority"`
	TemplateKey           string          `json:"templateKey"`
	TemplateVersion       int             `json:"templateVersion"`
	Payload               json.RawMessage `json:"payload"`
	CreatedAt             time.Time       `json:"createdAt"`
	ExpiresAt             *time.Time      `json:"expiresAt,omitempty"`
}

type RoutineTaskLifecycleEvent struct {
	EventId             uuid.UUID `json:"eventId"`
	RoutineTaskId       uuid.UUID `json:"routineTaskId"`
	RoutineTaskRecordId uuid.UUID `json:"routineTaskRecordId"`
	RoutineId           uuid.UUID `json:"routineId"`
	ActorUserPublicId   uuid.UUID `json:"actorUserPublicId"`
	Purpose             string    `json:"purpose"`
	Status              string    `json:"status"`
	Attempt             int32     `json:"attempt"`
	OccurredAt          time.Time `json:"occurredAt"`
}

type RealtimeLeaseCacheClient struct {
	cacheStore *RealtimeLeaseCacheStore
}

/* ============================== Constructor ============================== */

func NewRealtimeLeaseCacheClient(cacheStore *RealtimeLeaseCacheStore) *RealtimeLeaseCacheClient {
	return &RealtimeLeaseCacheClient{
		cacheStore: cacheStore,
	}
}

/* ============================== Auxiliary Methods ============================== */

func (s *RealtimeLeaseCacheClient) getRedisClient(identifier string) (*redis.Client, error) {
	if s == nil || s.cacheStore == nil {
		return nil, errors.New("realtime redis lease store is unavailable")
	}

	redisClient, _, err := s.cacheStore.ClientSet().ClientForKey(identifier)
	if err != nil {
		return nil, errors.New("realtime redis lease store is unavailable")
	}

	return redisClient, nil
}

func (s *RealtimeLeaseCacheClient) acquire(identifier string, key string, member string, maximumMembers int) (bool, int64, error) {
	redisClient, err := s.getRedisClient(identifier)
	if err != nil {
		return false, 0, err
	}

	if maximumMembers <= 0 {
		return false, 0, nil
	}

	now := time.Now()
	result, err := redisscripts.AcquireRealtimeLease.Eval(
		redisClient,
		[]string{key},
		now.UnixMilli(),
		now.Add(sconstants.RealtimeLeaseTTL).UnixMilli(),
		maximumMembers,
		sconstants.RealtimeLeaseTTL.Milliseconds(),
		member,
	).Result()
	if err != nil {
		return false, 0, err
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		return false, 0, errors.New("realtime redis lease acquisition returned an invalid result")
	}

	acquired, ok := values[0].(int64)
	if !ok {
		return false, 0, errors.New("realtime redis lease acquisition returned an invalid status")
	}
	activeMembers, ok := values[1].(int64)
	if !ok {
		return false, 0, errors.New("realtime redis lease acquisition returned an invalid member count")
	}

	return acquired == 1, activeMembers, nil
}

func (s *RealtimeLeaseCacheClient) refresh(identifier string, key string, member string) (bool, error) {
	redisClient, err := s.getRedisClient(identifier)
	if err != nil {
		return false, err
	}

	now := time.Now()
	result, err := redisscripts.RefreshRealtimeLease.Eval(
		redisClient,
		[]string{key},
		now.UnixMilli(),
		now.Add(sconstants.RealtimeLeaseTTL).UnixMilli(),
		sconstants.RealtimeLeaseTTL.Milliseconds(),
		member,
	).Result()
	if err != nil {
		return false, err
	}

	refreshed, ok := result.(int64)
	if !ok {
		return false, errors.New("realtime redis lease refresh returned an invalid result")
	}

	return refreshed == 1, nil
}

func (s *RealtimeLeaseCacheClient) blockPackParticipantKey(blockPackId uuid.UUID) string {
	return fmt.Sprintf("Realtime:blockPack:%s:participants", blockPackId)
}

func (s *RealtimeLeaseCacheClient) blockPackChannelRevocationKey() string {
	return "Realtime:blockPack:channel-revocations"
}

func (s *RealtimeLeaseCacheClient) blockPackPresenceKey() string {
	return "Realtime:blockPack:presence-events"
}

func (s *RealtimeLeaseCacheClient) userSessionRevocationKey() string {
	return "Realtime:user:session-revocations"
}

func (s *RealtimeLeaseCacheClient) userNotificationKey() string {
	return "Realtime:user:notifications"
}

func (s *RealtimeLeaseCacheClient) routineTaskLifecycleKey() string {
	return "Realtime:routine-task:lifecycle-events"
}

/* ============================== Lifecycle Methods ============================== */

func (s *RealtimeLeaseCacheClient) MarkLifecycleEventProcessed(eventId uuid.UUID) (bool, error) {
	if eventId == uuid.Nil {
		return false, fmt.Errorf("realtime lifecycle event ID is required")
	}

	redisClient, err := s.getRedisClient("lifecycle-events")
	if err != nil {
		return false, err
	}

	return redisClient.SetNX(
		fmt.Sprintf("Realtime:lifecycle:event:%s", eventId),
		"1",
		7*24*time.Hour,
	).Result()
}

func (s *RealtimeLeaseCacheClient) ReleaseLifecycleEvent(eventId uuid.UUID) error {
	redisClient, err := s.getRedisClient("lifecycle-events")
	if err != nil {
		return err
	}

	return redisClient.Del(fmt.Sprintf("Realtime:lifecycle:event:%s", eventId)).Err()
}

func (s *RealtimeLeaseCacheClient) PublishBlockPackChannelRevocation(
	revocation BlockPackChannelRevocation,
) error {
	redisClient, err := s.getRedisClient(revocation.BlockPackId.String())
	if err != nil {
		return err
	}

	payload, err := json.Marshal(revocation)
	if err != nil {
		return err
	}

	return redisClient.Publish(s.blockPackChannelRevocationKey(), payload).Err()
}

func (s *RealtimeLeaseCacheClient) SubscribeBlockPackChannelRevocations(
	handler func(BlockPackChannelRevocation),
) (func(), error) {
	redisClient, err := s.getRedisClient(s.blockPackChannelRevocationKey())
	if err != nil {
		return nil, err
	}

	pubsub := redisClient.Subscribe(s.blockPackChannelRevocationKey())
	if _, err := pubsub.Receive(); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	go func() {
		for message := range pubsub.Channel() {
			var revocation BlockPackChannelRevocation
			if err := json.Unmarshal([]byte(message.Payload), &revocation); err != nil {
				continue
			}

			handler(revocation)
		}
	}()

	return func() {
		_ = pubsub.Close()
	}, nil
}

func (s *RealtimeLeaseCacheClient) PublishUserSessionRevocation(
	revocation UserSessionRevocation,
) error {
	redisClient, err := s.getRedisClient(revocation.UserPublicId.String())
	if err != nil {
		return err
	}

	payload, err := json.Marshal(revocation)
	if err != nil {
		return err
	}

	return redisClient.Publish(s.userSessionRevocationKey(), payload).Err()
}

func (s *RealtimeLeaseCacheClient) SubscribeUserSessionRevocations(
	handler func(UserSessionRevocation),
) (func(), error) {
	redisClient, err := s.getRedisClient(s.userSessionRevocationKey())
	if err != nil {
		return nil, err
	}

	pubsub := redisClient.Subscribe(s.userSessionRevocationKey())
	if _, err := pubsub.Receive(); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	go func() {
		for message := range pubsub.Channel() {
			var revocation UserSessionRevocation
			if err := json.Unmarshal([]byte(message.Payload), &revocation); err != nil {
				continue
			}

			handler(revocation)
		}
	}()

	return func() {
		_ = pubsub.Close()
	}, nil
}

func (s *RealtimeLeaseCacheClient) PublishNotification(event NotificationEvent) error {
	if event.EventId == uuid.Nil || event.RecipientUserPublicId == uuid.Nil {
		return errors.New("realtime notification event is incomplete")
	}
	redisClient, err := s.getRedisClient(event.RecipientUserPublicId.String())
	if err != nil {
		return err
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return redisClient.Publish(s.userNotificationKey(), payload).Err()
}

func (s *RealtimeLeaseCacheClient) SubscribeNotifications(
	handler func(NotificationEvent),
) (func(), error) {
	redisClient, err := s.getRedisClient(s.userNotificationKey())
	if err != nil {
		return nil, err
	}
	pubsub := redisClient.Subscribe(s.userNotificationKey())
	if _, err := pubsub.Receive(); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	go func() {
		for message := range pubsub.Channel() {
			var event NotificationEvent
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				continue
			}
			handler(event)
		}
	}()

	return func() {
		_ = pubsub.Close()
	}, nil
}

func (s *RealtimeLeaseCacheClient) PublishRoutineTaskLifecycleEvent(
	event RoutineTaskLifecycleEvent,
) error {
	if event.EventId == uuid.Nil || event.RoutineTaskId == uuid.Nil ||
		event.RoutineTaskRecordId == uuid.Nil || event.RoutineId == uuid.Nil ||
		event.ActorUserPublicId == uuid.Nil || event.Purpose == "" || event.Status == "" ||
		event.Attempt <= 0 || event.OccurredAt.IsZero() {
		return errors.New("realtime routine task lifecycle event is incomplete")
	}

	redisClient, err := s.getRedisClient(event.RoutineTaskId.String())
	if err != nil {
		return err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return redisClient.Publish(s.routineTaskLifecycleKey(), payload).Err()
}

func (s *RealtimeLeaseCacheClient) SubscribeRoutineTaskLifecycleEvents(
	handler func(RoutineTaskLifecycleEvent),
) (func(), error) {
	redisClient, err := s.getRedisClient(s.routineTaskLifecycleKey())
	if err != nil {
		return nil, err
	}

	pubsub := redisClient.Subscribe(s.routineTaskLifecycleKey())
	if _, err := pubsub.Receive(); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	go func() {
		for message := range pubsub.Channel() {
			var event RoutineTaskLifecycleEvent
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				continue
			}

			handler(event)
		}
	}()

	return func() {
		_ = pubsub.Close()
	}, nil
}

func (s *RealtimeLeaseCacheClient) PublishResourceEvent(event ResourceEvent) error {
	redisClient, err := s.getRedisClient(s.resourceEventKey())
	if err != nil {
		return err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return redisClient.Publish(s.resourceEventKey(), payload).Err()
}

func (s *RealtimeLeaseCacheClient) SubscribeResourceEvents(
	handler func(ResourceEvent),
) (func(), error) {
	redisClient, err := s.getRedisClient(s.resourceEventKey())
	if err != nil {
		return nil, err
	}

	pubsub := redisClient.Subscribe(s.resourceEventKey())
	if _, err := pubsub.Receive(); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	go func() {
		for message := range pubsub.Channel() {
			var event ResourceEvent
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				continue
			}

			handler(event)
		}
	}()

	return func() {
		_ = pubsub.Close()
	}, nil
}

func (s *RealtimeLeaseCacheClient) resourceEventKey() string {
	return "Realtime:resource:events"
}

/* ============================== Ticket Methods ============================== */

func (s *RealtimeLeaseCacheClient) ConsumeTicket(
	ticketId string,
	expiresAt time.Time,
) (bool, error) {
	if ticketId == "" {
		return false, errors.New("realtime ticket identifier is required")
	}

	expiresIn := time.Until(expiresAt)
	if expiresIn <= 0 {
		return false, nil
	}

	redisClient, err := s.getRedisClient(ticketId)
	if err != nil {
		return false, err
	}

	return redisClient.SetNX(
		fmt.Sprintf("Realtime:ticket:%s", ticketId),
		"1",
		expiresIn,
	).Result()
}

func (s *RealtimeLeaseCacheClient) release(identifier string, key string, member string) error {
	redisClient, err := s.getRedisClient(identifier)
	if err != nil {
		return err
	}

	return redisscripts.ReleaseRealtimeLease.Eval(redisClient, []string{key}, member).Err()
}

/* ============================== User Connection Methods ============================== */

func (s *RealtimeLeaseCacheClient) AcquireUserConnection(
	identifier uuid.UUID,
	connectionId uuid.UUID,
	maximumConnections int,
) (bool, int64, error) {
	identifierString := identifier.String()

	return s.acquire(
		identifierString,
		fmt.Sprintf("Realtime:user:%s:connections", identifier),
		connectionId.String(),
		maximumConnections,
	)
}

func (s *RealtimeLeaseCacheClient) RefreshUserConnection(
	identifier uuid.UUID,
	connectionId uuid.UUID,
) (bool, error) {
	identifierString := identifier.String()

	return s.refresh(
		identifierString,
		fmt.Sprintf("Realtime:user:%s:connections", identifier),
		connectionId.String(),
	)
}

func (s *RealtimeLeaseCacheClient) ReleaseUserConnection(
	identifier uuid.UUID,
	connectionId uuid.UUID,
) error {
	identifierString := identifier.String()

	return s.release(
		identifierString,
		fmt.Sprintf("Realtime:user:%s:connections", identifier),
		connectionId.String(),
	)
}

/* ============================== Block Pack Subscriber Methods ============================== */

func (s *RealtimeLeaseCacheClient) AcquireBlockPackSubscriber(
	blockPackId uuid.UUID,
	member string,
	maximumSubscribers int32,
) (bool, int64, error) {
	identifier := blockPackId.String()

	return s.acquire(
		identifier,
		fmt.Sprintf("Realtime:blockPack:%s:subscribers", blockPackId),
		member,
		int(maximumSubscribers),
	)
}

func (s *RealtimeLeaseCacheClient) RefreshBlockPackSubscriber(
	blockPackId uuid.UUID,
	member string,
) (bool, error) {
	identifier := blockPackId.String()

	refreshed, err := s.refresh(
		identifier,
		fmt.Sprintf("Realtime:blockPack:%s:subscribers", blockPackId),
		member,
	)
	if err != nil || !refreshed {
		return refreshed, err
	}

	redisClient, err := s.getRedisClient(identifier)
	if err != nil {
		return false, err
	}

	pipeline := redisClient.TxPipeline()
	pipeline.PExpire(s.blockPackParticipantKey(blockPackId), sconstants.RealtimeLeaseTTL)
	if _, err := pipeline.Exec(); err != nil {
		return false, err
	}

	return true, nil
}

func (s *RealtimeLeaseCacheClient) ReleaseBlockPackSubscriber(
	blockPackId uuid.UUID,
	member string,
) error {
	identifier := blockPackId.String()

	if err := s.release(
		identifier,
		fmt.Sprintf("Realtime:blockPack:%s:subscribers", blockPackId), member); err != nil {
		return err
	}

	redisClient, err := s.getRedisClient(identifier)
	if err != nil {
		return err
	}

	return redisClient.HDel(s.blockPackParticipantKey(blockPackId), member).Err()
}

func (s *RealtimeLeaseCacheClient) GetBlockPackParticipantByMember(
	blockPackId uuid.UUID,
	member string,
) (*RealtimeBlockPackParticipant, error) {
	redisClient, err := s.getRedisClient(blockPackId.String())
	if err != nil {
		return nil, err
	}

	payload, err := redisClient.HGet(s.blockPackParticipantKey(blockPackId), member).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	participantPayload := struct {
		UserPublicId      string `json:"userPublicId"`
		ChannelPermission string `json:"channelPermission"`
	}{}
	if err := json.Unmarshal(payload, &participantPayload); err != nil {
		return nil, err
	}

	userPublicId, err := uuid.Parse(participantPayload.UserPublicId)
	if err != nil {
		return nil, err
	}

	return &RealtimeBlockPackParticipant{
		UserPublicId:      userPublicId,
		ChannelPermission: participantPayload.ChannelPermission,
		ConnectionCount:   1,
	}, nil
}

func (s *RealtimeLeaseCacheClient) GetBlockPackParticipants(
	blockPackId uuid.UUID,
) ([]RealtimeBlockPackParticipant, error) {
	redisClient, err := s.getRedisClient(blockPackId.String())
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	expiredMembers, err := redisClient.ZRangeByScore(
		fmt.Sprintf("Realtime:blockPack:%s:subscribers", blockPackId),
		redis.ZRangeBy{
			Min: "-inf",
			Max: strconv.FormatInt(now, 10),
		},
	).Result()
	if err != nil {
		return nil, err
	}

	pipeline := redisClient.TxPipeline()
	pipeline.ZRemRangeByScore(
		fmt.Sprintf("Realtime:blockPack:%s:subscribers", blockPackId),
		"-inf",
		strconv.FormatInt(now, 10),
	)
	if len(expiredMembers) > 0 {
		participantKey := s.blockPackParticipantKey(blockPackId)
		members := make([]string, len(expiredMembers))
		copy(members, expiredMembers)
		pipeline.HDel(participantKey, members...)
	}
	if _, err := pipeline.Exec(); err != nil {
		return nil, err
	}

	members, err := redisClient.ZRangeByScore(
		fmt.Sprintf("Realtime:blockPack:%s:subscribers", blockPackId),
		redis.ZRangeBy{
			Min: strconv.FormatInt(now+1, 10),
			Max: "+inf",
		},
	).Result()
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return []RealtimeBlockPackParticipant{}, nil
	}

	values, err := redisClient.HMGet(s.blockPackParticipantKey(blockPackId), members...).Result()
	if err != nil {
		return nil, err
	}

	participantsByPublicId := make(map[uuid.UUID]RealtimeBlockPackParticipant, len(values))
	for _, value := range values {
		var payload []byte
		switch typedValue := value.(type) {
		case string:
			payload = []byte(typedValue)
		case []byte:
			payload = typedValue
		default:
			continue
		}

		participant := struct {
			UserPublicId      string `json:"userPublicId"`
			ChannelPermission string `json:"channelPermission"`
		}{}
		if err := json.Unmarshal(payload, &participant); err != nil {
			continue
		}

		userPublicId, err := uuid.Parse(participant.UserPublicId)
		if err != nil {
			continue
		}

		aggregate := participantsByPublicId[userPublicId]
		aggregate.UserPublicId = userPublicId
		aggregate.ConnectionCount++
		if aggregate.ChannelPermission != "write" || participant.ChannelPermission == "write" {
			aggregate.ChannelPermission = participant.ChannelPermission
		}
		participantsByPublicId[userPublicId] = aggregate
	}

	participants := make([]RealtimeBlockPackParticipant, 0, len(participantsByPublicId))
	for _, participant := range participantsByPublicId {
		participants = append(participants, participant)
	}
	sort.Slice(participants, func(first int, second int) bool {
		return participants[first].UserPublicId.String() < participants[second].UserPublicId.String()
	})

	return participants, nil
}

func (s *RealtimeLeaseCacheClient) GetBlockPackSubscriberLeases(
	blockPackId uuid.UUID,
) ([]RealtimeBlockPackSubscriberLease, error) {
	redisClient, err := s.getRedisClient(blockPackId.String())
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	leases, err := redisClient.ZRangeByScoreWithScores(
		fmt.Sprintf("Realtime:blockPack:%s:subscribers", blockPackId),
		redis.ZRangeBy{Min: strconv.FormatInt(now+1, 10), Max: "+inf"},
	).Result()
	if err != nil {
		return nil, err
	}

	result := make([]RealtimeBlockPackSubscriberLease, len(leases))
	for index, lease := range leases {
		member, ok := lease.Member.(string)
		if !ok {
			return nil, errors.New("realtime redis subscriber lease returned an invalid member")
		}

		result[index] = RealtimeBlockPackSubscriberLease{
			Member:    member,
			ExpiresAt: time.UnixMilli(int64(lease.Score)),
		}
	}

	return result, nil
}

func (s *RealtimeLeaseCacheClient) PublishBlockPackPresenceEvent(
	event RealtimeBlockPackPresenceEvent,
) error {
	redisClient, err := s.getRedisClient(event.BlockPackId.String())
	if err != nil {
		return err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return redisClient.Publish(s.blockPackPresenceKey(), payload).Err()
}

func (s *RealtimeLeaseCacheClient) SubscribeBlockPackPresenceEvents(
	handler func(RealtimeBlockPackPresenceEvent),
) (func(), error) {
	redisClient, err := s.getRedisClient(s.blockPackPresenceKey())
	if err != nil {
		return nil, err
	}

	pubsub := redisClient.Subscribe(s.blockPackPresenceKey())
	if _, err := pubsub.Receive(); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	go func() {
		for message := range pubsub.Channel() {
			var event RealtimeBlockPackPresenceEvent
			if err := json.Unmarshal([]byte(message.Payload), &event); err != nil {
				continue
			}

			handler(event)
		}
	}()

	return func() {
		_ = pubsub.Close()
	}, nil
}

/* ============================== Block Pack Participant Methods ============================== */

func (s *RealtimeLeaseCacheClient) SetBlockPackParticipant(
	blockPackId uuid.UUID,
	member string,
	identifier uuid.UUID,
	channelPermission string,
) error {
	redisClient, err := s.getRedisClient(blockPackId.String())
	if err != nil {
		return err
	}

	payload, err := json.Marshal(struct {
		Member            string `json:"member"`
		UserPublicId      string `json:"userPublicId"`
		ChannelPermission string `json:"channelPermission"`
	}{
		Member:            member,
		UserPublicId:      identifier.String(),
		ChannelPermission: channelPermission,
	})
	if err != nil {
		return err
	}

	pipeline := redisClient.TxPipeline()
	pipeline.HSet(s.blockPackParticipantKey(blockPackId), member, payload)
	pipeline.PExpire(s.blockPackParticipantKey(blockPackId), sconstants.RealtimeLeaseTTL)

	_, err = pipeline.Exec()

	return err
}
