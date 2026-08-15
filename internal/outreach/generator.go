package outreach

import (
	"context"
	"encoding/json"
)

type Part string

const (
	Subject Part = "subject"
	Opening Part = "opening"
	Value   Part = "value"
	CTA     Part = "cta"
)

type Draft struct {
	Subject string `json:"subject"`
	Opening string `json:"opening"`
	Value   string `json:"value"`
	CTA     string `json:"cta"`
}
type SellerCompany struct {
	Name              string          `json:"name"`
	Tagline           string          `json:"tagline"`
	Website           string          `json:"website"`
	Description       string          `json:"description"`
	Products          json.RawMessage `json:"products"`
	ValuePropositions json.RawMessage `json:"valuePropositions"`
	BuyerPersonas     json.RawMessage `json:"buyerPersonas"`
	CommunicationDNA  json.RawMessage `json:"communicationDNA"`
}
type TargetAccount struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Industry    string `json:"industry"`
	Website     string `json:"website"`
	Description string `json:"description"`
}
type Analysis struct {
	Tier           *string `json:"tier"`
	NextBestAction *string `json:"nextBestAction"`
}
type Signal struct {
	ID             string        `json:"id"`
	Type           string        `json:"type"`
	Title          string        `json:"title"`
	Body           string        `json:"body"`
	Strength       string        `json:"strength"`
	Relevance      string        `json:"relevance"`
	SignalDate     string        `json:"signalDate"`
	EvidenceStatus string        `json:"evidenceStatus"`
	Verified       bool          `json:"verified"`
	ScoreEligible  bool          `json:"scoreEligible"`
	Source         *SignalSource `json:"source"`
}
type SignalSource struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}
type Input struct {
	RequestedParts    []Part          `json:"requestedParts"`
	CurrentDraft      Draft           `json:"currentDraft"`
	Persona           string          `json:"persona"`
	SellerCompany     SellerCompany   `json:"sellerCompany"`
	TargetAccount     TargetAccount   `json:"targetAccount"`
	LatestAnalysis    *Analysis       `json:"latestAnalysis,omitempty"`
	AnchorSignal      Signal          `json:"anchorSignal"`
	SupportingSignals []Signal        `json:"supportingSignals"`
	CommunicationDNA  json.RawMessage `json:"communicationDNA,omitempty"`
}
type ContextData struct {
	SellerCompany                           SellerCompany
	TargetAccount                           TargetAccount
	LatestAnalysis                          *Analysis
	AnchorSignal                            Signal
	SupportingSignals                       []Signal
	CommunicationDNA                        json.RawMessage
	AccountFound, AnchorFound, AnchorActive bool
}
type Output struct{ GeneratedParts Draft }

type Generator interface {
	Generate(context.Context, Input) (Output, error)
}

type ProviderUnavailableError struct{}

func (ProviderUnavailableError) Error() string { return "outreach generation provider is unavailable" }

type UnavailableGenerator struct{}

func (UnavailableGenerator) Generate(context.Context, Input) (Output, error) {
	return Output{}, ProviderUnavailableError{}
}
