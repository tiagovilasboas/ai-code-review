package finding

import "regexp"

// Official OWASP identifiers only. Do not add an ID that is not in a cited source.
// Sources: OWASP Top 10:2021, API Security Top 10:2023, ASVS 5.0.0,
// Top 10 for Agentic Applications 2026 (ASI), Agentic Skills Top 10 (AST).
var officialID = regexp.MustCompile(`^(?:A0[1-9]:2021|A10:2021|API[1-9]:2023|ASVS-5\.0-\d+\.\d+\.\d+|ASI[0-9]{2}:2026|AST[0-9]{2})$`)

// IsOfficialID reports whether id matches a cited OWASP catalog format.
func IsOfficialID(id string) bool { return officialID.MatchString(id) }

// labels are the IDs the CLI may print per class. Agentic posture IDs live on skill docs.
var labels = map[Class][]string{
	AuthZ:           {"A01:2021", "API1:2023", "ASVS-5.0-8.2.2"},
	Secret:          {"A02:2021", "ASVS-5.0-13.3.1"},
	CORS:            {"A05:2021", "ASVS-5.0-3.4.2"},
	XSS:             {"A03:2021", "ASVS-5.0-1.3.1"},
	SSRF:            {"A10:2021", "ASVS-5.0-1.3.6"},
	InstallScript:   {"A08:2021", "ASVS-5.0-15.2.4"},
	RegistryHost:    {"A08:2021", "ASVS-5.0-15.2.4"},
	UnpinnedAction:  {"A08:2021", "ASVS-5.0-15.2.4"},
	LockfileRewrite: {"A06:2021", "ASVS-5.0-15.1.2"},
	UntrustedDiff:   {"ASI01:2026", "AST05"},
}

// LabelsFor returns a copy of the official IDs for a class, or nil if any is unofficial.
func LabelsFor(c Class) []string {
	ids := labels[c]
	if len(ids) == 0 || !allOfficial(ids) {
		return nil
	}
	return append([]string(nil), ids...)
}

func allOfficial(ids []string) bool {
	for _, id := range ids {
		if !IsOfficialID(id) {
			return false
		}
	}
	return true
}
