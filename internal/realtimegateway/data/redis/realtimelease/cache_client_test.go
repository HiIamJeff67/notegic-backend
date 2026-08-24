package realtimelease

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/google/uuid"

	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"
)

func newTestRealtimeLeaseCacheClient(t *testing.T, redisClient *redis.Client) *RealtimeLeaseCacheClient {
	t.Helper()

	clientSet := platformredis.NewClientSetFromClients(redisClient)
	cacheStore := NewRealtimeLeaseCacheStore(clientSet)
	return NewRealtimeLeaseCacheClient(cacheStore)
}

func TestRealtimeLeaseCacheClientLimitsConcurrentUserConnections(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start test redis server: %v", err)
	}
	defer server.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer redisClient.Close()

	store := newTestRealtimeLeaseCacheClient(t, redisClient)
	userPublicId := uuid.New()

	var acquiredCount atomic.Int32
	var waitGroup sync.WaitGroup
	for range 32 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			acquired, _, err := store.AcquireUserConnection(userPublicId, uuid.New(), 1)
			if err != nil {
				t.Errorf("failed to acquire realtime user connection lease: %v", err)

				return
			}
			if acquired {
				acquiredCount.Add(1)
			}
		}()
	}
	waitGroup.Wait()

	if acquiredCount.Load() != 1 {
		t.Fatalf("expected exactly one concurrent lease acquisition, got %d", acquiredCount.Load())
	}
}

func TestRealtimeLeaseCacheClientConsumesTicketOnlyOnce(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start test redis server: %v", err)
	}
	defer server.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer redisClient.Close()

	store := newTestRealtimeLeaseCacheClient(t, redisClient)
	ticketId := uuid.NewString()
	expiresAt := time.Now().Add(time.Minute)

	var consumedCount atomic.Int32
	var waitGroup sync.WaitGroup
	for range 32 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			consumed, err := store.ConsumeTicket(ticketId, expiresAt)
			if err != nil {
				t.Errorf("failed to consume realtime ticket: %v", err)

				return
			}
			if consumed {
				consumedCount.Add(1)
			}
		}()
	}
	waitGroup.Wait()

	if consumedCount.Load() != 1 {
		t.Fatalf("expected exactly one ticket consumption, got %d", consumedCount.Load())
	}
}

func TestRealtimeLeaseCacheClientReclaimsExpiredUserConnectionLease(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start test redis server: %v", err)
	}
	defer server.Close()

	firstRedisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer firstRedisClient.Close()
	secondRedisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer secondRedisClient.Close()

	firstStore := newTestRealtimeLeaseCacheClient(t, firstRedisClient)
	secondStore := newTestRealtimeLeaseCacheClient(t, secondRedisClient)
	userPublicId := uuid.New()

	acquired, _, err := firstStore.AcquireUserConnection(userPublicId, uuid.New(), 1)
	if err != nil || !acquired {
		t.Fatalf("expected first user connection lease to be acquired: %v", err)
	}

	acquired, _, err = secondStore.AcquireUserConnection(userPublicId, uuid.New(), 1)
	if err != nil {
		t.Fatalf("failed to check second user connection lease: %v", err)
	}
	if acquired {
		t.Fatal("expected the second Redis client to observe the distributed user connection cap")
	}

	server.FastForward(constants.RealtimeLeaseTTL)

	acquired, _, err = secondStore.AcquireUserConnection(userPublicId, uuid.New(), 1)
	if err != nil || !acquired {
		t.Fatalf("expected expired user connection lease to be reclaimed: %v", err)
	}
}

func TestRealtimeLeaseCacheClientReleasesBlockPackSubscriber(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start test redis server: %v", err)
	}
	defer server.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer redisClient.Close()

	store := newTestRealtimeLeaseCacheClient(t, redisClient)
	blockPackId := uuid.New()

	acquired, _, err := store.AcquireBlockPackSubscriber(blockPackId, "connector-a:1", 1)
	if err != nil || !acquired {
		t.Fatalf("expected first BlockPack subscriber lease to be acquired: %v", err)
	}

	acquired, _, err = store.AcquireBlockPackSubscriber(blockPackId, "connector-b:1", 1)
	if err != nil {
		t.Fatalf("failed to check BlockPack subscriber lease: %v", err)
	}
	if acquired {
		t.Fatal("expected BlockPack subscriber capacity to reject the second lease")
	}

	leases, err := store.GetBlockPackSubscriberLeases(blockPackId)
	if err != nil {
		t.Fatalf("failed to list BlockPack subscriber leases: %v", err)
	}
	if len(leases) != 1 || leases[0].Member != "connector-a:1" || !leases[0].ExpiresAt.After(time.Now()) {
		t.Fatalf("unexpected BlockPack subscriber leases: %#v", leases)
	}

	if err := store.ReleaseBlockPackSubscriber(blockPackId, "connector-a:1"); err != nil {
		t.Fatalf("failed to release BlockPack subscriber lease: %v", err)
	}

	leases, err = store.GetBlockPackSubscriberLeases(blockPackId)
	if err != nil {
		t.Fatalf("failed to list BlockPack subscriber leases after release: %v", err)
	}
	if len(leases) != 0 {
		t.Fatalf("expected no BlockPack subscriber leases after release, got %#v", leases)
	}

	acquired, _, err = store.AcquireBlockPackSubscriber(blockPackId, "connector-b:1", 1)
	if err != nil || !acquired {
		t.Fatalf("expected released BlockPack subscriber capacity to be reusable: %v", err)
	}
}

