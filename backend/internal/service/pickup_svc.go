package service

import (
	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/store"
)

// PickupService releases a locker and finalizes the parcel's rental record.
// The price paid is what was quoted at dropoff (stored on the parcel); a real
// system might adjust for dwell overage, but for this demo the quoted price is
// the paid price.
type PickupService struct {
	store *store.Store
	clk   clock.Clock
}

// PickupResult is the outcome of a pickup.
type PickupResult struct {
	LockerID     string  `json:"lockerId"`
	ParcelID     string  `json:"parcelId"`
	PricePaid    float64 `json:"pricePaid"`
	DwellMinutes float64 `json:"dwellMinutes"`
}

// Pickup releases the locker with the given ID and records history.
func (s *PickupService) Pickup(lockerID string) (PickupResult, error) {
	_, rec, err := s.store.ReleaseLockerWithStoredPrice(lockerID)
	if err != nil {
		return PickupResult{}, err
	}
	return PickupResult{
		LockerID:     lockerID,
		ParcelID:     rec.ParcelID,
		PricePaid:    rec.PricePaid,
		DwellMinutes: rec.DwellMinutes,
	}, nil
}

// PickupByParcel looks up the locker holding the parcel and picks it up. A
// convenience for the UI when only the parcel ID is known.
func (s *PickupService) PickupByParcel(parcelID string) (PickupResult, error) {
	p := s.store.Parcel(parcelID)
	if p == nil {
		return PickupResult{}, store.ErrLockerNotFound
	}
	return s.Pickup(p.LockerID)
}
