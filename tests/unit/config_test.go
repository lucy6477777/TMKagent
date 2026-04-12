package unit_test

import (
	"os"
	"strings"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/config"
)

func TestLoad_MissingAPIKey(t *testing.T) {
	orig := os.Getenv("OPENAI_API_KEY")
	defer func() {
		if orig == "" {
			os.Unsetenv("OPENAI_API_KEY")
		} else {
			os.Setenv("OPENAI_API_KEY", orig)
		}
	}()
	os.Unsetenv("OPENAI_API_KEY")
	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when OPENAI_API_KEY is missing")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("error message should mention OPENAI_API_KEY, got: %v", err)
	}
}

func TestLoad_ReadsEnvVars(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-test")
	os.Setenv("OPENAI_BASE_URL", "https://custom.api/v1")
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("OPENAI_BASE_URL")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "sk-test" {
		t.Errorf("got APIKey %q, want %q", cfg.APIKey, "sk-test")
	}
	if cfg.BaseURL != "https://custom.api/v1" {
		t.Errorf("got BaseURL %q, want %q", cfg.BaseURL, "https://custom.api/v1")
	}
}

func TestLoad_DefaultBaseURL(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-test")
	os.Unsetenv("OPENAI_BASE_URL")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("got BaseURL %q, want default", cfg.BaseURL)
	}
}

func TestOverride(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-original")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg, _ := config.Load()
	cfg.Override("sk-new", "https://new.api/v1", "dg-test")
	if cfg.APIKey != "sk-new" {
		t.Errorf("Override did not update APIKey")
	}
	if cfg.BaseURL != "https://new.api/v1" {
		t.Errorf("Override did not update BaseURL")
	}
	if cfg.DeepgramAPIKey != "dg-test" {
		t.Errorf("Override did not update DeepgramAPIKey")
	}
}

func TestOverride_EmptyDoesNotReplace(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "sk-original")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg, _ := config.Load()
	cfg.Override("", "", "")
	if cfg.APIKey != "sk-original" {
		t.Errorf("Override with empty apiKey should not replace existing value")
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("Override with empty baseURL should not replace existing value")
	}
}
