package models

import (
	"fmt"
	"net/mail"
	"strings"
)

type LeadRequest struct {
	FirstName  string `json:"firstName"`
	FamilyName string `json:"familyName"`
	Company    string `json:"company"`
	WorkEmail  string `json:"workEmail"`
}

func (r LeadRequest) Validate() error {
	if strings.TrimSpace(r.FirstName) == "" {
		return fmt.Errorf("firstName is required")
	}
	if strings.TrimSpace(r.FamilyName) == "" {
		return fmt.Errorf("familyName is required")
	}
	if strings.TrimSpace(r.WorkEmail) == "" {
		return fmt.Errorf("workEmail is required")
	}
	if _, err := mail.ParseAddress(r.WorkEmail); err != nil {
		return fmt.Errorf("workEmail must be valid")
	}
	if strings.TrimSpace(r.Company) == "" {
		return fmt.Errorf("company is required")
	}
	return nil
}

func (r LeadRequest) CompanyDomain() (string, error) {
	email := strings.TrimSpace(r.WorkEmail)
	_, err := mail.ParseAddress(email)
	if err != nil {
		return "", fmt.Errorf("workEmail must be valid")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("workEmail must include a domain")
	}

	return strings.ToLower(strings.TrimSpace(parts[1])), nil
}

type LeadResponse struct {
	Message          string `json:"message"`
	HubSpotContactID string `json:"hubspot_contact_id,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
