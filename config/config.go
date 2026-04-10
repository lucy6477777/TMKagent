package config

import (
	"fmt"
	"os"
)

// Config holds runtime configuration loaded from environment.
type Config struct {
	APIKey  string
	BaseURL string
}

// Load reads OPENAI_API_KEY and OPENAI_BASE_URL from environment.
// Returns an error if OPENAI_API_KEY is not set.
func Load() (*Config, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf(
			"OPENAI_API_KEY is not set\n" +
				"  Run: export OPENAI_API_KEY=sk-...\n" +
				"  Or:  mini-tmk-agent --api-key sk-...",
		)
	}
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &Config{APIKey: apiKey, BaseURL: baseURL}, nil
}

// Override replaces config fields with non-empty CLI flag values.
func (c *Config) Override(apiKey, baseURL string) {
	if apiKey != "" {
		c.APIKey = apiKey
	}
	if baseURL != "" {
		c.BaseURL = baseURL
	}
}
