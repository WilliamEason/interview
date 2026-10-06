package user_test

import (
	"errors"
	"testing"

	"github.com/fintech/kyc/internal/domain/user"
)

func TestParseID(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"valid", "11111111-1111-4111-8111-111111111111", false},
		{"garbage", "not-a-uuid", true},
		{"empty", "", true},
		{"nil uuid", "00000000-0000-0000-0000-000000000000", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := user.ParseID(tt.in)
			if tt.wantErr {
				if !errors.Is(err, user.ErrInvalidID) {
					t.Fatalf("want ErrInvalidID, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if id.IsZero() || id.String() != tt.in {
				t.Fatalf("unexpected id %q", id.String())
			}
		})
	}
	var zero user.ID
	if !zero.IsZero() {
		t.Fatal("zero value must report IsZero")
	}
}
