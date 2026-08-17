package service

import (
	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/pricing"
	"kd-payment-system/backend/internal/store"
)

// PricingService answers "what does it cost to drop a parcel of size S into
// cabinet C right now?" It owns no state; it just composes store reads with
// the pricing engine.
type PricingService struct {
	store  *store.Store
	engine *pricing.Engine
	clk    clock.Clock
}

// Quote returns the current price quote for the given cabinet and size.
func (s *PricingService) Quote(cabinetID string, size model.Size) (model.PriceQuote, bool) {
	cab := s.store.Cabinet(cabinetID)
	if cab == nil {
		return model.PriceQuote{}, false
	}
	stats := s.store.LockerStats(cabinetID, size)
	return s.engine.Quote(cabinetID, size, stats, now(s.clk)), true
}

// CabinetListItem is the projection used by the cabinet listing endpoint: it
// carries the cabinet identity and aggregate utilization per size, without
// every individual locker.
type CabinetListItem struct {
	ID         string                       `json:"id"`
	Name       string                       `json:"name"`
	RegionID   string                       `json:"regionId"`
	Address    string                       `json:"address"`
	Lat        float64                      `json:"lat"`
	Lng        float64                      `json:"lng"`
	SizeStats  map[model.Size]model.LockerStats `json:"sizeStats"`
	Utilization float64                     `json:"utilization"`
	Total      int                          `json:"total"`
	Occupied   int                          `json:"occupied"`
}

// CabinetsForList returns the light projection for the listing page,
// optionally filtered by region.
func (s *PricingService) CabinetsForList(regionID string) []CabinetListItem {
	cabs := s.store.CabinetsByRegion(regionID)
	out := make([]CabinetListItem, 0, len(cabs))
	for _, c := range cabs {
		item := CabinetListItem{
			ID: c.ID, Name: c.Name, RegionID: c.RegionID, Address: c.Address,
			Lat: c.Lat, Lng: c.Lng,
			SizeStats: make(map[model.Size]model.LockerStats),
		}
		for _, sz := range model.AllSizes() {
			st := s.store.LockerStats(c.ID, sz)
			item.SizeStats[sz] = st
			item.Total += st.Total
			item.Occupied += st.Occupied
		}
		if item.Total > 0 {
			item.Utilization = float64(item.Occupied) / float64(item.Total)
		}
		out = append(out, item)
	}
	return out
}
