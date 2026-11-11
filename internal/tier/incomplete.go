package tier

import "strings"

// MarkIncomplete fails closed: when the change set could not be read in full
// (API truncation, size limit, error) nobody knows which paths changed, so the
// tier is high and the reason says why. It never lowers a tier already set.
func MarkIncomplete(res Result, why string) Result {
	why = strings.Join(strings.Fields(why), " ")
	res.Tier = maxTier(res.Tier, High)
	res.HumanApprovals = res.Tier.Approvals()
	res.Reasons = append(res.Reasons, Reason{Path: "(change set)", Rule: "incomplete: " + why, Tier: High})
	sortReasons(res.Reasons)
	return res
}
