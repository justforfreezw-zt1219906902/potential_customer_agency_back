package service

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/hubspot"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

type leadHubSpotStub struct {
	contact             *hubspot.Contact
	companyDomain       string
	createdCompanyName  string
	createdContact      int
	updatedContact      int
	associatedContactID string
	associatedCompanyID string
}

func (s *leadHubSpotStub) FindCompanyByDomain(_ context.Context, domain string) (*hubspot.Company, error) {
	s.companyDomain = domain
	return nil, nil
}
func (s *leadHubSpotStub) CreateCompany(_ context.Context, name, _ string) (hubspot.Company, error) {
	s.createdCompanyName = name
	return hubspot.Company{ID: "company-1"}, nil
}
func (s *leadHubSpotStub) FindContactByEmail(context.Context, string) (*hubspot.Contact, error) {
	return s.contact, nil
}
func (s *leadHubSpotStub) CreateContact(context.Context, models.LeadRequest) (hubspot.Contact, error) {
	s.createdContact++
	return hubspot.Contact{ID: "created-contact"}, nil
}
func (s *leadHubSpotStub) UpdateContact(_ context.Context, id string, _ models.LeadRequest) (hubspot.Contact, error) {
	s.updatedContact++
	return hubspot.Contact{ID: id}, nil
}
func (s *leadHubSpotStub) AssociateContactToPrimaryCompany(_ context.Context, contactID, companyID string) error {
	s.associatedContactID, s.associatedCompanyID = contactID, companyID
	return nil
}

type leadEmailStub struct {
	lead models.LeadRequest
	err  error
}

func (s *leadEmailStub) SendLeadNotification(_ context.Context, lead models.LeadRequest, _ time.Time) error {
	s.lead = lead
	return s.err
}

func TestLeadServicePreservesCreateAndUpdateFlows(t *testing.T) {
	for _, tc := range []struct {
		name            string
		contact         *hubspot.Contact
		wantCreateCount int
		wantUpdateCount int
		wantContactID   string
	}{
		{"create", nil, 1, 0, "created-contact"},
		{"update", &hubspot.Contact{ID: "existing-contact"}, 0, 1, "existing-contact"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hubSpot := &leadHubSpotStub{contact: tc.contact}
			email := &leadEmailStub{}
			lead := models.LeadRequest{FirstName: "Tom", FamilyName: "Zhao", Company: "Example", WorkEmail: "tom@example.com", Context: "Useful context"}
			response, err := NewLeadService(hubSpot, email, log.New(io.Discard, "", 0)).CreateLead(context.Background(), lead)
			if err != nil {
				t.Fatal(err)
			}
			if hubSpot.companyDomain != "example.com" || hubSpot.createdCompanyName != "Example" {
				t.Fatalf("company flow=%+v", hubSpot)
			}
			if hubSpot.createdContact != tc.wantCreateCount || hubSpot.updatedContact != tc.wantUpdateCount {
				t.Fatalf("contact flow=%+v", hubSpot)
			}
			if hubSpot.associatedContactID != tc.wantContactID || hubSpot.associatedCompanyID != "company-1" {
				t.Fatalf("association=%+v", hubSpot)
			}
			if email.lead.Context != "Useful context" || response.HubSpotContactID != tc.wantContactID || response.Message != "lead submitted successfully" {
				t.Fatalf("email=%+v response=%+v", email.lead, response)
			}
		})
	}
}

func TestLeadServiceNotificationFailureRemainsNonBlocking(t *testing.T) {
	hubSpot := &leadHubSpotStub{}
	email := &leadEmailStub{err: errors.New("resend unavailable")}
	lead := models.LeadRequest{FirstName: "Tom", FamilyName: "Zhao", Company: "Example", WorkEmail: "tom@example.com"}
	response, err := NewLeadService(hubSpot, email, log.New(io.Discard, "", 0)).CreateLead(context.Background(), lead)
	if err != nil || response.HubSpotContactID != "created-contact" {
		t.Fatalf("response=%+v err=%v", response, err)
	}
}
