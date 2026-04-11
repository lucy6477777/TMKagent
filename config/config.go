package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds runtime configuration loaded from environment.
type Config struct {
	APIKey         string
	BaseURL        string
	DeepgramAPIKey string

	LiveKitURL       string
	LiveKitAPIKey    string
	LiveKitAPISecret string
}

// Load reads API keys from environment.
// It first loads .env from the current directory (if present), then reads os env.
// OPENAI_API_KEY is required. Other keys are optional.
func Load() (*Config, error) {
	loadDotEnv(".env")

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
	return &Config{
		APIKey:           apiKey,
		BaseURL:          baseURL,
		DeepgramAPIKey:   os.Getenv("DEEPGRAM_API_KEY"),
		LiveKitURL:       os.Getenv("LIVEKIT_URL"),
		LiveKitAPIKey:    os.Getenv("LIVEKIT_API_KEY"),
		LiveKitAPISecret: os.Getenv("LIVEKIT_API_SECRET"),
	}, nil
}

// Override replaces config fields with non-empty CLI flag values.
func (c *Config) Override(apiKey, baseURL, deepgramAPIKey string) {
	if apiKey != "" {
		c.APIKey = apiKey
	}
	if baseURL != "" {
		c.BaseURL = baseURL
	}
	if deepgramAPIKey != "" {
		c.DeepgramAPIKey = deepgramAPIKey
	}
}

// loadDotEnv reads a .env file and sets any variables not already in the environment.
// Silently ignored if the file does not exist. No external dependency needed.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"'`)
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}
