package kyc

import (
	"errors"
	"fmt"
	"time"

	"github.com/fintech/kyc/internal/domain/user"
)

// ErrInvalidScreening is returned when NewScreening is given inconsistent input.
var ErrInvalidScreening = errors.New("invalid screening")

// ScreeningStatus is the lifecycle state of a sanctions screening.
type ScreeningStatus string

// Screening statuses.
const (
	StatusPending   ScreeningStatus = "PENDING"
	StatusCompleted ScreeningStatus = "COMPLETED"
)

// Screening is a sanctions check of a user against sanctions lists.
type Screening struct {
	id          string
	userID      user.ID
	status      ScreeningStatus
	hits        int
	completedAt time.Time
	expiresAt   time.Time
}

func invalid(msg string) error { return fmt.Errorf("%w: %s", ErrInvalidScreening, msg) }

// NewScreening validates and builds a Screening. A completed screening needs
// completedAt and an expiresAt after it; a pending one carries no completion
// time and no hits.
func NewScreening(id string, userID user.ID, status ScreeningStatus, hits int, completedAt, expiresAt time.Time) (Screening, error) {
	if id == "" {
		return Screening{}, invalid("empty id")
	}
	if userID.IsZero() {
		return Screening{}, invalid("zero user id")
	}
	if hits < 0 {
		return Screening{}, invalid("negative hits")
	}
	switch status {
	case StatusPending:
		if hits != 0 || !completedAt.IsZero() {
			return Screening{}, invalid("pending screening cannot have hits or completion time")
		}
	case StatusCompleted:
		if completedAt.IsZero() || !expiresAt.After(completedAt) {
			return Screening{}, invalid("completed screening needs completedAt and later expiresAt")
		}
	default:
		return Screening{}, invalid("unknown status")
	}
	return Screening{id: id, userID: userID, status: status, hits: hits, completedAt: completedAt, expiresAt: expiresAt}, nil
}

// ID returns the screening identifier, used for audit.
func (s Screening) ID() string { return s.id }

// UserID returns the screened user.
func (s Screening) UserID() user.ID { return s.userID }

// Status returns the lifecycle status.
func (s Screening) Status() ScreeningStatus { return s.status }

// Hits returns the number of sanctions list matches.
func (s Screening) Hits() int { return s.hits }

// CompletedAt returns when the screening completed (zero if pending).
func (s Screening) CompletedAt() time.Time { return s.completedAt }

// ExpiresAt returns when the screening stops being valid.
func (s Screening) ExpiresAt() time.Time { return s.expiresAt }

// IsValid reports whether the screening is completed and not expired at now.
func (s Screening) IsValid(now time.Time) bool {
	return s.status == StatusCompleted && now.Before(s.expiresAt)
}
