package hubspot

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestConfiguredOwnerIsSentForContactCreateAndUpdate(t *testing.T) {
	requests := make([]contactPropertiesRequest, 0, 2)
	methods := make([]string, 0, 2)
	paths := make([]string, 0, 2)
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var request contactPropertiesRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return nil, err
		}
		requests = append(requests, request)
		methods = append(methods, r.Method)
		paths = append(paths, r.URL.Path)
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"id":"contact-1"}`)), Header: make(http.Header)}, nil
	})

	client := NewClient("test-token", "123456", log.New(io.Discard, "", 0))
	client.httpClient = &http.Client{Transport: transport}
	lead := models.LeadRequest{FirstName: "Tom", FamilyName: "Zhao", Company: "Example", WorkEmail: "tom@example.com", Context: "private form context"}

	if _, err := client.CreateContact(context.Background(), lead); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateContact(context.Background(), "contact-1", lead); err != nil {
		t.Fatal(err)
	}
	if len(requests) != 2 {
		t.Fatalf("requests=%d", len(requests))
	}
	if methods[0] != http.MethodPost || paths[0] != "/crm/v3/objects/contacts" || methods[1] != http.MethodPatch || paths[1] != "/crm/v3/objects/contacts/contact-1" {
		t.Fatalf("requests=%v %v", methods, paths)
	}
	for i, request := range requests {
		if request.Properties["hubspot_owner_id"] != "123456" {
			t.Fatalf("request %d owner=%q", i, request.Properties["hubspot_owner_id"])
		}
		if request.Properties["email"] != lead.WorkEmail || request.Properties["firstname"] != lead.FirstName || request.Properties["lastname"] != lead.FamilyName {
			t.Fatalf("request %d properties=%v", i, request.Properties)
		}
		if _, ok := request.Properties["context"]; ok || len(request.Properties) != 4 {
			t.Fatalf("context or another public field leaked to HubSpot: %v", request.Properties)
		}
	}
}
