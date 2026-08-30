package config

import (
	"time"

	rate "golang.org/x/time/rate"

	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	ratelimitrecord "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/data/redis/ratelimitrecord"
)

type RateLimitConfig struct {
	RateLimit         rate.Limit
	Burst             int
	UserLimit         int32
	WindowDuration    time.Duration
	BackendServerName sredis.BackendServerName
	CacheClient       *ratelimitrecord.RateLimitRecordCacheClient

	RequestFrequencyExtraCapacity        int
	MinIntervalTimeOfLastRequest         time.Duration
	SynchronizationToWindowDurationRatio int64
	MinSynchronizationInterval           time.Duration
}

func DefaultUpgradeRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RateLimit:                            5,
		Burst:                                10,
		UserLimit:                            60,
		WindowDuration:                       time.Minute,
		BackendServerName:                    sredis.BackendServerName_EastAsia,
		RequestFrequencyExtraCapacity:        2,
		MinIntervalTimeOfLastRequest:         time.Microsecond,
		SynchronizationToWindowDurationRatio: 10,
		MinSynchronizationInterval:           time.Second,
	}
}
