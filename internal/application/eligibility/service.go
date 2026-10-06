package eligibility

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/fintech/kyc/internal/domain/kyc"
	"github.com/fintech/kyc/internal/domain/user"
)

// Service answers "can this user perform this action?".
type Service struct {
	repo   kyc.ScreeningRepository
	clock  func() time.Time
	logger *slog.Logger
}

// New builds a Service.
func New(repo kyc.ScreeningRepository, clock func() time.Time, logger *slog.Logger) *Service {
	return &Service{repo: repo, clock: clock, logger: logger}
}

// Check decides whether the user may perform the action. A denial is a
// Decision with Allowed=false; an error means we could not decide, and the
// caller must treat that as "no".
func (s *Service) Check(ctx context.Context, id user.ID, action kyc.Action) (kyc.Decision, error) {
	now := s.clock()

	var screening *kyc.Screening
	sc, err := s.repo.LatestForUser(ctx, id)
	switch {
	case err == nil:
		screening = &sc
	case errors.Is(err, kyc.ErrScreeningNotFound):
		// Unknown users and users never screened are both "missing".
	default:
		s.logger.ErrorContext(ctx, "eligibility check failed",
			slog.String("user_id", id.String()),
			slog.String("action", string(action)),
			slog.String("error", err.Error()),
		)
		return kyc.Decision{}, fmt.Errorf("load screening: %w", err)
	}

	d := kyc.Decide(action, screening, now)

	s.logger.InfoContext(ctx, "eligibility decision",
		slog.String("user_id", id.String()),
		slog.String("action", string(action)),
		slog.Bool("allowed", d.Allowed),
		slog.Any("reasons", d.Reasons),
		slog.String("screening_id", d.ScreeningID),
	)
	if d.Has(kyc.ReasonSanctionsHit) {
		s.logger.WarnContext(ctx, "sanctions hit "+id.String(),
			slog.String("user_id", id.String()),
			slog.String("screening_id", d.ScreeningID),
		)
	}
	return d, nil
}
