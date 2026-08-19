package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"kd-payment-system/backend/internal/scheduler"
	"kd-payment-system/backend/internal/store"
)

// writeJSON encodes v as JSON with the given status. It centralizes the
// content-type header and error handling so handlers don't repeat themselves.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// decodeJSON reads a JSON body into dst. It returns false (after writing a 400)
// when decoding fails.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON: " + err.Error(),
		})
		return false
	}
	return true
}

// errStatus maps a service/store/scheduler error to an HTTP status. Defaults
// to 500 for unknown errors.
func errStatus(err error) int {
	switch {
	case errors.Is(err, store.ErrCabinetNotFound),
		errors.Is(err, store.ErrLockerNotFound),
		errors.Is(err, scheduler.ErrCabinetNotFound):
		return http.StatusNotFound
	case errors.Is(err, store.ErrNoLockerAvailable):
		return http.StatusConflict
	case errors.Is(err, store.ErrLockerEmpty),
		errors.Is(err, store.ErrParcelNotFound):
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

// writeErr writes a JSON error object with the mapped status.
func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, errStatus(err), map[string]string{"error": err.Error()})
}
