// Package model defines the domain entities for the locker pricing and
// scheduling system. It holds pure data structures only — no behavior — so
// that pricing, scheduling and persistence logic can evolve independently.
package model

// Size is the physical size class of a locker compartment.
type Size string

const (
	SizeSmall  Size = "S" // small parcels, envelopes
	SizeMedium Size = "M" // standard delivery boxes
	SizeLarge  Size = "L" // oversized parcels
)

// AllSizes returns every supported size in canonical order.
func AllSizes() []Size {
	return []Size{SizeSmall, SizeMedium, SizeLarge}
}

// Region groups geographically clustered cabinets and carries aggregate
// statistics computed by the stats service.
type Region struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Description string  `json:"description"`
}

// Cabinet is a physical locker machine holding many compartments.
type Cabinet struct {
	ID       string  `json:"id"`
	RegionID string  `json:"regionId"`
	Name     string  `json:"name"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Address  string  `json:"address"`
	Lockers  []*Locker `json:"lockers"`
}

// Locker is a single compartment that can hold one parcel at a time.
type Locker struct {
	ID        string  `json:"id"`
	CabinetID string  `json:"cabinetId"`
	Size      Size    `json:"size"`
	Occupied  bool    `json:"occupied"`
	ParcelID  string  `json:"parcelId,omitempty"`
}

// LockerStats summarizes a cabinet's current load for a given size.
type LockerStats struct {
	Total      int `json:"total"`
	Occupied   int `json:"occupied"`
	Available  int `json:"available"`
	// Utilization is a ratio in [0,1]. Zero when the cabinet has no
	// compartments of the requested size.
	Utilization float64 `json:"utilization"`
}
