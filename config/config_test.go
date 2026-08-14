package config

import "testing"

func TestLoadGeminiConfiguration(t *testing.T) {
	t.Setenv("DEMO_COMPANY_PROFILE_ID", "00000000-0000-0000-0000-000000000001")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GEMINI_MODEL", "")
	cfg := Load()
	if cfg.GeminiAPIKey != "" || cfg.GeminiModel != "gemini-3.6-flash" {
		t.Fatalf("defaults=%+v", cfg)
	}
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GEMINI_MODEL", "custom-model")
	cfg = Load()
	if cfg.GeminiAPIKey != "test-key" || cfg.GeminiModel != "custom-model" {
		t.Fatalf("configured=%+v", cfg)
	}
}
