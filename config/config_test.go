package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad_MissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when OPENAI_API_KEY is missing")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("error message should mention OPENAI_API_KEY, got: %v", err)
	}
}

func TestLoad_ReadsEnvVars(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test")
	t.Setenv("OPENAI_BASE_URL", "https://custom.api/v1")

	cfg, err := Load()
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
	t.Setenv("OPENAI_API_KEY", "sk-test")
	t.Setenv("OPENAI_BASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("got BaseURL %q, want default", cfg.BaseURL)
	}
}

func TestOverride(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-original")

	cfg, _ := Load()
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
	t.Setenv("OPENAI_API_KEY", "sk-original")

	cfg, _ := Load()
	cfg.Override("", "", "")
	if cfg.APIKey != "sk-original" {
		t.Errorf("Override with empty apiKey should not replace existing value")
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("Override with empty baseURL should not replace existing value")
	}
}

func TestLoad_ReadsDotEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	dotEnv := strings.Join([]string{
		`OPENAI_API_KEY="sk-dotenv"`,
		`OPENAI_BASE_URL=https://example.test/v1`,
		`DEEPGRAM_API_KEY=dg-dotenv`,
		`WEB_PUBLIC_BASE_URL=http://192.168.1.10:8080`,
		`LIVEKIT_URL=wss://demo.livekit.cloud`,
		`LIVEKIT_API_KEY=lk-key`,
		`LIVEKIT_API_SECRET=lk-secret`,
	}, "\n")
	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(dotEnv), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("DEEPGRAM_API_KEY", "")
	t.Setenv("WEB_PUBLIC_BASE_URL", "")
	t.Setenv("LIVEKIT_URL", "")
	t.Setenv("LIVEKIT_API_KEY", "")
	t.Setenv("LIVEKIT_API_SECRET", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIKey != "sk-dotenv" {
		t.Fatalf("got APIKey %q, want %q", cfg.APIKey, "sk-dotenv")
	}
	if cfg.BaseURL != "https://example.test/v1" {
		t.Fatalf("got BaseURL %q, want %q", cfg.BaseURL, "https://example.test/v1")
	}
	if cfg.DeepgramAPIKey != "dg-dotenv" {
		t.Fatalf("got DeepgramAPIKey %q, want %q", cfg.DeepgramAPIKey, "dg-dotenv")
	}
	if cfg.PublicBaseURL != "http://192.168.1.10:8080" {
		t.Fatalf("got PublicBaseURL %q, want %q", cfg.PublicBaseURL, "http://192.168.1.10:8080")
	}
	if cfg.LiveKitURL != "wss://demo.livekit.cloud" {
		t.Fatalf("got LiveKitURL %q, want %q", cfg.LiveKitURL, "wss://demo.livekit.cloud")
	}
	if cfg.LiveKitAPIKey != "lk-key" {
		t.Fatalf("got LiveKitAPIKey %q, want %q", cfg.LiveKitAPIKey, "lk-key")
	}
	if cfg.LiveKitAPISecret != "lk-secret" {
		t.Fatalf("got LiveKitAPISecret %q, want %q", cfg.LiveKitAPISecret, "lk-secret")
	}
}

func TestLoad_EnvironmentOverridesDotEnv(t *testing.T) {
	tmpDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("OPENAI_API_KEY=sk-dotenv\nDEEPGRAM_API_KEY=dg-dotenv\n"), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	t.Setenv("OPENAI_API_KEY", "sk-env")
	t.Setenv("DEEPGRAM_API_KEY", "dg-env")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.APIKey != "sk-env" {
		t.Fatalf("got APIKey %q, want %q", cfg.APIKey, "sk-env")
	}
	if cfg.DeepgramAPIKey != "dg-env" {
		t.Fatalf("got DeepgramAPIKey %q, want %q", cfg.DeepgramAPIKey, "dg-env")
	}
}
