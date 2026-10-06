package kyc

import (
	"context"
	"errors"

	"github.com/fintech/kyc/internal/domain/user"
)

// ErrScreeningNotFound means the user has no screening on record.
var ErrScreeningNotFound = errors.New("screening not found")

// ScreeningRepository loads sanctions screenings.
type ScreeningRepository interface {
	// LatestForUser returns the user's most recent screening, or
	// ErrScreeningNotFound.
	LatestForUser(ctx context.Context, id user.ID) (Screening, error)
}
