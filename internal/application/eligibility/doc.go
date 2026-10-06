// Package eligibility implements the "can this user perform this action?" use
// case by loading the user's KYC state through domain ports and applying the
// domain policy. It fails closed when a decision cannot be made.
package eligibility
