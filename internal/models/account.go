package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID                uuid.UUID        `json:"id"`
	Name              string           `json:"name"`
	Industry          *string          `json:"industry"`
	HQ                *string          `json:"hq"`
	Lifecycle         string           `json:"lifecycle"`
	Analysis          *AccountAnalysis `json:"analysis"`
	ActiveSignalCount int              `json:"activeSignalCount"`
}

type AccountAnalysis struct {
	ICPScore       float64 `json:"icpScore"`
	ICPFit         string  `json:"icpFit"`
	SignalScore    float64 `json:"signalScore"`
	ResonanceScore float64 `json:"resonanceScore"`
	Tier           string  `json:"tier"`
	NextBestAction *string `json:"nextBestAction"`
}

type AccountListResponse struct {
	Items []Account `json:"items"`
}

type AccountOverview struct {
	ID          uuid.UUID                `json:"id"`
	Name        string                   `json:"name"`
	Domain      string                   `json:"domain"`
	WebURL      string                   `json:"webUrl"`
	Industry    *string                  `json:"industry"`
	HQ          *string                  `json:"hq"`
	Employees   *int32                   `json:"employees"`
	Revenue     *Revenue                 `json:"revenue"`
	Founded     *int16                   `json:"founded"`
	Description *string                  `json:"description"`
	Lifecycle   string                   `json:"lifecycle"`
	Analysis    *AccountOverviewAnalysis `json:"analysis"`
}

type Revenue struct {
	AmountM  *float64 `json:"amountM"`
	Currency *string  `json:"currency"`
}

type AccountOverviewAnalysis struct {
	ICPScore       float64         `json:"icpScore"`
	ICPFit         string          `json:"icpFit"`
	SignalScore    float64         `json:"signalScore"`
	ResonanceScore float64         `json:"resonanceScore"`
	Tier           string          `json:"tier"`
	WhyThisAccount *string         `json:"whyThisAccount"`
	WhyNow         *string         `json:"whyNow"`
	NextBestAction *NextBestAction `json:"nextBestAction"`
}

type NextBestAction struct {
	Action     string  `json:"action"`
	Rationale  *string `json:"rationale"`
	TimeWindow *string `json:"timeWindow"`
	Priority   *string `json:"priority"`
}

type AccountSignalsResponse struct {
	Summary SignalSummary `json:"summary"`
	Items   []Signal      `json:"items"`
}

type CommunicationDNAResponse struct {
	Data *CommunicationDNA `json:"data"`
}
type CommunicationDNA struct {
	ID                  uuid.UUID       `json:"id"`
	AccountID           uuid.UUID       `json:"accountId"`
	Tone                DNAStyle        `json:"tone"`
	Vocabulary          DNAVocabulary   `json:"vocabulary"`
	ValuePropositions   []DNAEvidence   `json:"valuePropositions"`
	ProblemFraming      DNAProblem      `json:"problemFraming"`
	ProofStyle          DNAStyle        `json:"proofStyle"`
	CTAPatterns         DNACallToAction `json:"ctaPatterns"`
	RecurringPhrases    []DNAPhrase     `json:"recurringPhrases"`
	DoRules             []string        `json:"doRules"`
	DontRules           []string        `json:"dontRules"`
	BuyingSignalSources []SignalSource  `json:"buyingSignalSources"`
	CreatedAt           string          `json:"createdAt"`
}
type DNAStyle struct {
	Primary     *string     `json:"primary"`
	Secondary   *string     `json:"secondary"`
	Description *string     `json:"description"`
	Status      string      `json:"status"`
	Sources     []DNASource `json:"sources"`
}

