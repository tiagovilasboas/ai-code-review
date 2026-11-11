package tier

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarkIncomplete(t *testing.T) {
	cases := []struct {
		name, why, wantRule string
		start               Tier
	}{
		{"from low", "diff too large", "incomplete: diff too large", Low},
		{"from high stays high", "api error", "incomplete: api error", High},
		{"whitespace folded", "a\n b  c", "incomplete: a b c", Medium},
	}
	for _, c := range cases {
		res := MarkIncomplete(Result{Tier: c.start, Reasons: []Reason{}}, c.why)
		if res.Tier != High || res.HumanApprovals != 2 {
			t.Errorf("%s: %+v", c.name, res)
		}
		if len(res.Reasons) != 1 || res.Reasons[0].Rule != c.wantRule {
			t.Errorf("%s: reasons %+v", c.name, res.Reasons)
		}
	}
}

func TestMainIncomplete(t *testing.T) {
	cfg := writeCfg(t, cfgJSON)
	code, out, _ := run(t, "", "--config", cfg, "--format", "json", "--author", "owner", "--incomplete", "files list truncated")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var r Result
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatal(err)
	}
	if r.Tier != High || r.HumanApprovals != 2 || !strings.Contains(r.Reasons[0].Rule, "files list truncated") {
		t.Fatalf("result = %+v", r)
	}
	for name, args := range map[string][]string{
		"with file":  {"--config", cfg, "--incomplete", "x", "a.diff"},
		"with range": {"--config", cfg, "--incomplete", "x", "--from", "a", "--to", "b"},
	} {
		if code, _, _ := run(t, "", args...); code != 2 {
			t.Errorf("%s: code %d", name, code)
		}
	}
}
