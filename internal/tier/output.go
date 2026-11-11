package tier

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSON renders the result as the stable machine-readable report.
func JSON(r Result) (string, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

// Text renders the result for a terminal or a job summary.
func Text(r Result) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Risk tier: %s\n", r.Tier)
	fmt.Fprintf(&b, "Human approvals required: %d\n", r.HumanApprovals)
	fmt.Fprintf(&b, "Always needs a person: %t\n", r.AlwaysHuman)
	if r.ApprovalMode == Advisory {
		b.WriteString("Approval check: advisory (warns, does not fail)\n")
	}
	if len(r.Reasons) == 0 {
		b.WriteString("Reasons: none (empty change)\n")
		return b.String()
	}
	b.WriteString("Reasons:\n")
	for _, re := range r.Reasons {
		fmt.Fprintf(&b, "  [%s] %s  (%s)\n", re.Tier, re.Path, re.Rule)
	}
	return b.String()
}
