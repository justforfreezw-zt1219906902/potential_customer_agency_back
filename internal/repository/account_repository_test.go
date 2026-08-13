package repository

import "testing"

func TestParseAnalysisMapsActionOnly(t *testing.T) {
	analysis, err := parseAnalysis("86.50", "High", "91", "89.20", "Focus Accounts", []byte(`{"action":"Prepare signal-led outreach","rationale":"internal"}`))
	if err != nil {
		t.Fatalf("parseAnalysis returned error: %v", err)
	}
	if analysis.ICPScore != 86.5 || analysis.SignalScore != 91 || analysis.ResonanceScore != 89.2 {
		t.Fatalf("unexpected scores: %+v", analysis)
	}
	if analysis.NextBestAction == nil || *analysis.NextBestAction != "Prepare signal-led outreach" {
		t.Fatalf("unexpected action: %v", analysis.NextBestAction)
	}
}

func TestParseAnalysisAllowsNullAction(t *testing.T) {
	analysis, err := parseAnalysis("1", "Low", "2", "3", "Below ICP", nil)
	if err != nil {
		t.Fatalf("parseAnalysis returned error: %v", err)
	}
	if analysis.NextBestAction != nil {
		t.Fatalf("expected nil next best action, got %q", *analysis.NextBestAction)
	}
}

func TestParseAnalysisRejectsMalformedAction(t *testing.T) {
	if _, err := parseAnalysis("1", "Low", "2", "3", "Below ICP", []byte(`{"rationale":"missing action"}`)); err == nil {
		t.Fatal("expected malformed action error")
	}
}
