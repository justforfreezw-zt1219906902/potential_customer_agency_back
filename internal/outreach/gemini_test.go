package outreach

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"google.golang.org/genai"
)

type geminiCallerStub struct {
	response *genai.GenerateContentResponse
	err      error
	config   *genai.GenerateContentConfig
	payload  string
	model    string
}

func (s *geminiCallerStub) GenerateContent(_ context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	s.model = model
	s.config = config
	if len(contents) > 0 && len(contents[0].Parts) > 0 {
		s.payload = contents[0].Parts[0].Text
	}
	return s.response, s.err
}
func geminiResponse(value string) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{Candidates: []*genai.Candidate{{Content: &genai.Content{Parts: []*genai.Part{{Text: value}}}}}}
}
func testInput(parts ...Part) Input {
	return Input{RequestedParts: parts, Persona: "marketing", CurrentDraft: Draft{Subject: "manual subject", Opening: "manual opening", Value: "manual value", CTA: "manual cta"}, SellerCompany: SellerCompany{Products: json.RawMessage(`[{"name":"Camtasia Editor","desc":"Screen recording + professional video editing"}]`), ValuePropositions: json.RawMessage(`[{"text":"structured value"}]`), BuyerPersonas: json.RawMessage(`[{"name":"marketing leader"}]`), CommunicationDNA: json.RawMessage(`{"tone":{"primary":"technical"}}`)}, TargetAccount: TargetAccount{ID: "account", Name: "Example"}, AnchorSignal: Signal{ID: "anchor"}, SupportingSignals: []Signal{{ID: "support"}}}
}

func TestGeminiGeneratorBuildsExactSchemasAndStructuredPayload(t *testing.T) {
	for _, tc := range []struct {
		parts    []Part
		required []string
	}{{[]Part{Subject}, []string{"subject"}}, {[]Part{Opening, CTA}, []string{"opening", "cta"}}, {[]Part{Subject, Opening, Value, CTA}, []string{"subject", "opening", "value", "cta"}}} {
		result := map[string]string{}
		for _, part := range tc.parts {
			result[string(part)] = string(part)
		}
		raw, _ := json.Marshal(result)
		caller := &geminiCallerStub{response: geminiResponse(string(raw))}
		_, err := NewGeminiOutreachGeneratorWithCaller(caller, "gemini-3.6-flash").Generate(context.Background(), testInput(tc.parts...))
		if err != nil {
			t.Fatal(err)
		}
		if len(caller.config.ResponseSchema.Required) != len(tc.required) {
			t.Fatalf("required=%v", caller.config.ResponseSchema.Required)
		}
		for _, key := range tc.required {
			if caller.config.ResponseSchema.Properties[key].Description == "" {
				t.Fatalf("missing description for %s", key)
			}
		}
		if caller.model != "gemini-3.6-flash" || caller.config.Tools != nil {
			t.Fatalf("model/tools mismatch: %s %#v", caller.model, caller.config.Tools)
		}
		for i, v := range tc.required {
			if caller.config.ResponseSchema.Required[i] != v {
				t.Fatalf("required=%v", caller.config.ResponseSchema.Required)
			}
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(caller.payload), &payload); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"persona", "currentDraft", "sellerCompany", "targetAccount", "anchorSignal", "supportingSignals"} {
			if _, ok := payload[key]; !ok {
				t.Fatalf("payload missing %s", key)
			}
		}
		if !json.Valid([]byte(caller.payload)) {
			t.Fatal("payload is not JSON")
		}
		if !contains(caller.payload, "Camtasia Editor") || !contains(caller.payload, "professional video editing") {
			t.Fatal("rich seller JSON was lost")
		}
	}
}

func TestGeminiSystemInstructionContainsStableEmailRules(t *testing.T) {
	for _, phrase := range []string{"lowercase", "2-4 words", "does not force or assume a meeting", "below 100 words", "Never invent"} {
		if !contains(geminiSystemInstruction, phrase) {
			t.Fatalf("system instruction missing %q", phrase)
		}
	}
}

type blockingGeminiCaller struct{}

func (blockingGeminiCaller) GenerateContent(ctx context.Context, _ string, _ []*genai.Content, _ *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}
func TestGeminiGeneratorHonorsBoundedTimeout(t *testing.T) {
	generator := NewGeminiOutreachGeneratorWithCaller(blockingGeminiCaller{}, "model")
	generator.timeout = 10 * time.Millisecond
	started := time.Now()
	_, err := generator.Generate(context.Background(), testInput(Subject))
	if err == nil || time.Since(started) > time.Second {
		t.Fatalf("timeout error=%v duration=%v", err, time.Since(started))
	}
}
func contains(value, part string) bool {
	return len(value) >= len(part) && stringIndex(value, part) >= 0
}
func stringIndex(value, part string) int {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return i
		}
	}
	return -1
}
func TestGeminiGeneratorValidatesProviderResponses(t *testing.T) {
	for _, raw := range []string{`not json`, `{"opening":"missing"}`, `{"subject":"   "}`, `{"subject":"ok","extra":"bad"}`} {
		caller := &geminiCallerStub{response: geminiResponse(raw)}
		_, err := NewGeminiOutreachGeneratorWithCaller(caller, "model").Generate(context.Background(), testInput(Subject))
		if err == nil {
			t.Fatalf("expected error for %q", raw)
		}
	}
	caller := &geminiCallerStub{err: errors.New("provider failed")}
	_, err := NewGeminiOutreachGeneratorWithCaller(caller, "model").Generate(context.Background(), testInput(Subject))
	if err == nil {
		t.Fatal("expected provider error")
	}
}
