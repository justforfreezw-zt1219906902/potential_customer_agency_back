package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	HubSpotAccessToken string
	CORSAllowedOrigins string
	ResendAPIKey       string
	NotificationEmails []string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:               getEnv("PORT", "8080"),
		HubSpotAccessToken: os.Getenv("HUBSPOT_ACCESS_TOKEN"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		ResendAPIKey:       os.Getenv("RESEND_API_KEY"),
		NotificationEmails: parseEmailList(getEnv("NOTIFICATION_EMAILS", "sales@mi-goto.com")),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func parseEmailList(value string) []string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	var emails []string
	for _, item := range strings.Split(value, ",") {
		email := strings.TrimSpace(item)
		email = strings.Trim(email, `"'`)
		if email == "" {
			continue
		}
		emails = append(emails, email)
	}

	return emails
}
