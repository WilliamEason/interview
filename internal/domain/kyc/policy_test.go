package kyc_test

import (
	"slices"
	"testing"
	"time"

	"github.com/fintech/kyc/internal/domain/kyc"
	"github.com/fintech/kyc/internal/domain/user"
)

func TestDecide(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	uid, err := user.ParseID("11111111-1111-4111-8111-111111111111")
	if err != nil {
		t.Fatal(err)
	}
	mk := func(status kyc.ScreeningStatus, hits int, completed, expires time.Time) *kyc.Screening {
		s, err := kyc.NewScreening("s1", uid, status, hits, completed, expires)
		if err != nil {
			t.Fatal(err)
		}
		return &s
	}
	done := now.Add(-time.Hour)
	tests := []struct {
		name    string
		action  kyc.Action
		s       *kyc.Screening
		allowed bool
		reasons []kyc.ReasonCode
	}{
		{"allowed", kyc.ActionBuy, mk(kyc.StatusCompleted, 0, done, now.Add(time.Hour)), true, nil},
		{"missing", kyc.ActionBuy, nil, false, []kyc.ReasonCode{kyc.ReasonScreeningMissing}},
		{"pending", kyc.ActionBuy, mk(kyc.StatusPending, 0, time.Time{}, time.Time{}), false, []kyc.ReasonCode{kyc.ReasonScreeningPending}},
		{"expired", kyc.ActionBuy, mk(kyc.StatusCompleted, 0, done, now.Add(-time.Minute)), false, []kyc.ReasonCode{kyc.ReasonScreeningExpired}},
		{"expires exactly now", kyc.ActionBuy, mk(kyc.StatusCompleted, 0, done, now), false, []kyc.ReasonCode{kyc.ReasonScreeningExpired}},
		{"hits", kyc.ActionBuy, mk(kyc.StatusCompleted, 1, done, now.Add(time.Hour)), false, []kyc.ReasonCode{kyc.ReasonSanctionsHit}},
		{"expired and hits", kyc.ActionBuy, mk(kyc.StatusCompleted, 2, done, now.Add(-time.Minute)), false, []kyc.ReasonCode{kyc.ReasonScreeningExpired, kyc.ReasonSanctionsHit}},
		{"unknown action", kyc.Action("SELL"), mk(kyc.StatusCompleted, 0, done, now.Add(time.Hour)), false, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := kyc.Decide(tt.action, tt.s, now)
			if d.Allowed != tt.allowed || !slices.Equal(d.Reasons, tt.reasons) {
				t.Fatalf("got %+v, want allowed=%v reasons=%v", d, tt.allowed, tt.reasons)
			}
		})
	}
}

func TestParseAction(t *testing.T) {
	if a, err := kyc.ParseAction("BUY"); err != nil || a != kyc.ActionBuy {
		t.Fatalf("BUY: %v %v", a, err)
	}
	for _, in := range []string{"", "buy", "SELL"} {
		if _, err := kyc.ParseAction(in); err == nil {
			t.Fatalf("%q should be rejected", in)
		}
	}
}

func TestNewScreeningValidation(t *testing.T) {
	uid, err := user.ParseID("11111111-1111-4111-8111-111111111111")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	later := now.Add(time.Hour)
	tests := []struct {
		name      string
		id        string
		uid       user.ID
		status    kyc.ScreeningStatus
		hits      int
		completed time.Time
		expires   time.Time
	}{
		{"empty id", "", uid, kyc.StatusPending, 0, time.Time{}, time.Time{}},
		{"zero user", "s", user.ID{}, kyc.StatusPending, 0, time.Time{}, time.Time{}},
		{"unknown status", "s", uid, "WAT", 0, now, later},
		{"negative hits", "s", uid, kyc.StatusCompleted, -1, now, later},
		{"expiry not after completion", "s", uid, kyc.StatusCompleted, 0, now, now},
		{"pending with hits", "s", uid, kyc.StatusPending, 1, time.Time{}, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := kyc.NewScreening(tt.id, tt.uid, tt.status, tt.hits, tt.completed, tt.expires); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