func (s *DNAStyle) UnmarshalJSON(data []byte) error {
	type styleAlias DNAStyle
	var value struct {
		styleAlias
		Source     *string `json:"source"`
		SourceType *string `json:"sourceType"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*s = DNAStyle(value.styleAlias)
	if s.Sources == nil {
		s.Sources = sourcesFromLegacy(value.Source, value.SourceType)
	}
	return nil
}

type DNASource struct {
	Name *string `json:"name"`
	Type *string `json:"type"`
	URL  *string `json:"url"`
}

func (s *DNASource) UnmarshalJSON(data []byte) error {
	raw := bytes.TrimSpace(data)
	if len(raw) > 0 && raw[0] == '"' {
		var sourceURL string
		if err := json.Unmarshal(raw, &sourceURL); err != nil {
			return err
		}
		if strings.TrimSpace(sourceURL) == "" {
			return fmt.Errorf("DNA source string must not be empty")
		}
		s.Name, s.Type, s.URL = nil, nil, &sourceURL
		return nil
	}

	type sourceAlias DNASource
	var source sourceAlias
	if err := json.Unmarshal(raw, &source); err != nil {
		return err
	}
	*s = DNASource(source)
	return nil
}

type DNAVocabulary struct {
	Status string    `json:"status"`
	Terms  []DNATerm `json:"terms"`
}
type DNATerm struct {
	Term      string      `json:"term"`
	Context   *string     `json:"context"`
	Frequency *string     `json:"frequency"`
	Sources   []DNASource `json:"sources"`
}

func (t *DNATerm) UnmarshalJSON(data []byte) error {
	type termAlias DNATerm
	var value struct {
		termAlias
		Word       string  `json:"word"`
		Source     *string `json:"source"`
		SourceType *string `json:"sourceType"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*t = DNATerm(value.termAlias)
	if t.Term == "" {
		t.Term = value.Word
	}
	if t.Sources == nil {
		t.Sources = sourcesFromLegacy(value.Source, value.SourceType)
	}
	return nil
}

type DNAEvidence struct {
	Quote   string      `json:"quote"`
	Status  string      `json:"status"`
	Sources []DNASource `json:"sources"`
}

func (e *DNAEvidence) UnmarshalJSON(data []byte) error {
	type evidenceAlias DNAEvidence
	var value struct {
		evidenceAlias
		Prop       string  `json:"prop"`
		Source     *string `json:"source"`
		SourceType *string `json:"sourceType"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*e = DNAEvidence(value.evidenceAlias)
	if e.Quote == "" {
		e.Quote = value.Prop
	}
	if e.Sources == nil {
		e.Sources = sourcesFromLegacy(value.Source, value.SourceType)
	}
	return nil
}

type DNAProblem struct {
	Description *string     `json:"description"`
	Quote       *string     `json:"quote"`
	Status      string      `json:"status"`
	Sources     []DNASource `json:"sources"`
}

func (p *DNAProblem) UnmarshalJSON(data []byte) error {
	type problemAlias DNAProblem
	var value struct {
		problemAlias
		Source     *string `json:"source"`
		SourceType *string `json:"sourceType"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = DNAProblem(value.problemAlias)
	if p.Sources == nil {
		p.Sources = sourcesFromLegacy(value.Source, value.SourceType)
	}
	return nil
}

type DNACallToAction struct {
	Style       *string     `json:"style"`
	Description *string     `json:"description"`
	Examples    []string    `json:"examples"`
	Status      string      `json:"status"`
	Sources     []DNASource `json:"sources"`
}

func (c *DNACallToAction) UnmarshalJSON(data []byte) error {
	type callToActionAlias DNACallToAction
	var value struct {
		callToActionAlias
		Source     *string `json:"source"`
		SourceType *string `json:"sourceType"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*c = DNACallToAction(value.callToActionAlias)
	if c.Sources == nil {
		c.Sources = sourcesFromLegacy(value.Source, value.SourceType)
	}
	return nil
}

type DNAPhrase struct {
	Quote       string      `json:"quote"`
	Description *string     `json:"description"`
	Status      string      `json:"status"`
	Sources     []DNASource `json:"sources"`
}

func (p *DNAPhrase) UnmarshalJSON(data []byte) error {
	raw := bytes.TrimSpace(data)
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &p.Quote); err != nil {
			return err
		}
		p.Status = "INSUFFICIENT_DATA"
		p.Sources = make([]DNASource, 0)
		return nil
	}

	type phraseAlias DNAPhrase
	var value struct {
		phraseAlias
		Source     *string `json:"source"`
		SourceType *string `json:"sourceType"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	*p = DNAPhrase(value.phraseAlias)
	if p.Sources == nil {
		p.Sources = sourcesFromLegacy(value.Source, value.SourceType)
	}
	return nil
}

func sourcesFromLegacy(source, sourceType *string) []DNASource {
	if source == nil {
		return nil
	}
	values := strings.Split(*source, ",")
	sources := make([]DNASource, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		url := value
		sources = append(sources, DNASource{Type: sourceType, URL: &url})
	}
	return sources
}

type SignalSummary struct {
	Total  int            `json:"total"`
	Active int            `json:"active"`
	ByType map[string]int `json:"byType"`
}

type Signal struct {
	ID             uuid.UUID     `json:"id"`
	Type           string        `json:"type"`
	Title          string        `json:"title"`
	Body           *string       `json:"body"`
	Strength       string        `json:"strength"`
	Relevance      *string       `json:"relevance"`
	SignalDate     *string       `json:"signalDate"`
	SignalDateRaw  *string       `json:"signalDateRaw"`
	FreshnessLabel *string       `json:"freshnessLabel"`
	EvidenceStatus string        `json:"evidenceStatus"`
	Verified       bool          `json:"verified"`
	IsActive       bool          `json:"isActive"`
	ScoreEligible  bool          `json:"scoreEligible"`
	Source         *SignalSource `json:"source"`
}

type SignalPulseResponse struct {
	Metrics  SignalPulseMetrics   `json:"metrics"`
	Accounts []SignalPulseAccount `json:"accounts"`
}
type SignalPulseMetrics struct {
	ActiveSignals int `json:"activeSignals"`
	NewThisWeek   int `json:"newThisWeek"`
	HotAccounts   int `json:"hotAccounts"`
	GoingCold     int `json:"goingCold"`
}
type SignalPulseAccount struct {
	AccountID              uuid.UUID            `json:"accountId"`
	Name                   string               `json:"name"`
	Industry               *string              `json:"industry"`
	Tier                   *string              `json:"tier"`
	Urgency                string               `json:"urgency"`
	ActiveSignalCount      int                  `json:"activeSignalCount"`
	HighActiveSignalCount  int                  `json:"highActiveSignalCount"`
	LatestActiveSignalDate *string              `json:"latestActiveSignalDate"`
	NextBestAction         *string              `json:"nextBestAction"`
	Signals                []SignalPulsePreview `json:"signals"`
}
type SignalPulsePreview struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Strength   string    `json:"strength"`
	SignalDate *string   `json:"signalDate"`
}

type SignalPulseRow struct {
	AccountID                               uuid.UUID
	Name                                    string
	Industry                                *string
	Tier                                    *string
	NextBestAction                          *string
	SignalID                                *uuid.UUID
	SignalType, SignalTitle, SignalStrength *string
	SignalDate                              *time.Time
	SignalCreatedAt                         *time.Time
	IsActive                                *bool
}

type SignalSource struct {
	Name *string `json:"name"`
	Type *string `json:"type"`
	URL  string  `json:"url"`
}

type DNAPortfolioRow struct {
	AccountID         uuid.UUID
	Name              string
	Industry          *string
	Tier              *string
	DNA               *CommunicationDNA
	ActiveSignalCount int
	SignalTypes       []string
}

type DNAPortfolioResponse struct {
	Summary DNAPortfolioSummary `json:"summary"`
	Items   []DNAPortfolioItem  `json:"items"`
}
type DNAPortfolioSummary struct {
	TotalProfiles int            `json:"totalProfiles"`
	ByTier        map[string]int `json:"byTier"`
	ByIndustry    map[string]int `json:"byIndustry"`
}
type DNAPortfolioItem struct {
	AccountID         uuid.UUID `json:"accountId"`
	Name              string    `json:"name"`
	Industry          *string   `json:"industry"`
	Tier              *string   `json:"tier"`
	ActiveSignalCount int       `json:"activeSignalCount"`
	Tone              *string   `json:"tone"`
	Vocabulary        []string  `json:"vocabulary"`
	ProblemFraming    *string   `json:"problemFraming"`
	ProofStyle        *string   `json:"proofStyle"`
	CTAStyle          *string   `json:"ctaStyle"`
	DoRules           []string  `json:"doRules"`
	DontRules         []string  `json:"dontRules"`
	SignalTypes       []string  `json:"signalTypes"`
}
type DNACompareRequest struct {
	AccountIDs []uuid.UUID `json:"accountIds"`
}
type DNACompareResponse struct {
	SelectedCount    int                     `json:"selectedCount"`
	DominantTone     []DNAValueCount         `json:"dominantTone"`
	SharedVocabulary []DNAValueCount         `json:"sharedVocabulary"`
	UniqueVocabulary []DNAValueCount         `json:"uniqueVocabulary"`
	ProofStyles      []DNAValueCount         `json:"proofStyles"`
	CTAStyles        []DNAValueCount         `json:"ctaStyles"`
	DoRules          []DNAValueCount         `json:"doRules"`
	DontRules        []DNAValueCount         `json:"dontRules"`
	SignalTypes      []DNAValueCount         `json:"signalTypes"`
	ProblemFraming   []DNAProblemFramingItem `json:"problemFraming"`
}
type DNAValueCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}
type DNAProblemFramingItem struct {
	AccountID   uuid.UUID `json:"accountId"`
	AccountName string    `json:"accountName"`
	Value       *string   `json:"value"`
}
