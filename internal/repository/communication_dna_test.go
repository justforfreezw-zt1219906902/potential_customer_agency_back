package repository

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
)

func validDNAJSON() map[string][]byte {
	source := `[{"name":"Evidence Snapshot","type":null,"url":null}]`
	return map[string][]byte{
		"tone":               []byte(`{"primary":"Technical","status":"DERIVED","sources":` + source + `}`),
		"vocabulary":         []byte(`{"status":"DERIVED","terms":[{"term":"platform","sources":[]}]}`),
		"value_propositions": []byte(`[{"quote":"Built for scale","status":"SOURCE_BACKED","sources":[]}]`),
		"problem_framing":    []byte(`{"description":null,"quote":null,"status":"DERIVED","sources":[]}`),
		"proof_style":        []byte(`{"primary":null,"secondary":null,"description":null,"status":"DERIVED","sources":[]}`),
		"cta_patterns":       []byte(`{"style":null,"description":null,"examples":[],"status":"DERIVED","sources":[]}`),
		"recurring_phrases":  []byte(`[{"quote":"Built for scale","description":null,"status":"DERIVED","sources":[]}]`),
		"do_rules":           []byte(`["Be precise"]`), "dont_rules": []byte(`[]`),
	}
}

func parseTestDNA(t *testing.T, fields map[string][]byte) (*models.CommunicationDNA, error) {
	t.Helper()
	id, account := uuid.New(), uuid.New()
	return parseCommunicationDNA(id, account, fields["tone"], fields["vocabulary"], fields["value_propositions"], fields["problem_framing"], fields["proof_style"], fields["cta_patterns"], fields["recurring_phrases"], fields["do_rules"], fields["dont_rules"], time.Date(2026, 8, 10, 10, 0, 0, 0, time.UTC))
}

func TestParseCommunicationDNAValidCanonicalPayload(t *testing.T) {
	dna, err := parseTestDNA(t, validDNAJSON())
	if err != nil {
		t.Fatal(err)
	}
	if dna.Tone.Status != "DERIVED" || len(dna.Vocabulary.Terms) != 1 || len(dna.CTAPatterns.Examples) != 0 || dna.Tone.Sources[0].URL != nil {
		t.Fatalf("unexpected DNA: %+v", dna)
	}
}

func TestParseCommunicationDNARejectsRequiredArrayShapes(t *testing.T) {
	cases := []struct{ name, field, value string }{
		{"terms null", "vocabulary", `{"status":"DERIVED","terms":null}`},
		{"terms missing", "vocabulary", `{"status":"DERIVED"}`},
		{"value propositions null", "value_propositions", `null`},
		{"cta examples null", "cta_patterns", `{"status":"DERIVED","sources":[],"examples":null}`},
		{"cta examples missing", "cta_patterns", `{"status":"DERIVED","sources":[]}`},
		{"recurring phrases null", "recurring_phrases", `null`},
		{"do rules null", "do_rules", `null`},
		{"sources null", "tone", `{"status":"DERIVED","sources":null}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fields := validDNAJSON()
			fields[tc.field] = []byte(tc.value)
			if _, err := parseTestDNA(t, fields); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestParseCommunicationDNARejectsInvalidStatusAndRuleType(t *testing.T) {
	fields := validDNAJSON()
	fields["tone"] = []byte(`{"status":"SOURCE-BACKED","sources":[]}`)
	if _, err := parseTestDNA(t, fields); err == nil {
		t.Fatal("expected invalid status error")
	}
	fields = validDNAJSON()
	fields["do_rules"] = []byte(`["valid",123]`)
	if _, err := parseTestDNA(t, fields); err == nil {
		t.Fatal("expected invalid rule error")
	}
}

func TestDNASourceNullableFieldsSerializeAsNull(t *testing.T) {
	dna, err := parseTestDNA(t, validDNAJSON())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(dna.Tone.Sources[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"name":"Evidence Snapshot","type":null,"url":null}` {
		t.Fatalf("unexpected source JSON: %s", raw)
	}
}
