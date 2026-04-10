package display

import "fmt"

const (
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorReset  = "\033[0m"
)

// Pair holds a recognised source text and its translation.
type Pair struct {
	Source string
	Target string
}

// Print writes a Pair to stdout with ANSI colours.
// [SRC] is yellow, [TGT] is green.
func Print(p Pair) {
	fmt.Printf("%s[SRC]%s %s\n", colorYellow, colorReset, p.Source)
	fmt.Printf("%s[TGT]%s %s\n\n", colorGreen, colorReset, p.Target)
}
