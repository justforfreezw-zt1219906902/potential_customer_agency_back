package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	apperrors "github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/outreach"
)

type outreachReaderStub struct {
	accountReaderStub
	data outreach.ContextData
	err  error
}

func (s outreachReaderStub) LoadOutreachContext(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (outreach.ContextData, error) {
	return s.data, s.err
}

type outreachGeneratorStub struct {
	output outreach.Output
	err    error
	called bool
	input  outreach.Input
}

func (s *outreachGeneratorStub) Generate(_ context.Context, input outreach.Input) (outreach.Output, error) {
	s.called = true
	s.input = input
	return s.output, s.err
}

func outreachRequest(parts ...string) models.OutreachGenerationRequest {
	return models.OutreachGenerationRequest{Persona: "marketing", AnchorSignalID: uuid.MustParse("00000000-0000-0000-0000-000000000801"), Parts: parts, CurrentDraft: models.OutreachDraft{Subject: "draft subject", Opening: "draft opening", Value: "draft value", CTA: "draft cta"}}
}
func outreachContext(active bool) outreach.ContextData {
	return outreach.ContextData{AccountFound: true, AnchorFound: true, AnchorActive: active, AnchorSignal: outreach.Signal{ID: "00000000-0000-0000-0000-000000000801"}, SupportingSignals: []outreach.Signal{}, TargetAccount: outreach.TargetAccount{ID: "00000000-0000-0000-0000-000000000802"}}
}

func TestOutreachServiceRejectsInactiveAnchorWithoutCallingGenerator(t *testing.T) {
	generator := &outreachGeneratorStub{}
	_, err := NewOutreachService(outreachReaderStub{data: outreachContext(false)}, uuid.New(), generator).Generate(context.Background(), uuid.New(), outreachRequest("subject"))
	appErr, ok := apperrors.AsAppError(err)
	if !ok || appErr.Status != 400 || generator.called {
		t.Fatalf("err=%v called=%v", err, generator.called)
	}
}

func TestOutreachServiceReturnsOnlyRequestedPartAndPreservesDraft(t *testing.T) {
	generator := &outreachGeneratorStub{output: outreach.Output{GeneratedParts: outreach.Draft{Subject: "new subject", Opening: "new opening", Value: "new value", CTA: "new cta"}}}
	response, err := NewOutreachService(outreachReaderStub{data: outreachContext(true)}, uuid.New(), generator).Generate(context.Background(), uuid.New(), outreachRequest("subject"))
	if err != nil || response.GeneratedParts.Subject == nil || response.GeneratedParts.Opening != nil || response.GeneratedParts.Value != nil || response.GeneratedParts.CTA != nil {
		t.Fatalf("response=%+v err=%v", response, err)
	}
	if generator.input.CurrentDraft.Subject != "draft subject" || generator.input.CurrentDraft.Opening != "draft opening" || generator.input.CurrentDraft.Value != "draft value" || generator.input.CurrentDraft.CTA != "draft cta" {
		t.Fatalf("draft not preserved: %+v", generator.input.CurrentDraft)
	}
}

func TestOutreachServiceProviderFailuresReturn502(t *testing.T) {
	for _, output := range []outreach.Output{{}, {GeneratedParts: outreach.Draft{Subject: "   "}}} {
		generator := &outreachGeneratorStub{output: output}
		_, err := NewOutreachService(outreachReaderStub{data: outreachContext(true)}, uuid.New(), generator).Generate(context.Background(), uuid.New(), outreachRequest("subject"))
		appErr, ok := apperrors.AsAppError(err)
		if !ok || appErr.Status != 502 {
			t.Fatalf("err=%v", err)
		}
	}
}
