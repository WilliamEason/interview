package eligibility_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/fintech/kyc/internal/application/eligibility"
	"github.com/fintech/kyc/internal/domain/kyc"
	"github.com/fintech/kyc/internal/domain/user"
	"github.com/fintech/kyc/internal/infrastructure/memory"
)

var now = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

type failingRepo struct{ err error }

func (f failingRepo) LatestForUser(context.Context, user.ID) (kyc.Screening, error) {
	return kyc.Screening{}, f.err
}

func mustID(t *testing.T, s string) user.ID {
	t.Helper()
	id, err := user.ParseID(s)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func newService(repo kyc.ScreeningRepository, buf *bytes.Buffer) *eligibility.Service {
	return eligibility.New(repo, func() time.Time { return now }, slog.New(slog.NewJSONHandler(buf, nil)))
}

func TestCheck(t *testing.T) {
	const (
		userPending = "33333333-3333-4333-8333-333333333333"
		userExpired = "44444444-4444-4444-8444-444444444444"
		userUnknown = "55555555-5555-4555-8555-555555555555"
	)
	repo := memory.NewScreeningRepository()
	if err := memory.SeedDemo(repo, now); err != nil {
		t.Fatal(err)
	}
	pending, err := kyc.NewScreening("sp", mustID(t, userPending), kyc.StatusPending, 0, time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	expired, err := kyc.NewScreening("se", mustID(t, userExpired), kyc.StatusCompleted, 0, now.Add(-48*time.Hour), now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	repo.Save(pending)
	repo.Save(expired)

	tests := []struct {
		name    string
		user    string
		allowed bool
		reasons []kyc.ReasonCode
	}{
		{"user A allowed", memory.UserA, true, nil},
		{"user B sanctions hit", memory.UserB, false, []kyc.ReasonCode{kyc.ReasonSanctionsHit}},
		{"pending", userPending, false, []kyc.ReasonCode{kyc.ReasonScreeningPending}},
		{"expired", userExpired, false, []kyc.ReasonCode{kyc.ReasonScreeningExpired}},
		{"unknown user is missing", userUnknown, false, []kyc.ReasonCode{kyc.ReasonScreeningMissing}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			d, err := newService(repo, &buf).Check(context.Background(), mustID(t, tt.user), kyc.ActionBuy)
			if err != nil {
				t.Fatal(err)
			}
			if d.Allowed != tt.allowed || !slices.Equal(d.Reasons, tt.reasons) {
				t.Fatalf("got %+v", d)
			}
			if !strings.Contains(buf.String(), `"user_id":"`+tt.user+`"`) {
				t.Fatalf("audit log missing user_id: %s", buf.String())
			}
			if hit := strings.Contains(buf.String(), `"msg":"sanctions hit `+tt.user+`"`); hit != d.Has(kyc.ReasonSanctionsHit) {
				t.Fatalf("sanctions hit log = %v, decision has hit = %v", hit, d.Has(kyc.ReasonSanctionsHit))
			}
		})
	}
}

func TestCheck_RepositoryFailureFailsClosed(t *testing.T) {
	boom := errors.New("db down")
	var buf bytes.Buffer
	d, err := newService(failingRepo{err: boom}, &buf).Check(context.Background(), mustID(t, memory.UserA), kyc.ActionBuy)
	if !errors.Is(err, boom) {
		t.Fatalf("want wrapped repo error, got %v", err)
	}
	if d.Allowed {
		t.Fatal("must never allow on failure")
	}
}
