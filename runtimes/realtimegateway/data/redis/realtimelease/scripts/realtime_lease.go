package redisscripts

import (
	_ "embed"

	"github.com/go-redis/redis"
)

var (
	//go:embed realtime_lease_acquire.lua
	acquireRealtimeLeaseContent string

	//go:embed realtime_lease_refresh.lua
	refreshRealtimeLeaseContent string

	//go:embed realtime_lease_release.lua
	releaseRealtimeLeaseContent string

	AcquireRealtimeLease = redis.NewScript(acquireRealtimeLeaseContent)
	RefreshRealtimeLease = redis.NewScript(refreshRealtimeLeaseContent)
	ReleaseRealtimeLease = redis.NewScript(releaseRealtimeLeaseContent)
)
