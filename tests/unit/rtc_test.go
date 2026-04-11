package unit_test

import (
	"encoding/json"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/rtc"
)

func TestRelayMsg_InterimJSON(t *testing.T) {
	msg := rtc.RelayMsg{Type: "interim", Text: "你好"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded rtc.RelayMsg
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Type != "interim" {
		t.Errorf("got type %q, want %q", decoded.Type, "interim")
	}
	if decoded.Text != "你好" {
		t.Errorf("got text %q, want %q", decoded.Text, "你好")
	}
	if decoded.Source != "" || decoded.Target != "" {
		t.Error("interim should have empty source and target")
	}
}

func TestRelayMsg_PairJSON(t *testing.T) {
	msg := rtc.RelayMsg{Type: "pair", Source: "你好世界", Target: "Hello World"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded rtc.RelayMsg
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if decoded.Type != "pair" {
		t.Errorf("got type %q, want %q", decoded.Type, "pair")
	}
	if decoded.Source != "你好世界" {
		t.Errorf("got source %q, want %q", decoded.Source, "你好世界")
	}
	if decoded.Target != "Hello World" {
		t.Errorf("got target %q, want %q", decoded.Target, "Hello World")
	}
}

func TestRelayMsg_OmitEmpty(t *testing.T) {
	msg := rtc.RelayMsg{Type: "interim", Text: "hello"}
	data, _ := json.Marshal(msg)
	s := string(data)
	if contains(s, "source") || contains(s, "target") {
		t.Errorf("interim JSON should omit empty source/target; got: %s", s)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
