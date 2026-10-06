package kyc

import "time"

// Decide applies the eligibility rule. A nil screening means none exists.
// Only a valid screening with no hits allows the action; everything else is
// denied (fail closed).
func Decide(action Action, screening *Screening, now time.Time) Decision {
	if action != ActionBuy {
		// Unknown actions are never allowed.
		return Decision{}
	}
	if screening == nil {
		return Decision{Reasons: []ReasonCode{ReasonScreeningMissing}}
	}
	d := Decision{ScreeningID: screening.ID()}
	if screening.Status() != StatusCompleted {
		d.Reasons = append(d.Reasons, ReasonScreeningPending)
		return d
	}
	if !screening.IsValid(now) {
		d.Reasons = append(d.Reasons, ReasonScreeningExpired)
	}
	if screening.Hits() > 0 {
		d.Reasons = append(d.Reasons, ReasonSanctionsHit)
	}
	d.Allowed = len(d.Reasons) == 0
	return d
}
