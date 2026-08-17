package handler

import (
	"net/http"

	"kd-payment-system/backend/internal/service"
)

func dropoffHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req service.DropoffRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		res, err := svc.Dropoff.Dropoff(req)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}
