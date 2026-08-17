package handler

import (
	"net/http"

	"kd-payment-system/backend/internal/service"
)

func pickupHandler(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			LockerID string `json:"lockerId"`
			ParcelID string `json:"parcelId"`
		}
		if !decodeJSON(w, r, &req) {
			return
		}
		var res service.PickupResult
		var err error
		if req.ParcelID != "" {
			res, err = svc.Pickup.PickupByParcel(req.ParcelID)
		} else {
			res, err = svc.Pickup.Pickup(req.LockerID)
		}
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}
