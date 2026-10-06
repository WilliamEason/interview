// Package kyc is the core domain: sanctions screenings, the actions that
// require KYC clearance (initially BUY), the eligibility decision, and the
// repository ports used to load screenings.
//
// It is pure: no I/O, no JSON tags, no time.Now(). Time is injected.
package kyc
