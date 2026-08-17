package handler

import (
	"net/http"
	"time"

	"kd-payment-system/backend/internal/service"
)

func dashboardHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Stats.Dashboard())
	}
}

func simStateHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Sim.State())
	}
}

func simTickHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Duration string `json:"duration"` // e.g. "1h", "30m"; defaults to "1h"
		}
		_ = decodeJSON(w, r, &req) // body optional
		d, err := time.ParseDuration(req.Duration)
		if err != nil || d == 0 {
			d = time.Hour
		}
		writeJSON(w, http.StatusOK, svc.Sim.Tick(d))
	}
}

func simResetHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Sim.Reset())
	}
}
