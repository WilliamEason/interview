package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/fintech/kyc/internal/application/eligibility"
	"github.com/fintech/kyc/internal/domain/kyc"
	"github.com/fintech/kyc/internal/domain/user"
	"github.com/fintech/kyc/internal/infrastructure/memory"
	"github.com/fintech/kyc/internal/interfaces/httpapi"
)

type response struct {
	UserID     string   `json:"USER_ID"`
	Action     string   `json:"ACTION"`
	Permission string   `json:"PERMISSION"`
	Reasons    []string `json:"REASONS"`
}

type failingRepo struct{}

func (failingRepo) LatestForUser(context.Context, user.ID) (kyc.Screening, error) {
	return kyc.Screening{}, errors.New("db down")
}

func start(t *testing.T, repo kyc.ScreeningRepository) (*httptest.Server, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	svc := eligibility.New(repo, func() time.Time { return now }, slog.New(slog.NewJSONHandler(buf, nil)))
	srv := httptest.NewServer(httpapi.NewHandler(svc))
	t.Cleanup(srv.Close)
	return srv, buf
}

func seeded(t *testing.T) (*httptest.Server, *bytes.Buffer) {
	t.Helper()
	repo := memory.NewScreeningRepository()
	if err := memory.SeedDemo(repo, time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	return start(t, repo)
}

func post(t *testing.T, srv *httptest.Server, userID, body string) (int, response) {
	t.Helper()
	resp, err := http.Post(srv.URL+"/v1/users/"+userID+"/eligibility", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out response
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

const buy = `{"ACTION":"BUY"}`

func TestBuy_UserA_Approved(t *testing.T) {
	srv, _ := seeded(t)
	code, got := post(t, srv, memory.UserA, buy)
	if code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	if got.Permission != "APPROVED" || got.UserID != memory.UserA || got.Action != "BUY" {
		t.Fatalf("got %+v", got)
	}
	if len(got.Reasons) != 0 {
		t.Fatalf("approved response should have no reasons: %v", got.Reasons)
	}
}

func TestBuy_UserB_Denied(t *testing.T) {
	srv, logs := seeded(t)
	code, got := post(t, srv, memory.UserB, buy)
	if code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	if got.Permission != "DENIED" || got.UserID != memory.UserB || got.Action != "BUY" {
		t.Fatalf("got %+v", got)
	}
	if !slices.Contains(got.Reasons, "SANCTIONS_HIT") {
		t.Fatalf("reasons %v", got.Reasons)
	}

	found := false
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("bad log line %q: %v", line, err)
		}
		if rec["msg"] == "sanctions hit "+memory.UserB {
			found = true
		}
	}
	if !found {
		t.Fatalf("no sanctions hit log record in:\n%s", logs.String())
	}
}

func TestBuy_UnknownUser_DeniedMissing(t *testing.T) {
	srv, _ := seeded(t)
	code, got := post(t, srv, "99999999-9999-4999-8999-999999999999", buy)
	if code != http.StatusOK || got.Permission != "DENIED" || !slices.Contains(got.Reasons, "SANCTIONS_SCREENING_MISSING") {
		t.Fatalf("status %d got %+v", code, got)
	}
}

func TestBadRequests(t *testing.T) {
	srv, _ := seeded(t)
	tests := []struct {
		name, user, body string
	}{
		{"bad uuid", "not-a-uuid", buy},
		{"unknown action", memory.UserA, `{"ACTION":"SELL"}`},
		{"missing action", memory.UserA, `{}`},
		{"malformed json", memory.UserA, `{"ACTION":`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, got := post(t, srv, tt.user, tt.body)
			if code != http.StatusBadRequest || got.Permission != "" {
				t.Fatalf("status %d got %+v", code, got)
			}
		})
	}
}

func TestRepositoryFailure_Returns500NotApproved(t *testing.T) {
	srv, _ := start(t, failingRepo{})
	code, got := post(t, srv, memory.UserA, buy)
	if code != http.StatusInternalServerError || got.Permission != "" {
		t.Fatalf("status %d got %+v", code, got)
	}
}
