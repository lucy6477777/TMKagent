package unit_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/display"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrint_ContainsSourceAndTarget(t *testing.T) {
	out := captureStdout(func() {
		display.Print(display.Pair{Source: "你好世界", Target: "Hello World"})
	})
	if !strings.Contains(out, "你好世界") {
		t.Errorf("output missing source text; got: %q", out)
	}
	if !strings.Contains(out, "Hello World") {
		t.Errorf("output missing target text; got: %q", out)
	}
}

func TestPrint_ContainsSRCAndTGTLabels(t *testing.T) {
	out := captureStdout(func() {
		display.Print(display.Pair{Source: "x", Target: "y"})
	})
	if !strings.Contains(out, "[SRC]") {
		t.Errorf("output missing [SRC] label; got: %q", out)
	}
	if !strings.Contains(out, "[TGT]") {
		t.Errorf("output missing [TGT] label; got: %q", out)
	}
}
