package outreach

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type geminiCaller interface {
	GenerateContent(context.Context, string, []*genai.Content, *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
}
type GeminiOutreachGenerator struct {
	caller geminiCaller
	model  string
}

func NewGeminiOutreachGenerator(ctx context.Context, apiKey, model string) (*GeminiOutreachGenerator, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey, Backend: genai.BackendGeminiAPI})
	if err != nil {
		return nil, err
	}
	return &GeminiOutreachGenerator{caller: client.Models, model: model}, nil
}
func NewGeminiOutreachGeneratorWithCaller(caller geminiCaller, model string) *GeminiOutreachGenerator {
	return &GeminiOutreachGenerator{caller: caller, model: model}
}

func (g *GeminiOutreachGenerator) Generate(ctx context.Context, input Input) (Output, error) {
	prompt, err := BuildOutreachEmailPrompt(input)
	if err != nil {
		return Output{}, err
	}
	schema := schemaFor(input.RequestedParts)
	config := &genai.GenerateContentConfig{ResponseMIMEType: "application/json", ResponseSchema: schema, SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: prompt.SystemInstruction}}}}
	response, err := g.caller.GenerateContent(ctx, g.model, []*genai.Content{{Role: genai.RoleUser, Parts: []*genai.Part{{Text: prompt.UserContent}}}}, config)
	if err != nil {
		return Output{}, err
	}
	if response == nil {
		return Output{}, fmt.Errorf("empty Gemini response")
	}
	var generated map[string]string
	if err := json.Unmarshal([]byte(response.Text()), &generated); err != nil {
		return Output{}, fmt.Errorf("invalid Gemini JSON: %w", err)
	}
	allowed := map[string]bool{}
	for _, part := range input.RequestedParts {
		allowed[string(part)] = true
	}
	for key := range generated {
		if !allowed[key] {
			return Output{}, fmt.Errorf("unexpected generated part %s", key)
		}
	}
	out := Output{}
	for _, part := range input.RequestedParts {
		value, ok := generated[string(part)]
		if !ok || strings.TrimSpace(value) == "" {
			return Output{}, fmt.Errorf("missing generated part %s", part)
		}
		switch part {
		case Subject:
			out.GeneratedParts.Subject = value
		case Opening:
			out.GeneratedParts.Opening = value
		case Value:
			out.GeneratedParts.Value = value
		case CTA:
			out.GeneratedParts.CTA = value
		}
	}
	return out, nil
}
func schemaFor(parts []Part) *genai.Schema {
	schema := &genai.Schema{Type: genai.TypeObject, Properties: map[string]*genai.Schema{}, Required: []string{}}
	for _, part := range parts {
		key := string(part)
		description := map[Part]string{Subject: "Lowercase concise outreach subject, ideally 2-4 words.", Opening: "One concise sentence grounded in the prospect problem, change, or signal.", Value: "One concrete sentence connecting persisted seller value to prospect context.", CTA: "Concise interest-oriented CTA that does not force a meeting."}[part]
		schema.Properties[key] = &genai.Schema{Type: genai.TypeString, Description: description}
		schema.Required = append(schema.Required, key)
	}
	return schema
}
