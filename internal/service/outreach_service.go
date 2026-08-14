package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	apperrors "github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/errors"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/models"
	"github.com/justforfreezw-zt1219906902/potential_customer_agency_back/internal/outreach"
)

type OutreachService struct {
	repository AccountReader
	companyID  uuid.UUID
	generator  outreach.Generator
}

func NewOutreachService(repository AccountReader, companyID uuid.UUID, generator outreach.Generator) *OutreachService {
	return &OutreachService{repository: repository, companyID: companyID, generator: generator}
}

func (s *OutreachService) Generate(ctx context.Context, accountID uuid.UUID, request models.OutreachGenerationRequest) (models.OutreachGenerationResponse, error) {
	data, err := s.repository.LoadOutreachContext(ctx, s.companyID, accountID, request.AnchorSignalID)
	if err != nil {
		return models.OutreachGenerationResponse{}, err
	}
	if !data.AccountFound {
		return models.OutreachGenerationResponse{}, apperrors.NotFound("account not found", nil)
	}
	if !data.AnchorFound || !data.AnchorActive {
		return models.OutreachGenerationResponse{}, apperrors.BadRequest("anchor signal must belong to account and be active", nil)
	}
	parts := make([]outreach.Part, len(request.Parts))
	for i, p := range request.Parts {
		parts[i] = outreach.Part(p)
	}
	input := outreach.Input{RequestedParts: parts, Persona: request.Persona, CurrentDraft: outreach.Draft{Subject: request.CurrentDraft.Subject, Opening: request.CurrentDraft.Opening, Value: request.CurrentDraft.Value, CTA: request.CurrentDraft.CTA}, SellerCompany: data.SellerCompany, TargetAccount: data.TargetAccount, LatestAnalysis: data.LatestAnalysis, AnchorSignal: data.AnchorSignal, SupportingSignals: data.SupportingSignals, CommunicationDNA: data.CommunicationDNA}
	output, err := s.generator.Generate(ctx, input)
	if err != nil {
		return models.OutreachGenerationResponse{}, apperrors.ExternalService("outreach generation unavailable", err)
	}
	response := models.OutreachGenerationResponse{Traceability: models.OutreachTraceability{AnchorSignalID: request.AnchorSignalID, SupportingSignalIDs: make([]uuid.UUID, 0), CommunicationDNAUsed: data.CommunicationDNA != nil, AnalysisUsed: data.LatestAnalysis != nil}}
	for _, signal := range data.SupportingSignals {
		if id, e := uuid.Parse(signal.ID); e == nil {
			response.Traceability.SupportingSignalIDs = append(response.Traceability.SupportingSignalIDs, id)
		}
	}
	for _, part := range parts {
		var value string
		switch part {
		case outreach.Subject:
			value = output.GeneratedParts.Subject
		case outreach.Opening:
			value = output.GeneratedParts.Opening
		case outreach.Value:
			value = output.GeneratedParts.Value
		case outreach.CTA:
			value = output.GeneratedParts.CTA
		}
		if strings.TrimSpace(value) == "" {
			return models.OutreachGenerationResponse{}, apperrors.ExternalService("outreach generation unavailable", fmt.Errorf("missing generated part %s", part))
		}
		switch part {
		case outreach.Subject:
			response.GeneratedParts.Subject = &value
		case outreach.Opening:
			response.GeneratedParts.Opening = &value
		case outreach.Value:
			response.GeneratedParts.Value = &value
		case outreach.CTA:
			response.GeneratedParts.CTA = &value
		}
	}
	return response, nil
}