func TestRealtimeLeaseCacheClientAggregatesBlockPackParticipants(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start test redis server: %v", err)
	}
	defer server.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer redisClient.Close()

	store := newTestRealtimeLeaseCacheClient(t, redisClient)
	blockPackId := uuid.New()
	firstUserPublicId := uuid.New()
	secondUserPublicId := uuid.New()
	for _, lease := range []struct {
		member            string
		userPublicId      uuid.UUID
		channelPermission string
	}{
		{
			member:            "connector-a:1",
			userPublicId:      firstUserPublicId,
			channelPermission: "read",
		},
		{
			member:            "connector-a:2",
			userPublicId:      firstUserPublicId,
			channelPermission: "write",
		},
		{
			member:            "connector-b:1",
			userPublicId:      secondUserPublicId,
			channelPermission: "read",
		},
	} {
		acquired, _, err := store.AcquireBlockPackSubscriber(blockPackId, lease.member, 3)
		if err != nil || !acquired {
			t.Fatalf("acquire BlockPack subscriber %s: %v", lease.member, err)
		}
		if err := store.SetBlockPackParticipant(
			blockPackId,
			lease.member,
			lease.userPublicId,
			lease.channelPermission,
		); err != nil {
			t.Fatalf("set BlockPack participant %s: %v", lease.member, err)
		}
	}

	participants, err := store.GetBlockPackParticipants(blockPackId)
	if err != nil {
		t.Fatalf("get BlockPack participants: %v", err)
	}
	if len(participants) != 2 {
		t.Fatalf("expected two aggregated participants, got %#v", participants)
	}

	participantsByPublicId := make(map[uuid.UUID]RealtimeBlockPackParticipant, len(participants))
	for _, participant := range participants {
		participantsByPublicId[participant.UserPublicId] = participant
	}
	firstParticipant := participantsByPublicId[firstUserPublicId]
	if firstParticipant.ConnectionCount != 2 || firstParticipant.ChannelPermission != "write" {
		t.Fatalf("unexpected aggregated first participant: %#v", firstParticipant)
	}
	secondParticipant := participantsByPublicId[secondUserPublicId]
	if secondParticipant.ConnectionCount != 1 || secondParticipant.ChannelPermission != "read" {
		t.Fatalf("unexpected aggregated second participant: %#v", secondParticipant)
	}

	participant, err := store.GetBlockPackParticipantByMember(blockPackId, "connector-a:2")
	if err != nil {
		t.Fatalf("get BlockPack participant by member: %v", err)
	}
	if participant == nil || participant.UserPublicId != firstUserPublicId {
		t.Fatalf("unexpected participant by member: %#v", participant)
	}
}

func TestRealtimeLeaseCacheClientPublishesBlockPackPresenceEvent(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start test redis server: %v", err)
	}
	defer server.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer redisClient.Close()

	store := newTestRealtimeLeaseCacheClient(t, redisClient)
	blockPackId := uuid.New()
	userPublicId := uuid.New()
	received := make(chan RealtimeBlockPackPresenceEvent, 1)
	shutdown, err := store.SubscribeBlockPackPresenceEvents(func(event RealtimeBlockPackPresenceEvent) {
		received <- event
	})
	if err != nil {
		t.Fatalf("subscribe to BlockPack presence events: %v", err)
	}
	defer shutdown()

	if err := store.PublishBlockPackPresenceEvent(RealtimeBlockPackPresenceEvent{
		Type:        RealtimeBlockPackPresenceEventType_Joined,
		BlockPackId: blockPackId,
		Participant: RealtimeBlockPackParticipant{
			UserPublicId:      userPublicId,
			ChannelPermission: "read",
			ConnectionCount:   1,
		},
	}); err != nil {
		t.Fatalf("publish BlockPack presence event: %v", err)
	}

	select {
	case event := <-received:
		if event.Type != RealtimeBlockPackPresenceEventType_Joined ||
			event.BlockPackId != blockPackId ||
			event.Participant.UserPublicId != userPublicId {
			t.Fatalf("unexpected BlockPack presence event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for BlockPack presence event")
	}
}
