package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/outreach"
)

type outreachReaderStub struct {
	accountReaderStub
	data    outreach.ContextData
	err     error
	ctxSink *context.Context
}

func (s outreachReaderStub) LoadOutreachContext(ctx context.Context, _ uuid.UUID, _ uuid.UUID, _ uuid.UUID) (outreach.ContextData, error) {
	if s.ctxSink != nil {
		*s.ctxSink = ctx
	}
	return s.data, s.err
}

type outreachGeneratorStub struct {
	output outreach.Output
	err    error
	called bool
	input  outreach.Input
	ctx    context.Context
}

func (s *outreachGeneratorStub) Generate(ctx context.Context, input outreach.Input) (outreach.Output, error) {
	s.called = true
	s.input = input
	s.ctx = ctx
	return s.output, s.err
}

func TestOutreachServicePassesSameContextToRepositoryAndGenerator(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var repositoryCtx context.Context
	reader := outreachReaderStub{data: outreachContext(true), ctxSink: &repositoryCtx}
	generator := &outreachGeneratorStub{output: outreach.Output{GeneratedParts: outreach.Draft{Subject: "generated"}}}
	_, err := NewOutreachService(reader, uuid.New(), generator).Generate(ctx, uuid.New(), outreachRequest("subject"))
	if err != nil {
		t.Fatal(err)
	}
	if repositoryCtx != ctx || generator.ctx != ctx {
		t.Fatal("Outreach context was replaced between service dependencies")
	}
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

func TestOutreachServiceValidationAndProviderError(t *testing.T) {
	ids := []uuid.UUID{uuid.MustParse("00000000-0000-0000-0000-000000000911"), uuid.MustParse("00000000-0000-0000-0000-000000000912")}
	cases := []struct {
		name   string
		data   outreach.ContextData
		status int
	}{{"missing account", outreach.ContextData{}, 404}, {"missing anchor", outreach.ContextData{AccountFound: true}, 400}}
	for _, tc := range cases {
		generator := &outreachGeneratorStub{}
		_, err := NewOutreachService(outreachReaderStub{data: tc.data}, uuid.New(), generator).Generate(context.Background(), ids[0], outreachRequest("subject"))
		appErr, ok := apperrors.AsAppError(err)
		if !ok || appErr.Status != tc.status || generator.called {
			t.Fatalf("%s err=%v called=%v", tc.name, err, generator.called)
		}
	}
	runtimeErr := errors.New("provider failed")
	generator := &outreachGeneratorStub{err: runtimeErr}
	_, err := NewOutreachService(outreachReaderStub{data: outreachContext(true)}, uuid.New(), generator).Generate(context.Background(), ids[0], outreachRequest("subject"))
	appErr, ok := apperrors.AsAppError(err)
	if !ok || appErr.Status != 502 || appErr.Message != "outreach generation unavailable" {
		t.Fatalf("provider error=%v", err)
	}
}

func TestOutreachServicePartialGenerationAndTraceability(t *testing.T) {
	output := outreach.Output{GeneratedParts: outreach.Draft{Subject: "subject", Opening: "opening", Value: "value", CTA: "cta"}}
	supporting := []outreach.Signal{{ID: "00000000-0000-0000-0000-000000000922", EvidenceStatus: "SOURCE_BACKED", Verified: true, ScoreEligible: true}, {ID: "00000000-0000-0000-0000-000000000923"}}
	data := outreachContext(true)
	data.SupportingSignals = supporting
	data.CommunicationDNA = json.RawMessage(`{"tone":{"primary":"technical"}}`)
	tier := "Tier 1"
	data.LatestAnalysis = &outreach.Analysis{Tier: &tier}
	cases := []struct {
		part  string
		check func(models.OutreachGeneratedParts) bool
	}{{"subject", func(p models.OutreachGeneratedParts) bool {
		return p.Subject != nil && p.Opening == nil && p.Value == nil && p.CTA == nil
	}}, {"opening", func(p models.OutreachGeneratedParts) bool {
		return p.Subject == nil && p.Opening != nil && p.Value == nil && p.CTA == nil
	}}, {"value", func(p models.OutreachGeneratedParts) bool {
		return p.Subject == nil && p.Opening == nil && p.Value != nil && p.CTA == nil
	}}, {"cta", func(p models.OutreachGeneratedParts) bool {
		return p.Subject == nil && p.Opening == nil && p.Value == nil && p.CTA != nil
	}}, {"all", func(p models.OutreachGeneratedParts) bool {
		return p.Subject != nil && p.Opening != nil && p.Value != nil && p.CTA != nil
	}}}
	for _, tc := range cases {
		generator := &outreachGeneratorStub{output: output}
		parts := []string{tc.part}
		if tc.part == "all" {
			parts = []string{"subject", "opening", "value", "cta"}
		}
		response, err := NewOutreachService(outreachReaderStub{data: data}, uuid.New(), generator).Generate(context.Background(), uuid.New(), outreachRequest(parts...))
		if err != nil || !tc.check(response.GeneratedParts) || !generator.called || generator.input.CurrentDraft != (outreach.Draft{Subject: "draft subject", Opening: "draft opening", Value: "draft value", CTA: "draft cta"}) {
			t.Fatalf("%s response=%+v err=%v input=%+v", tc.part, response, err, generator.input)
		}
		if tc.part == "all" && (response.Traceability.AnchorSignalID.String() != "00000000-0000-0000-0000-000000000801" || len(response.Traceability.SupportingSignalIDs) != 2 || !response.Traceability.CommunicationDNAUsed || !response.Traceability.AnalysisUsed) {
			t.Fatalf("traceability=%+v", response.Traceability)
		}
	}
}
