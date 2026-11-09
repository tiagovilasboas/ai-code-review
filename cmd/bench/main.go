// Command bench measures the review engine against the labelled corpus.
// It is a measurement: it exits 0 whatever the scores, and non-zero only
// when the corpus or the output file cannot be read or written.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/tiagovilasboas/ai-code-review/internal/bench"
)

func main() {
	corpus := flag.String("corpus", "bench/corpus", "corpus directory")
	jsonOut := flag.String("json", "bench/results.json", "results file to write; empty to skip")
	flag.Parse()
	if err := run(*corpus, *jsonOut); err != nil {
		fmt.Fprintln(os.Stderr, "bench:", err)
		os.Exit(1)
	}
}

func run(corpus, jsonOut string) error {
	cases, err := bench.LoadCorpus(corpus)
	if err != nil {
		return err
	}
	report := bench.Evaluate(cases, bench.DefaultAnalyzer)
	fmt.Print(bench.Markdown(report))
	if jsonOut == "" {
		return nil
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(jsonOut, append(data, '\n'), 0o644)
}
