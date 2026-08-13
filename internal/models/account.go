package models

import "github.com/google/uuid"

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
