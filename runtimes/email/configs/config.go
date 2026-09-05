package config

import (
	"fmt"
	"net/mail"
	"net/url"
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
	links, err := loadEmailLinksConfig()
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
		Renderers:     loadRendererConfig(links),
		Kafka:         kafka,
		KafkaConsumer: kafkaConsumer,
	}, nil
}

func loadEmailLinksConfig() (EmailLinksConfig, error) {
	webBaseUrl := strings.TrimRight(strings.TrimSpace(os.Getenv("NOTEGIC_WEB_BASE_URL")), "/")
	termsUrl := strings.TrimSpace(os.Getenv("NOTEGIC_TERMS_URL"))
	contactEmail := strings.TrimSpace(os.Getenv("NOTEGIC_OFFICIAL_GMAIL"))
	if webBaseUrl == "" || termsUrl == "" || contactEmail == "" {
		return EmailLinksConfig{}, fmt.Errorf("NOTEGIC_WEB_BASE_URL, NOTEGIC_TERMS_URL, and NOTEGIC_OFFICIAL_GMAIL are required")
	}

	webBase, err := url.Parse(webBaseUrl)
	if err != nil || (webBase.Scheme != "http" && webBase.Scheme != "https") || webBase.Host == "" {
		return EmailLinksConfig{}, fmt.Errorf("NOTEGIC_WEB_BASE_URL must be an absolute HTTP or HTTPS URL")
	}
	terms, err := url.Parse(termsUrl)
	if err != nil || (terms.Scheme != "http" && terms.Scheme != "https") || terms.Host == "" {
		return EmailLinksConfig{}, fmt.Errorf("NOTEGIC_TERMS_URL must be an absolute HTTP or HTTPS URL")
	}
	parsedContact, err := mail.ParseAddress(contactEmail)
	if err != nil || parsedContact.Address != contactEmail {
		return EmailLinksConfig{}, fmt.Errorf("NOTEGIC_OFFICIAL_GMAIL must be a valid email address")
	}

	return EmailLinksConfig{
		WebBaseUrl:   webBaseUrl,
		TermsUrl:     termsUrl,
		ContactEmail: contactEmail,
	}, nil
}
