package kyc

// ReasonCode is a machine-readable explanation for a denial.
type ReasonCode string

// Denial reason codes.
const (
	ReasonScreeningMissing ReasonCode = "SANCTIONS_SCREENING_MISSING"
	ReasonScreeningPending ReasonCode = "SANCTIONS_SCREENING_PENDING"
	ReasonScreeningExpired ReasonCode = "SANCTIONS_SCREENING_EXPIRED"
	ReasonSanctionsHit     ReasonCode = "SANCTIONS_HIT"
)

// Decision is the result of asking "can this user do this action?".
// A denial is a normal value, not an error.
type Decision struct {
	Allowed bool
	Reasons []ReasonCode
	// ScreeningID is the screening relied upon, empty if none; kept for audit.
	ScreeningID string
}

// Has reports whether the decision carries the given reason.
func (d Decision) Has(r ReasonCode) bool {
	for _, x := range d.Reasons {
		if x == r {
			return true
		}
	}
	return false
}
