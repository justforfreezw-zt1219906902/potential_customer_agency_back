package models

import (
	"fmt"
	"net/mail"
	"net/url"
	"strings"
)

type LeadRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Company     string `json:"company"`
	Website     string `json:"website"`
	PhoneNumber string `json:"phoneNumber"`
	Owner       string `json:"owner"`
}

func (r LeadRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(r.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return fmt.Errorf("email must be valid")
	}
	if strings.TrimSpace(r.Company) == "" {
		return fmt.Errorf("company is required")
	}
	if strings.TrimSpace(r.Website) == "" {
		return fmt.Errorf("website is required")
	}
	parsedWebsite, err := url.ParseRequestURI(r.Website)
	if err != nil || parsedWebsite.Scheme == "" || parsedWebsite.Host == "" {
		return fmt.Errorf("website must be a valid URL")
	}
	return nil
}

func (r LeadRequest) CompanyDomain() (string, error) {
	website := strings.TrimSpace(r.Website)
	parsedWebsite, err := url.ParseRequestURI(website)
	if err != nil || parsedWebsite.Host == "" {
		return "", fmt.Errorf("website must be a valid URL")
	}

	return strings.TrimPrefix(parsedWebsite.Hostname(), "www."), nil
}

type LeadResponse struct {
	Message          string `json:"message"`
	HubSpotContactID string `json:"hubspot_contact_id,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
