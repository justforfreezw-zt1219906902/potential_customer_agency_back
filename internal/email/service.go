package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

const (
	resendBaseURL = "https://api.resend.com"
)

type EmailService interface {
	SendLeadNotification(ctx context.Context, lead models.LeadRequest, hubSpotContactID string, submittedAt time.Time) error
}

type ResendEmailService struct {
	apiKey     string
	from       string
	recipients []string
	httpClient *http.Client
	logger     *log.Logger
}

func NewResendEmailService(apiKey string, from string, recipients []string, logger *log.Logger) *ResendEmailService {
	return &ResendEmailService{
		apiKey:     apiKey,
		from:       from,
		recipients: recipients,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (s *ResendEmailService) SendLeadNotification(ctx context.Context, lead models.LeadRequest, hubSpotContactID string, submittedAt time.Time) error {
	if strings.TrimSpace(s.apiKey) == "" {
		return fmt.Errorf("RESEND_API_KEY is required")
	}
	if strings.TrimSpace(s.from) == "" {
		return fmt.Errorf("RESEND_FROM_EMAIL is required")
	}
	if len(s.recipients) == 0 {
		return fmt.Errorf("at least one notification email is required")
	}

	payload := sendEmailRequest{
		From:    s.from,
		To:      s.recipients,
		Subject: fmt.Sprintf("New Lead Submitted - %s", lead.Company),
		Text:    buildLeadNotificationBody(lead, hubSpotContactID, submittedAt),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal Resend request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendBaseURL+"/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create Resend request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call Resend API: %w", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Resend API returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	s.logger.Printf("Internal notification email sent: recipients=%s", strings.Join(s.recipients, ","))
	return nil
}

type sendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

func buildLeadNotificationBody(lead models.LeadRequest, hubSpotContactID string, submittedAt time.Time) string {
	return fmt.Sprintf(
		"Name: %s\nEmail: %s\nCompany: %s\nWebsite: %s\nPhone Number: %s\nHubSpot Contact ID: %s\nSubmission Time: %s",
		lead.Name,
		lead.Email,
		lead.Company,
		lead.Website,
		lead.PhoneNumber,
		hubSpotContactID,
		submittedAt.UTC().Format(time.RFC3339),
	)
}
