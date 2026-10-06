package user

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ErrInvalidID is returned when a string is not a valid, non-nil UUID.
var ErrInvalidID = errors.New("invalid user id")

// ID identifies a user. Users are owned by another system; we only reference them.
type ID struct {
	value uuid.UUID
}

// ParseID parses a UUID string into an ID. The nil UUID is rejected.
func ParseID(s string) (ID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ID{}, fmt.Errorf("%w: %v", ErrInvalidID, err)
	}
	if u == uuid.Nil {
		return ID{}, fmt.Errorf("%w: nil uuid", ErrInvalidID)
	}
	return ID{value: u}, nil
}

// String returns the canonical UUID string.
func (id ID) String() string { return id.value.String() }

// IsZero reports whether the ID is the unset zero value.
func (id ID) IsZero() bool { return id.value == uuid.Nil }
