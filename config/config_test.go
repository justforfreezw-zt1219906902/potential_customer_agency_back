package config

import (
	"testing"
	"time"
)

func TestLoadGeminiConfiguration(t *testing.T) {
	t.Setenv("DEMO_COMPANY_PROFILE_ID", "00000000-0000-0000-0000-000000000001")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GEMINI_MODEL", "")
	t.Setenv("OUTREACH_REQUEST_TIMEOUT_SECONDS", "")
	cfg := Load()
	if cfg.GeminiAPIKey != "" || cfg.GeminiModel != "gemini-3.6-flash" || cfg.OutreachRequestTimeout != 30*time.Second {
		t.Fatalf("defaults=%+v", cfg)
	}
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GEMINI_MODEL", "custom-model")
	t.Setenv("OUTREACH_REQUEST_TIMEOUT_SECONDS", "45")
	cfg = Load()
	if cfg.GeminiAPIKey != "test-key" || cfg.GeminiModel != "custom-model" || cfg.OutreachRequestTimeout != 45*time.Second {
		t.Fatalf("configured=%+v", cfg)
	}
}

func TestLoadRejectsInvalidOutreachRequestTimeout(t *testing.T) {
	for _, value := range []string{"0", "-1", "abc"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("DEMO_COMPANY_PROFILE_ID", "00000000-0000-0000-0000-000000000001")
			t.Setenv("OUTREACH_REQUEST_TIMEOUT_SECONDS", value)
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic for %q", value)
				}
			}()
			_ = Load()
		})
	}
}
