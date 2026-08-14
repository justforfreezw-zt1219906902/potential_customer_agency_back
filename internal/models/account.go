package models

import "github.com/google/uuid"
import "time"

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
type DNASource struct {
	Name *string `json:"name"`
	Type *string `json:"type"`
	URL  *string `json:"url"`
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
type DNAEvidence struct {
	Quote   string      `json:"quote"`
	Status  string      `json:"status"`
	Sources []DNASource `json:"sources"`
}
type DNAProblem struct {
	Description *string     `json:"description"`
	Quote       *string     `json:"quote"`
	Status      string      `json:"status"`
	Sources     []DNASource `json:"sources"`
}
type DNACallToAction struct {
	Style       *string     `json:"style"`
	Description *string     `json:"description"`
	Examples    []string    `json:"examples"`
	Status      string      `json:"status"`
	Sources     []DNASource `json:"sources"`
}
type DNAPhrase struct {
	Quote       string      `json:"quote"`
	Description *string     `json:"description"`
	Status      string      `json:"status"`
	Sources     []DNASource `json:"sources"`
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
