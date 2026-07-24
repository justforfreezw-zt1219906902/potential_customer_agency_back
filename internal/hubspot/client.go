package hubspot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

const (
	defaultBaseURL                    = "https://api.hubapi.com"
	contactToCompanyAssociationTypeID = 1
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	logger     *log.Logger
}

func NewClient(token string, logger *log.Logger) *Client {
	return &Client{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		token:  token,
		logger: logger,
	}
}

func (c *Client) FindCompanyByDomain(ctx context.Context, domain string) (*Company, error) {
	payload := searchRequest{
		FilterGroups: []filterGroup{
			{
				Filters: []filter{
					{
						PropertyName: "domain",
						Operator:     "EQ",
						Value:        domain,
					},
				},
			},
		},
		Limit:      1,
		Properties: []string{"domain", "name"},
	}

	body, err := c.doJSON(ctx, http.MethodPost, "/crm/v3/objects/companies/search", payload)
	if err != nil {
		return nil, err
	}

	var response companySearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode HubSpot company search response: %w", err)
	}
	if len(response.Results) == 0 {
		return nil, nil
	}

	companyID := response.Results[0].ID
	c.logger.Printf("HubSpot company found: id=%s domain=%s", companyID, domain)
	return &Company{ID: companyID}, nil
}

func (c *Client) CreateCompany(ctx context.Context, name string, domain string) (Company, error) {
	payload := createCompanyRequest{
		Properties: map[string]string{
			"name":   name,
			"domain": domain,
		},
	}

	body, err := c.doJSON(ctx, http.MethodPost, "/crm/v3/objects/companies", payload)
	if err != nil {
		return Company{}, err
	}

	var response createCompanyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return Company{}, fmt.Errorf("decode HubSpot company create response: %w", err)
	}
	if strings.TrimSpace(response.ID) == "" {
		return Company{}, fmt.Errorf("HubSpot company create response did not include id")
	}

	c.logger.Printf("HubSpot company created: id=%s domain=%s", response.ID, domain)
	return Company{ID: response.ID}, nil
}

func (c *Client) FindContactByEmail(ctx context.Context, email string) (*Contact, error) {
	payload := searchRequest{
		FilterGroups: []filterGroup{
			{
				Filters: []filter{
					{
						PropertyName: "email",
						Operator:     "EQ",
						Value:        email,
					},
				},
			},
		},
		Limit:      1,
		Properties: []string{"email"},
	}

	body, err := c.doJSON(ctx, http.MethodPost, "/crm/v3/objects/contacts/search", payload)
	if err != nil {
		return nil, err
	}

	var response contactSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode HubSpot contact search response: %w", err)
	}
	if len(response.Results) == 0 {
		return nil, nil
	}

	contactID := response.Results[0].ID
	c.logger.Printf("HubSpot contact found: id=%s email=%s", contactID, email)
	return &Contact{ID: contactID}, nil
}

func (c *Client) CreateContact(ctx context.Context, lead models.LeadRequest) (Contact, error) {
	payload := contactPropertiesRequest{
		Properties: contactProperties(lead),
	}

	body, err := c.doJSON(ctx, http.MethodPost, "/crm/v3/objects/contacts", payload)
	if err != nil {
		return Contact{}, err
	}

	var response contactObjectResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return Contact{}, fmt.Errorf("decode HubSpot contact create response: %w", err)
	}
	if strings.TrimSpace(response.ID) == "" {
		return Contact{}, fmt.Errorf("HubSpot contact create response did not include id")
	}

	c.logger.Printf("HubSpot contact created: id=%s email=%s", response.ID, lead.WorkEmail)
	return Contact{ID: response.ID}, nil
}

func (c *Client) UpdateContact(ctx context.Context, contactID string, lead models.LeadRequest) (Contact, error) {
	payload := contactPropertiesRequest{
		Properties: contactProperties(lead),
	}

	body, err := c.doJSON(ctx, http.MethodPatch, "/crm/v3/objects/contacts/"+url.PathEscape(contactID), payload)
	if err != nil {
		return Contact{}, err
	}

	var response contactObjectResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return Contact{}, fmt.Errorf("decode HubSpot contact update response: %w", err)
	}
	if strings.TrimSpace(response.ID) == "" {
		return Contact{}, fmt.Errorf("HubSpot contact update response did not include id")
	}

	c.logger.Printf("HubSpot contact updated: id=%s email=%s", response.ID, lead.WorkEmail)
	return Contact{ID: response.ID}, nil
}

func (c *Client) AssociateContactToPrimaryCompany(ctx context.Context, contactID string, companyID string) error {
	payload := associationRequest{
		FromObjectID: contactID,
		ToObjectID:   companyID,
		Category:     "HUBSPOT_DEFINED",
		DefinitionID: strconv.Itoa(contactToCompanyAssociationTypeID),
	}

	if _, err := c.doJSON(ctx, http.MethodPut, "/crm-associations/v1/associations", payload); err != nil {
		return err
	}

	c.logger.Printf("HubSpot contact associated to primary company: contact_id=%s company_id=%s", contactID, companyID)
	return nil
}

func (c *Client) doJSON(ctx context.Context, method string, path string, payload any) ([]byte, error) {
	statusCode, body, err := c.doJSONWithStatus(ctx, method, path, payload)
	if err != nil {
		return nil, err
	}
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("HubSpot API returned status %d: %s", statusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}

func (c *Client) doJSONWithStatus(ctx context.Context, method string, path string, payload any) (int, []byte, error) {
	var requestBody io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, fmt.Errorf("marshal HubSpot request: %w", err)
		}
		requestBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, requestBody)
	if err != nil {
		return 0, nil, fmt.Errorf("create HubSpot request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("call HubSpot API: %w", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, responseBody, nil
}

type searchRequest struct {
	FilterGroups []filterGroup `json:"filterGroups"`
	Limit        int           `json:"limit"`
	Properties   []string      `json:"properties"`
}

type filterGroup struct {
	Filters []filter `json:"filters"`
}

type filter struct {
	PropertyName string `json:"propertyName"`
	Operator     string `json:"operator"`
	Value        string `json:"value"`
}

type companySearchResponse struct {
	Results []companySearchResult `json:"results"`
}

type companySearchResult struct {
	ID string `json:"id"`
}

type createCompanyRequest struct {
	Properties map[string]string `json:"properties"`
}

type createCompanyResponse struct {
	ID string `json:"id"`
}

type contactPropertiesRequest struct {
	Properties map[string]string `json:"properties"`
}

type contactSearchResponse struct {
	Results []contactSearchResult `json:"results"`
}

type contactSearchResult struct {
	ID string `json:"id"`
}

type contactObjectResponse struct {
	ID string `json:"id"`
}

type associationRequest struct {
	FromObjectID string `json:"fromObjectId"`
	ToObjectID   string `json:"toObjectId"`
	Category     string `json:"category"`
	DefinitionID string `json:"definitionId"`
}

func contactProperties(lead models.LeadRequest) map[string]string {
	properties := map[string]string{
		"email":     lead.WorkEmail,
		"firstname": lead.FirstName,
		"lastname":  lead.FamilyName,
	}

	for key, value := range properties {
		if strings.TrimSpace(value) == "" {
			delete(properties, key)
		}
	}

	return properties
}
