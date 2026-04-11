package display

import (
	"fmt"
	"io"
	"os"
)

const (
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorDim    = "\033[2m"
	colorReset  = "\033[0m"
)

// Pair holds a recognised source text and its translation.
type Pair struct {
	Source string
	Target string
}

// Writer controls terminal output for the stream pipeline.
// It supports interim (overwriting) and final (fixed) display modes.
type Writer struct {
	out           io.Writer
	hasInterim    bool // true if the current line is an interim that should be overwritten
}

// NewWriter creates a display Writer that writes to stdout.
func NewWriter() *Writer {
	return &Writer{out: os.Stdout}
}

// NewWriterTo creates a display Writer that writes to the given io.Writer (for testing).
func NewWriterTo(w io.Writer) *Writer {
	return &Writer{out: w}
}

// PrintInterim shows a temporary transcription that will be overwritten by the next call.
// Uses \r to return to line start and \033[K to clear the line.
func (w *Writer) PrintInterim(text string) {
	fmt.Fprintf(w.out, "\r\033[K%s[...] %s%s", colorDim, text, colorReset)
	w.hasInterim = true
}

// ClearInterim removes the current interim line if present.
func (w *Writer) ClearInterim() {
	if w.hasInterim {
		fmt.Fprintf(w.out, "\r\033[K")
		w.hasInterim = false
	}
}

// PrintFinal writes a permanent source/target pair (not overwritable).
func (w *Writer) PrintFinal(p Pair) {
	w.ClearInterim()
	fmt.Fprintf(w.out, "%s[SRC]%s %s\n", colorYellow, colorReset, p.Source)
	fmt.Fprintf(w.out, "%s[TGT]%s %s\n\n", colorGreen, colorReset, p.Target)
}

// Print writes a Pair to stdout with ANSI colours (legacy API, used by old pipeline).
func Print(p Pair) {
	fmt.Printf("%s[SRC]%s %s\n", colorYellow, colorReset, p.Source)
	fmt.Printf("%s[TGT]%s %s\n\n", colorGreen, colorReset, p.Target)
}
