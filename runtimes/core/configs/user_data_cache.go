package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type UserDataCacheConfig struct {
	CacheExpiresIn     time.Duration
	MaxRotationRetries int
}

func loadUserDataCacheConfig() (UserDataCacheConfig, error) {
	cacheExpiresIn, err := time.ParseDuration(strings.TrimSpace(os.Getenv("CORE_USER_DATA_CACHE_EXPIRES_IN")))
	if err != nil || cacheExpiresIn <= 0 {
		return UserDataCacheConfig{}, fmt.Errorf("CORE_USER_DATA_CACHE_EXPIRES_IN must be a positive Go duration")
	}
	maxRotationRetries, err := strconv.Atoi(strings.TrimSpace(os.Getenv("CORE_USER_DATA_CACHE_MAX_ROTATION_RETRIES")))
	if err != nil || maxRotationRetries <= 0 {
		return UserDataCacheConfig{}, fmt.Errorf("CORE_USER_DATA_CACHE_MAX_ROTATION_RETRIES must be a positive integer")
	}

	return UserDataCacheConfig{
		CacheExpiresIn:     cacheExpiresIn,
		MaxRotationRetries: maxRotationRetries,
	}, nil
}
