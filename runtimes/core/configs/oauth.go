package config

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthGoogleConfig struct {
	ClientId     string
	ClientSecret string
	RedirectUrl  string
}

func loadOAuthGoogleConfig() (OAuthGoogleConfig, error) {
	config := OAuthGoogleConfig{
		ClientId:     strings.TrimSpace(os.Getenv("OAUTH_GOOGLE_CLIENT_ID")),
		ClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
		RedirectUrl:  strings.TrimSpace(os.Getenv("OAUTH_GOOGLE_REDIRECT_URL")),
	}
	if config.ClientId == "" || config.ClientSecret == "" || config.RedirectUrl == "" {
		return OAuthGoogleConfig{}, fmt.Errorf("OAUTH_GOOGLE_CLIENT_ID, OAUTH_GOOGLE_CLIENT_SECRET, and OAUTH_GOOGLE_REDIRECT_URL are required")
	}

	return config, nil
}

func (c OAuthGoogleConfig) OAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ClientId,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
