package email

import (
	"strings"
	"testing"
	"time"

	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

func TestBuildLeadNotificationBodyIncludesContext(t *testing.T) {
	lead := models.LeadRequest{FirstName: "Tom", FamilyName: "Zhao", Company: "Example", WorkEmail: "tom@example.com", Context: "Interested in account intelligence."}
	body := buildLeadNotificationBody(lead, time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC))
	if !strings.Contains(body, "Context: Interested in account intelligence.") || strings.Contains(body, "HubSpot Contact ID") || strings.Contains(body, "contact-1") {
		t.Fatalf("body=%q", body)
	}
}

func TestBuildLeadNotificationBodyUsesDeterministicEmptyContext(t *testing.T) {
	lead := models.LeadRequest{FirstName: "Tom", FamilyName: "Zhao", Company: "Example", WorkEmail: "tom@example.com", Context: "   "}
	body := buildLeadNotificationBody(lead, time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC))
	if !strings.Contains(body, "Context: (not provided)") {
		t.Fatalf("body=%q", body)
	}
}
