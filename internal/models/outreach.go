package models

import "github.com/google/uuid"

type OutreachGenerationRequest struct {
	Persona        string        `json:"persona"`
	AnchorSignalID uuid.UUID     `json:"anchorSignalId"`
	Parts          []string      `json:"parts"`
	CurrentDraft   OutreachDraft `json:"currentDraft"`
}
type OutreachDraft struct {
	Subject string `json:"subject"`
	Opening string `json:"opening"`
	Value   string `json:"value"`
	CTA     string `json:"cta"`
}
type OutreachGenerationResponse struct {
	GeneratedParts OutreachGeneratedParts `json:"generatedParts"`
	Traceability   OutreachTraceability   `json:"traceability"`
}
type OutreachGeneratedParts struct {
	Subject *string `json:"subject,omitempty"`
	Opening *string `json:"opening,omitempty"`
	Value   *string `json:"value,omitempty"`
	CTA     *string `json:"cta,omitempty"`
}
type OutreachTraceability struct {
	AnchorSignalID       uuid.UUID   `json:"anchorSignalId"`
	SupportingSignalIDs  []uuid.UUID `json:"supportingSignalIds"`
	CommunicationDNAUsed bool        `json:"communicationDnaUsed"`
	AnalysisUsed         bool        `json:"analysisUsed"`
}
