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

type Draft struct{ Subject, Opening, Value, CTA string }
type SellerCompany struct {
	Name, Tagline, Website, Description        string
	Products, ValuePropositions, BuyerPersonas json.RawMessage
	CommunicationDNA                           json.RawMessage
}
type TargetAccount struct {
	ID                                   string
	Name, Industry, Website, Description string
}
type Analysis struct{ Tier, NextBestAction *string }
type Signal struct {
	ID, Type, Title, Body, Strength, Relevance, SignalDate, EvidenceStatus string
	Verified, ScoreEligible                                                bool
	Source                                                                 *SignalSource
}
type SignalSource struct{ Name, Type, URL string }
type Input struct {
	RequestedParts    []Part
	CurrentDraft      Draft
	Persona           string
	SellerCompany     SellerCompany
	TargetAccount     TargetAccount
	LatestAnalysis    *Analysis
	AnchorSignal      Signal
	SupportingSignals []Signal
	CommunicationDNA  any
}
type ContextData struct {
	SellerCompany                           SellerCompany
	TargetAccount                           TargetAccount
	LatestAnalysis                          *Analysis
	AnchorSignal                            Signal
	SupportingSignals                       []Signal
	CommunicationDNA                        any
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
