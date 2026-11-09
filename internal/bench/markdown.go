package bench

import (
	"fmt"
	"strings"
)

func pct(v float64, ok bool) string {
	if !ok {
		return "n/a"
	}
	return fmt.Sprintf("%.2f", v)
}

func row(name string, c Counts) string {
	p, pok := c.Precision()
	r, rok := c.Recall()
	f, fok := c.F1()
	return fmt.Sprintf("| %s | %d | %d | %d | %s | %s | %s |\n", name, c.TP, c.FP, c.FN, pct(p, pok), pct(r, rok), pct(f, fok))
}

// Markdown renders the per-class table, the overall row and the misses.
func Markdown(r Report) string {
	var b strings.Builder
	b.WriteString("| Class | TP | FP | FN | Precision | Recall | F1 |\n|---|---|---|---|---|---|---|\n")
	for _, c := range r.Classes {
		b.WriteString(row(c.Class, c.Counts))
	}
	b.WriteString(row("**overall**", r.Overall))
	fmt.Fprintf(&b, "\nCases: %d. Clean cases (must stay silent): %d, silent: %d.\n", r.Cases, r.CleanCases, r.CleanSilent)
	b.WriteString(misses(r))
	return b.String()
}

func misses(r Report) string {
	var b strings.Builder
	for _, c := range r.CaseResults {
		for _, k := range c.Missed {
			fmt.Fprintf(&b, "- FN %s: %s:%d (%s)\n", c.ID, k.Path, k.Line, k.Class)
		}
		for _, k := range c.Extra {
			fmt.Fprintf(&b, "- FP %s: %s:%d (%s)\n", c.ID, k.Path, k.Line, k.Class)
		}
	}
	if b.Len() == 0 {
		return ""
	}
	return "\nMisses:\n" + b.String()
}
