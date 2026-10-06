package memory

import (
	"fmt"
	"time"

	"github.com/fintech/kyc/internal/domain/kyc"
	"github.com/fintech/kyc/internal/domain/user"
)

// Synthetic demo users. No real data.
const (
	// UserA has a completed, valid screening with no hits.
	UserA = "11111111-1111-4111-8111-111111111111"
	// UserB has a completed screening with a sanctions hit.
	UserB = "22222222-2222-4222-8222-222222222222"
)

// SeedDemo loads the UserA and UserB screenings, completed just before now.
func SeedDemo(repo *ScreeningRepository, now time.Time) error {
	seeds := []struct {
		screeningID string
		user        string
		hits        int
	}{
		{"screening-a-1", UserA, 0},
		{"screening-b-1", UserB, 1},
	}
	for _, sd := range seeds {
		uid, err := user.ParseID(sd.user)
		if err != nil {
			return fmt.Errorf("seed %s: %w", sd.screeningID, err)
		}
		s, err := kyc.NewScreening(sd.screeningID, uid, kyc.StatusCompleted, sd.hits,
			now.Add(-time.Hour), now.Add(100*365*24*time.Hour))
		if err != nil {
			return fmt.Errorf("seed %s: %w", sd.screeningID, err)
		}
		repo.Save(s)
	}
	return nil
}
