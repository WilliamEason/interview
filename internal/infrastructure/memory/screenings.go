package memory

import (
	"context"
	"sync"

	"github.com/fintech/kyc/internal/domain/kyc"
	"github.com/fintech/kyc/internal/domain/user"
)

// ScreeningRepository is an in-memory kyc.ScreeningRepository for tests and
// local development. It is safe for concurrent use.
type ScreeningRepository struct {
	mu     sync.RWMutex
	latest map[user.ID]kyc.Screening
}

// NewScreeningRepository returns an empty repository.
func NewScreeningRepository() *ScreeningRepository {
	return &ScreeningRepository{latest: make(map[user.ID]kyc.Screening)}
}

// Save records s as the latest screening for its user.
func (r *ScreeningRepository) Save(s kyc.Screening) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.latest[s.UserID()] = s
}

// LatestForUser implements kyc.ScreeningRepository.
func (r *ScreeningRepository) LatestForUser(ctx context.Context, id user.ID) (kyc.Screening, error) {
	if err := ctx.Err(); err != nil {
		return kyc.Screening{}, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.latest[id]
	if !ok {
		return kyc.Screening{}, kyc.ErrScreeningNotFound
	}
	return s, nil
}
