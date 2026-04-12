package display

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
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
		Print(Pair{Source: "你好世界", Target: "Hello World"})
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
		Print(Pair{Source: "x", Target: "y"})
	})
	if !strings.Contains(out, "[SRC]") {
		t.Errorf("output missing [SRC] label; got: %q", out)
	}
	if !strings.Contains(out, "[TGT]") {
		t.Errorf("output missing [TGT] label; got: %q", out)
	}
}

func TestWriter_PrintInterim(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriterTo(&buf)
	w.PrintInterim("hello")

	out := buf.String()
	if !strings.Contains(out, "hello") {
		t.Errorf("interim output missing text; got: %q", out)
	}
	if !strings.Contains(out, "[...]") {
		t.Errorf("interim output missing [...] prefix; got: %q", out)
	}
	if !strings.HasPrefix(out, "\r") {
		t.Errorf("interim should start with \\r for line overwrite; got: %q", out)
	}
}

func TestWriter_PrintFinal(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriterTo(&buf)
	w.PrintFinal(Pair{Source: "你好", Target: "Hello"})

	out := buf.String()
	if !strings.Contains(out, "[SRC]") {
		t.Errorf("final output missing [SRC]; got: %q", out)
	}
	if !strings.Contains(out, "[TGT]") {
		t.Errorf("final output missing [TGT]; got: %q", out)
	}
	if !strings.Contains(out, "你好") {
		t.Errorf("final output missing source text; got: %q", out)
	}
	if !strings.Contains(out, "Hello") {
		t.Errorf("final output missing target text; got: %q", out)
	}
}

func TestWriter_ClearInterim(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriterTo(&buf)
	w.PrintInterim("temp")
	buf.Reset()
	w.ClearInterim()

	out := buf.String()
	if !strings.HasPrefix(out, "\r") {
		t.Errorf("clear should start with \\r; got: %q", out)
	}
}

func TestWriter_ClearInterim_NoopWhenNoInterim(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriterTo(&buf)
	w.ClearInterim()

	if buf.Len() != 0 {
		t.Errorf("ClearInterim with no interim should write nothing; got: %q", buf.String())
	}
}
