package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestLeadRequestPublicContract(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"with context", `{"firstName":"Tom","familyName":"Zhao","company":"Example Company","workEmail":"tom@example.com","context":"Interested in discussing the product."}`},
		{"without context", `{"firstName":"Tom","familyName":"Zhao","company":"Example Company","workEmail":"tom@example.com"}`},
		{"empty context", `{"firstName":"Tom","familyName":"Zhao","company":"Example Company","workEmail":"tom@example.com","context":""}`},
		{"submitted owner ignored", `{"firstName":"Tom","familyName":"Zhao","company":"Example Company","workEmail":"tom@example.com","owner":"999999"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var request LeadRequest
			if err := json.Unmarshal([]byte(tc.body), &request); err != nil {
				t.Fatal(err)
			}
			if err := request.Validate(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLeadRequestKeepsFamilyNameAndRemovesOwner(t *testing.T) {
	typeOfRequest := reflect.TypeOf(LeadRequest{})
	if _, ok := typeOfRequest.FieldByName("Owner"); ok {
		t.Fatal("LeadRequest must not expose Owner")
	}
	familyName, ok := typeOfRequest.FieldByName("FamilyName")
	if !ok || familyName.Tag.Get("json") != "familyName" {
		t.Fatalf("familyName JSON contract changed: %+v", familyName)
	}

	raw, err := json.Marshal(LeadRequest{FirstName: "Tom", FamilyName: "Zhao", Company: "Example", WorkEmail: "tom@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "owner") || strings.Contains(string(raw), "lastName") {
		t.Fatalf("unexpected public field: %s", raw)
	}
}

func TestLeadRequestStillRequiresFamilyName(t *testing.T) {
	request := LeadRequest{FirstName: "Tom", Company: "Example", WorkEmail: "tom@example.com"}
	if err := request.Validate(); err == nil || err.Error() != "familyName is required" {
		t.Fatalf("error=%v", err)
	}
}
