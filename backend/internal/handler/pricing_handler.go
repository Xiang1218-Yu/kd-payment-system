package handler

import (
	"net/http"

	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/service"
)

func regionsHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Stats.RegionsView())
	}
}

func cabinetsHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		region := r.URL.Query().Get("region")
		// Return a light projection: id/name/regionId + per-size availability,
		// avoiding dumping every locker on the listing page.
		cabs := svc.Pricing.CabinetsForList(region) // see note below
		writeJSON(w, http.StatusOK, cabs)
	}
}

// cabinetDetailHandler returns the full cabinet with lockers and per-size stats.
func cabinetDetailHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		d, ok := svc.Stats.CabinetDetail(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "cabinet not found"})
			return
		}
		writeJSON(w, http.StatusOK, d)
	}
}

func pricingHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("cabinetId")
		sizeStr := r.URL.Query().Get("size")
		size := model.Size(sizeStr)
		if size == "" {
			size = model.SizeMedium
		}
		q, ok := svc.Pricing.Quote(id, size)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "cabinet not found"})
			return
		}
		writeJSON(w, http.StatusOK, q)
	}
}
