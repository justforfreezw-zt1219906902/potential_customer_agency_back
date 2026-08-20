package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	HubSpotAccessToken     string
	HubSpotOwnerID         string
	CORSAllowedOrigins     string
	ResendAPIKey           string
	ResendFromEmail        string
	NotificationEmails     []string
	DatabaseURL            string
	PostgresDB             string
	PostgresUser           string
	PostgresPassword       string
	PostgresSSLMode        string
	DemoCompanyID          uuid.UUID
	GeminiAPIKey           string
	GeminiModel            string
	OutreachRequestTimeout time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	demoCompanyID, err := uuid.Parse(strings.TrimSpace(os.Getenv("DEMO_COMPANY_PROFILE_ID")))
	if err != nil {
		panic(fmt.Sprintf("DEMO_COMPANY_PROFILE_ID must be a valid UUID: %v", err))
	}

	return Config{
		Port:                   getEnv("PORT", "8080"),
		HubSpotAccessToken:     os.Getenv("HUBSPOT_ACCESS_TOKEN"),
		HubSpotOwnerID:         requiredEnv("HUBSPOT_OWNER_ID"),
		CORSAllowedOrigins:     getEnv("CORS_ALLOWED_ORIGINS", "*"),
		ResendAPIKey:           os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:        getEnv("RESEND_FROM_EMAIL", "Mi Goto <onboarding@resend.dev>"),
		NotificationEmails:     parseEmailList(getEnv("NOTIFICATION_EMAILS", "sales@mi-goto.com")),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		PostgresDB:             firstEnv("POSTGRES_DB", "PGDATABASE"),
		PostgresUser:           firstEnv("POSTGRES_USER", "PGUSER"),
		PostgresPassword:       firstEnv("POSTGRES_PASSWORD", "PGPASSWORD"),
		PostgresSSLMode:        getEnv("POSTGRES_SSLMODE", "disable"),
		DemoCompanyID:          demoCompanyID,
		GeminiAPIKey:           os.Getenv("GEMINI_API_KEY"),
		GeminiModel:            getEnv("GEMINI_MODEL", "gemini-3.6-flash"),
		OutreachRequestTimeout: parsePositiveSeconds("OUTREACH_REQUEST_TIMEOUT_SECONDS", 30),
	}
}

func requiredEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return value
}

func parsePositiveSeconds(key string, fallback int) time.Duration {
	raw := strings.TrimSpace(getEnv(key, strconv.Itoa(fallback)))
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		panic(fmt.Sprintf("%s must be a positive integer number of seconds", key))
	}
	return time.Duration(seconds) * time.Second
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
