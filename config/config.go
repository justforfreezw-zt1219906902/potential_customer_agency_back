package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	HubSpotAccessToken string
	CORSAllowedOrigins string
	ResendAPIKey       string
	ResendFromEmail    string
	NotificationEmails []string
	DatabaseURL        string
	PostgresDB         string
	PostgresUser       string
	PostgresPassword   string
	PostgresSSLMode    string
	DemoCompanyID      uuid.UUID
}

func Load() Config {
	_ = godotenv.Load()

	demoCompanyID, err := uuid.Parse(strings.TrimSpace(os.Getenv("DEMO_COMPANY_PROFILE_ID")))
	if err != nil {
		panic(fmt.Sprintf("DEMO_COMPANY_PROFILE_ID must be a valid UUID: %v", err))
	}

	return Config{
		Port:               getEnv("PORT", "8080"),
		HubSpotAccessToken: os.Getenv("HUBSPOT_ACCESS_TOKEN"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		ResendAPIKey:       os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:    getEnv("RESEND_FROM_EMAIL", "Mi Goto <onboarding@resend.dev>"),
		NotificationEmails: parseEmailList(getEnv("NOTIFICATION_EMAILS", "sales@mi-goto.com")),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		PostgresDB:         firstEnv("POSTGRES_DB", "PGDATABASE"),
		PostgresUser:       firstEnv("POSTGRES_USER", "PGUSER"),
		PostgresPassword:   firstEnv("POSTGRES_PASSWORD", "PGPASSWORD"),
		PostgresSSLMode:    getEnv("POSTGRES_SSLMODE", "disable"),
		DemoCompanyID:      demoCompanyID,
	}
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
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
