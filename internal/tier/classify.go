package tier

import (
	"sort"
	"strings"
)

// Reason explains why one path (or the author) set the tier.
type Reason struct {
	Path string `json:"path"`
	Rule string `json:"rule"`
	Tier Tier   `json:"tier"`
}

// Result is the classification of one change.
type Result struct {
	SchemaVersion  int          `json:"schemaVersion"`
	Tier           Tier         `json:"tier"`
	HumanApprovals int          `json:"humanApprovals"`
	AlwaysHuman    bool         `json:"alwaysHuman"`
	ApprovalMode   ApprovalMode `json:"approvalMode"`
	Reasons        []Reason     `json:"reasons"`
}

// Classify assigns a tier to the changed paths. author is the PR author login;
// empty skips the trusted-author rule. The highest tier of any path wins, and
// a path with no matching rule gets cfg.Default. An empty change is low.
func Classify(cfg Config, paths []string, author string) Result {
	res := Result{SchemaVersion: SchemaVersion, Tier: Low, ApprovalMode: cfg.approvalMode(), Reasons: []Reason{}}
	for _, p := range paths {
		pt, rule := pathTier(cfg, p)
		res.Tier = maxTier(res.Tier, pt)
		res.Reasons = append(res.Reasons, Reason{Path: p, Rule: rule, Tier: pt})
		if g, ok := alwaysHumanRule(cfg, p); ok {
			res.AlwaysHuman = true
			res.Reasons = append(res.Reasons, Reason{Path: p, Rule: "alwaysHuman: " + g, Tier: pt})
		}
	}
	if untrusted(cfg, author) {
		res.Tier = maxTier(res.Tier, Medium)
		res.Reasons = append(res.Reasons, Reason{
			Path: "(author) " + author, Rule: "untrusted-author: not in trustedAuthors", Tier: Medium,
		})
	}
	res.HumanApprovals = res.Tier.Approvals()
	if res.AlwaysHuman && res.HumanApprovals < 1 {
		res.HumanApprovals = 1
	}
	sortReasons(res.Reasons)
	return res
}

// pathTier returns the highest tier among the globs matching p, or the default.
func pathTier(cfg Config, p string) (Tier, string) {
	for _, l := range []struct {
		t     Tier
		globs []string
	}{{High, cfg.High}, {Medium, cfg.Medium}, {Low, cfg.Low}} {
		if g, ok := firstMatch(l.globs, p); ok {
			return l.t, string(l.t) + ": " + g
		}
	}
	return cfg.Default, "default"
}

func alwaysHumanRule(cfg Config, p string) (string, bool) {
	if g, ok := firstMatch(cfg.AlwaysHuman, p); ok {
		return g, true
	}
	if matchGlob(workflowGlob, p) {
		return workflowGlob + " (built in)", true
	}
	return "", false
}

// untrusted reports whether the author fails the trusted list. With no list or
// no author the rule does not apply.
func untrusted(cfg Config, author string) bool {
	if author == "" || len(cfg.TrustedAuthors) == 0 {
		return false
	}
	for _, t := range cfg.TrustedAuthors {
		if strings.EqualFold(t, author) {
			return false
		}
	}
	return true
}

func sortReasons(rs []Reason) {
	sort.SliceStable(rs, func(i, j int) bool {
		if rs[i].Tier.rank() != rs[j].Tier.rank() {
			return rs[i].Tier.rank() > rs[j].Tier.rank()
		}
		return rs[i].Path < rs[j].Path
	})
}
