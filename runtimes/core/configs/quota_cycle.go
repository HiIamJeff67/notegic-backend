package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type QuotaCycleWorkerConfig struct {
	Interval time.Duration
}

func loadQuotaCycleWorkerConfig() (QuotaCycleWorkerConfig, error) {
	interval, err := time.ParseDuration(
		strings.TrimSpace(os.Getenv("CORE_QUOTA_CYCLE_WORKER_INTERVAL")),
	)
	if err != nil || interval <= 0 {
		return QuotaCycleWorkerConfig{}, fmt.Errorf("CORE_QUOTA_CYCLE_WORKER_INTERVAL must be a positive Go duration")
	}

	return QuotaCycleWorkerConfig{
		Interval: interval,
	}, nil
}
