package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SMTPConfig struct {
	Host     string
	Port     int
	UserName string
	Password string
	From     string
}

func loadSMTPConfig() (SMTPConfig, error) {
	port, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SMTP_PORT")))
	if err != nil || port <= 0 {
		return SMTPConfig{}, fmt.Errorf("SMTP_PORT must be a positive integer")
	}
	name := strings.TrimSpace(os.Getenv("NOTEGIC_OFFICIAL_NAME"))
	address := strings.TrimSpace(os.Getenv("NOTEGIC_OFFICIAL_GMAIL"))
	config := SMTPConfig{
		Host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		Port:     port,
		UserName: address,
		Password: os.Getenv("NOTEGIC_OFFICIAL_GOOGLE_APPLICATION_PASSWORD"),
		From:     name + " <" + address + ">",
	}
	if config.Host == "" || config.UserName == "" || config.Password == "" || name == "" {
		return SMTPConfig{}, fmt.Errorf("SMTP_HOST, NOTEGIC_OFFICIAL_NAME, NOTEGIC_OFFICIAL_GMAIL, and NOTEGIC_OFFICIAL_GOOGLE_APPLICATION_PASSWORD are required")
	}

	return config, nil
}
