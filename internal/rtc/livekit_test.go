package rtc

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRelayMsg_InterimJSON(t *testing.T) {
	msg := RelayMsg{Type: "interim", Text: "你好"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded RelayMsg
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
	msg := RelayMsg{Type: "pair", Source: "你好世界", Target: "Hello World"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded RelayMsg
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
	msg := RelayMsg{Type: "interim", Text: "hello"}
	data, _ := json.Marshal(msg)
	s := string(data)
	if strings.Contains(s, "source") || strings.Contains(s, "target") {
		t.Errorf("interim JSON should omit empty source/target; got: %s", s)
	}
}

func TestRelayMsg_PairOmitsText(t *testing.T) {
	msg := RelayMsg{Type: "pair", Source: "src", Target: "tgt"}
	data, _ := json.Marshal(msg)
	s := string(data)
	if strings.Contains(s, `"text"`) {
		t.Errorf("pair JSON should omit empty text; got: %s", s)
	}
}
