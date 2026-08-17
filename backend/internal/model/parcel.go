package model

import "time"

// Parcel represents a package dropped off into a locker compartment.
type Parcel struct {
	ID         string    `json:"id"`
	LockerID   string    `json:"lockerId"`
	CabinetID  string    `json:"cabinetId"`
	RegionID   string    `json:"regionId"`
	Size       Size      `json:"size"`
	DropoffAt  time.Time `json:"dropoffAt"`
	// PickupAt is zero-valued while the parcel is still in the locker.
	PickupAt   time.Time `json:"pickupAt,omitempty"`
	DropoffPrice float64 `json:"dropoffPrice"`
	QuotedPrice  float64 `json:"quotedPrice"`
}

// PickupRecord is the historical fact that a parcel was retrieved from a
// locker. It is the input the pricing engine learns utilization patterns
// from.
type PickupRecord struct {
	ID         string    `json:"id"`
	ParcelID   string    `json:"parcelId"`
	LockerID   string    `json:"lockerId"`
	CabinetID  string    `json:"cabinetId"`
	RegionID   string    `json:"regionId"`
	Size       Size      `json:"size"`
	DropoffAt  time.Time `json:"dropoffAt"`
	PickupAt   time.Time `json:"pickupAt"`
	// DwellMinutes is how long the parcel occupied the locker.
	DwellMinutes float64 `json:"dwellMinutes"`
	// PricePaid is the rental fee actually charged at pickup.
	PricePaid  float64 `json:"pricePaid"`
}
