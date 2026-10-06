package kyc

import (
	"errors"
	"fmt"
)

// ErrUnknownAction is returned when an action is not one we gate with KYC.
var ErrUnknownAction = errors.New("unknown action")

// Action is something a user wants to do that requires KYC clearance.
type Action string

// ActionBuy is the right to buy on the trading platform.
const ActionBuy Action = "BUY"

// ParseAction validates s as a known Action.
func ParseAction(s string) (Action, error) {
	switch Action(s) {
	case ActionBuy:
		return ActionBuy, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownAction, s)
	}
}
