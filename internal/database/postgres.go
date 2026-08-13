package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	URL      string
	Database string
	Username string
	Password string
	SSLMode  string
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	connectionURL, err := buildConnectionURL(cfg)
	if err != nil {
		return nil, err
	}

	poolConfig, err := pgxpool.ParseConfig(connectionURL)
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL configuration: %w", err)
	}
	poolConfig.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return pool, nil
}

func buildConnectionURL(cfg Config) (string, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return "", fmt.Errorf("DATABASE_URL is required")
	}
	if strings.TrimSpace(cfg.Username) == "" || strings.TrimSpace(cfg.Password) == "" {
		return "", fmt.Errorf("POSTGRES_USER and POSTGRES_PASSWORD are required")
	}

	parsed, err := url.Parse(cfg.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("DATABASE_URL must be a valid PostgreSQL address")
	}
	parsed.User = url.UserPassword(cfg.Username, cfg.Password)
	if cfg.Database != "" {
		parsed.Path = "/" + strings.TrimPrefix(cfg.Database, "/")
	}
	query := parsed.Query()
	if query.Get("sslmode") == "" && cfg.SSLMode != "" {
		query.Set("sslmode", cfg.SSLMode)
		parsed.RawQuery = query.Encode()
	}
	return parsed.String(), nil
}

func ValidateCompanyID(id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("DEMO_COMPANY_PROFILE_ID must not be empty")
	}
	return nil
}
