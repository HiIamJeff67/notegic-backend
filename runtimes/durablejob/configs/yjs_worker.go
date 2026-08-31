package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type YjsDocumentInitializationConfig struct {
	Endpoint       string
	UpdateEndpoint string
	Timeout        time.Duration
}

func loadYjsDocumentInitializationConfig() (YjsDocumentInitializationConfig, error) {
	config := YjsDocumentInitializationConfig{
		Endpoint: strings.TrimRight(
			strings.TrimSpace(os.Getenv("YJS_DOCUMENT_INITIALIZATION_WORKER_URL")),
			"/",
		),
	}
	if config.Endpoint == "" {
		return YjsDocumentInitializationConfig{}, fmt.Errorf("YJS_DOCUMENT_INITIALIZATION_WORKER_URL is required")
	}
	config.UpdateEndpoint = strings.TrimSuffix(
		config.Endpoint,
		"/core/yjs-document-initialization/v1",
	) + "/core/yjs-block-pack-update/v1"
	timeout, err := time.ParseDuration(strings.TrimSpace(os.Getenv("YJS_DOCUMENT_INITIALIZATION_WORKER_TIMEOUT")))
	if err != nil || timeout <= 0 {
		return YjsDocumentInitializationConfig{}, fmt.Errorf("YJS_DOCUMENT_INITIALIZATION_WORKER_TIMEOUT must be a positive Go duration")
	}
	config.Timeout = timeout
	return config, nil
}
