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
