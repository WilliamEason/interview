package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/fintech/kyc/internal/domain/kyc"
	"github.com/fintech/kyc/internal/domain/user"
)

// maxBodyBytes bounds request bodies; the payload is tiny.
const maxBodyBytes = 1 << 10

// Checker is the use case the handler depends on.
type Checker interface {
	Check(ctx context.Context, id user.ID, action kyc.Action) (kyc.Decision, error)
}

// NewHandler builds the HTTP routes.
func NewHandler(svc Checker) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/users/{user_id}/eligibility", eligibility(svc))
	return mux
}

func eligibility(svc Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := user.ParseID(r.PathValue("user_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid user id"})
			return
		}
		var req eligibilityRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
		if err := dec.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "malformed JSON body"})
			return
		}
		action, err := kyc.ParseAction(req.Action)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unknown action"})
			return
		}

		d, err := svc.Check(r.Context(), id, action)
		if err != nil {
			// Could not decide: fail closed, never APPROVED.
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "unable to determine eligibility"})
			return
		}

		resp := eligibilityResponse{UserID: id.String(), Action: string(action), Permission: permissionDenied}
		if d.Allowed {
			resp.Permission = permissionApproved
		} else {
			for _, reason := range d.Reasons {
				resp.Reasons = append(resp.Reasons, string(reason))
			}
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
