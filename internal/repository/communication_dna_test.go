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

func TestParseCommunicationDNANormalizesLegacyStringSources(t *testing.T) {
	fields := validDNAJSON()
	fields["tone"] = []byte(`{"primary":"Technical","status":"DERIVED","sources":["nvidia.com",{"name":"Evidence Snapshot","type":null,"url":null}]}`)

	dna, err := parseTestDNA(t, fields)
	if err != nil {
		t.Fatal(err)
	}
	if len(dna.Tone.Sources) != 2 || dna.Tone.Sources[0].Name != nil || dna.Tone.Sources[0].Type != nil || dna.Tone.Sources[0].URL == nil || *dna.Tone.Sources[0].URL != "nvidia.com" {
		t.Fatalf("legacy source was not normalized: %+v", dna.Tone.Sources)
	}
	if dna.Tone.Sources[1].Name == nil || *dna.Tone.Sources[1].Name != "Evidence Snapshot" {
		t.Fatalf("canonical source changed: %+v", dna.Tone.Sources[1])
	}

	raw, err := json.Marshal(dna.Tone.Sources[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"name":null,"type":null,"url":"nvidia.com"}` {
		t.Fatalf("unexpected normalized API source: %s", raw)
	}
}

func TestParseCommunicationDNARejectsInvalidSourceScalar(t *testing.T) {
	fields := validDNAJSON()
	fields["tone"] = []byte(`{"primary":"Technical","status":"DERIVED","sources":[123]}`)
	if _, err := parseTestDNA(t, fields); err == nil {
		t.Fatal("expected invalid source scalar error")
	}
}

func TestParseCommunicationDNANormalizesLegacyDatabaseShape(t *testing.T) {
	fields := validDNAJSON()
	fields["tone"] = []byte(`{"primary":"Authoritative","status":"DERIVED","sources":["nvidia.com"]}`)
	fields["vocabulary"] = []byte(`{"status":"SOURCE_BACKED","terms":[{"word":"AI factory","context":"Core","frequency":"high","source":"nvidia.com/ai","sourceType":"Product Page"}]}`)
	fields["value_propositions"] = []byte(`[{"prop":"Built for scale","source":"nvidia.com/product","status":"SOURCE_BACKED"}]`)
	fields["problem_framing"] = []byte(`{"description":"A platform shift","source":"nvidia.com, developer.nvidia.com","status":"SOURCE_BACKED"}`)
	fields["proof_style"] = []byte(`{"primary":"Quantified claims","source":"nvidia.com/proof","status":"SOURCE_BACKED"}`)
	fields["cta_patterns"] = []byte(`{"style":"Explore","examples":["Learn More"],"source":"nvidia.com/learn","status":"SOURCE_BACKED"}`)
	fields["recurring_phrases"] = []byte(`["accelerated computing"]`)

	dna, err := parseTestDNA(t, fields)
	if err != nil {
		t.Fatal(err)
	}
	if dna.Vocabulary.Terms[0].Term != "AI factory" || *dna.Vocabulary.Terms[0].Sources[0].URL != "nvidia.com/ai" || *dna.Vocabulary.Terms[0].Sources[0].Type != "Product Page" {
		t.Fatalf("legacy vocabulary was not normalized: %+v", dna.Vocabulary.Terms[0])
	}
	if dna.ValuePropositions[0].Quote != "Built for scale" || *dna.ValuePropositions[0].Sources[0].URL != "nvidia.com/product" {
		t.Fatalf("legacy value proposition was not normalized: %+v", dna.ValuePropositions[0])
	}
	if len(dna.ProblemFraming.Sources) != 2 || *dna.ProofStyle.Sources[0].URL != "nvidia.com/proof" || *dna.CTAPatterns.Sources[0].URL != "nvidia.com/learn" {
		t.Fatalf("legacy singular sources were not normalized: %+v %+v %+v", dna.ProblemFraming.Sources, dna.ProofStyle.Sources, dna.CTAPatterns.Sources)
	}
	if dna.RecurringPhrases[0].Quote != "accelerated computing" || dna.RecurringPhrases[0].Status != "INSUFFICIENT_DATA" || dna.RecurringPhrases[0].Sources == nil {
		t.Fatalf("legacy phrase was not normalized conservatively: %+v", dna.RecurringPhrases[0])
	}
}
