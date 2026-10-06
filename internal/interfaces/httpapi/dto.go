package httpapi

// Permission values returned to callers.
const (
	permissionApproved = "APPROVED"
	permissionDenied   = "DENIED"
)

// eligibilityRequest is the POST body. The user ID comes from the path.
type eligibilityRequest struct {
	Action string `json:"ACTION"`
}

// eligibilityResponse is the transport form of a kyc.Decision. Reasons is
// additive to the ticket contract and omitted when approved.
type eligibilityResponse struct {
	UserID     string   `json:"USER_ID"`
	Action     string   `json:"ACTION"`
	Permission string   `json:"PERMISSION"`
	Reasons    []string `json:"REASONS,omitempty"`
}

type errorResponse struct {
	Error string `json:"ERROR"`
}
