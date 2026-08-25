package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type YjsMaintenanceStrategyConfig struct {
	MaximumPendingHints    int
	MaximumDispatchBatch   int
	MaximumDispatchWorkers int
	MaximumRequestAttempts int
}

func loadYjsMaintenanceStrategyConfig() (YjsMaintenanceStrategyConfig, error) {
	maximumPendingHints, err := positiveInteger("CORE_YJS_MAINTENANCE_MAXIMUM_PENDING_HINTS")
	if err != nil {
		return YjsMaintenanceStrategyConfig{}, err
	}
	maximumDispatchBatch, err := positiveInteger("CORE_YJS_MAINTENANCE_MAXIMUM_DISPATCH_BATCH")
	if err != nil {
		return YjsMaintenanceStrategyConfig{}, err
	}
	maximumDispatchWorkers, err := positiveInteger("CORE_YJS_MAINTENANCE_MAXIMUM_DISPATCH_WORKERS")
	if err != nil {
		return YjsMaintenanceStrategyConfig{}, err
	}
	maximumRequestAttempts, err := positiveInteger("CORE_YJS_MAINTENANCE_MAXIMUM_REQUEST_ATTEMPTS")
	if err != nil {
		return YjsMaintenanceStrategyConfig{}, err
	}

	return YjsMaintenanceStrategyConfig{
		MaximumPendingHints:    maximumPendingHints,
		MaximumDispatchBatch:   maximumDispatchBatch,
		MaximumDispatchWorkers: maximumDispatchWorkers,
		MaximumRequestAttempts: maximumRequestAttempts,
	}, nil
}

func positiveInteger(name string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}
