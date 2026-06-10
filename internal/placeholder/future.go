package placeholder

import "context"

// Repository is a placeholder for future PostgreSQL persistence.
type Repository interface{}

// JobQueue is a placeholder for future background jobs.
type JobQueue interface{}

// AIAnalyzer is a placeholder for future AI lead analysis.
type AIAnalyzer interface {
	AnalyzeLead(ctx context.Context, leadID string) error
}

// ReportGenerator is a placeholder for future report generation.
type ReportGenerator interface{}

// Authenticator is a placeholder for future authentication.
type Authenticator interface{}

// EmailNotifier is a placeholder for future email notifications.
type EmailNotifier interface{}
