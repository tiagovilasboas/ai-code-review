package bench

// Counts are confusion counts; the metrics are derived, never stored.
type Counts struct {
	TP int `json:"tp"`
	FP int `json:"fp"`
	FN int `json:"fn"`
}

// Precision is TP/(TP+FP); ok is false when the engine reported nothing.
func (c Counts) Precision() (float64, bool) { return ratio(c.TP, c.TP+c.FP) }

// Recall is TP/(TP+FN); ok is false when nothing was expected.
func (c Counts) Recall() (float64, bool) { return ratio(c.TP, c.TP+c.FN) }

// F1 is the harmonic mean of precision and recall; ok is false if either is undefined.
func (c Counts) F1() (float64, bool) {
	p, pok := c.Precision()
	r, rok := c.Recall()
	if !pok || !rok {
		return 0, false
	}
	if p+r == 0 {
		return 0, true
	}
	return 2 * p * r / (p + r), true
}

func ratio(n, d int) (float64, bool) {
	if d == 0 {
		return 0, false
	}
	return float64(n) / float64(d), true
}

// Match compares expected and got as sets of path+line+class keys; duplicates
// on either side count once. hit are true positives, missed are false
// negatives, extra are false positives.
func Match(expected, got []Key) (hit, missed, extra []Key) {
	want := make(map[Key]bool, len(expected))
	for _, k := range expected {
		want[k] = true
	}
	seen := make(map[Key]bool, len(got))
	for _, k := range got {
		if seen[k] {
			continue
		}
		seen[k] = true
		if want[k] {
			hit = append(hit, k)
		} else {
			extra = append(extra, k)
		}
	}
	done := make(map[Key]bool, len(expected))
	for _, k := range expected {
		if done[k] {
			continue
		}
		done[k] = true
		if !seen[k] {
			missed = append(missed, k)
		}
	}
	return hit, missed, extra
}
