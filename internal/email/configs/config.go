package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ListenAddress string
	SMTP          SMTPConfig
	Renderers     RendererConfigs
	Kafka         KafkaConnectionConfig
	KafkaConsumer KafkaConsumerConfig
}

func LoadConfig() (Config, error) {
	listenAddress := strings.TrimSpace(os.Getenv("EMAIL_LISTEN_ADDRESS"))
	if listenAddress == "" {
		return Config{}, fmt.Errorf("EMAIL_LISTEN_ADDRESS is required")
	}

	smtp, err := loadSMTPConfig()
	if err != nil {
		return Config{}, err
	}
	kafka, kafkaConsumer, err := loadKafkaConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		ListenAddress: listenAddress,
		SMTP:          smtp,
		Renderers:     loadRendererConfig(),
		Kafka:         kafka,
		KafkaConsumer: kafkaConsumer,
	}, nil
}
