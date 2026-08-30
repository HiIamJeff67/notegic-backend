package ratelimiter

import (
	"sync"
	"time"

	gatewayconfig "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/configs"
)

// The weak rate limiter is an implementation of Leaky Bucket algorithm
// which in our case is not powerful then the bybrid rate limiter
// with redis cache for cross servers request sources and Token Bucket algorithm based rate limiter
type WeakRateLimiter struct {
	requestArrivalTimes []time.Time
	capacity            int
	minInterval         time.Duration
	minRetention        time.Duration
	mutex               sync.Mutex
}

func NewWeakRateLimiter(requestsPerSecond int, config gatewayconfig.RateLimitConfig) *WeakRateLimiter {
	minInterval := time.Second / time.Duration(requestsPerSecond)
	return &WeakRateLimiter{
		requestArrivalTimes: make([]time.Time, 0),
		capacity:            requestsPerSecond + config.RequestFrequencyExtraCapacity,
		minInterval:         minInterval,
		minRetention:        config.MinIntervalTimeOfLastRequest,
	}
}

func (lb *WeakRateLimiter) Allow() bool {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	now := time.Now()

	validRequests := make([]time.Time, 0)
	for _, reqArrivalTime := range lb.requestArrivalTimes {
		if now.Sub(reqArrivalTime) < lb.minRetention {
			validRequests = append(validRequests, reqArrivalTime)
		}
	}
	lb.requestArrivalTimes = validRequests

	if len(lb.requestArrivalTimes) >= lb.capacity {
		return false
	}

	if len(lb.requestArrivalTimes) > 0 {
		lastReqArrivalTime := lb.requestArrivalTimes[len(lb.requestArrivalTimes)-1]
		if now.Sub(lastReqArrivalTime) < lb.minInterval {
			return false
		}
	}

	lb.requestArrivalTimes = append(lb.requestArrivalTimes, now)
	return true
}
