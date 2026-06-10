package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	HubSpotAccessToken string
	CORSAllowedOrigins string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:               getEnv("PORT", "8080"),
		HubSpotAccessToken: os.Getenv("HUBSPOT_ACCESS_TOKEN"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
