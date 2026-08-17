// Package store is the in-memory persistence layer. It owns the locker
// topology (regions, cabinets, lockers), live parcels, and pickup history,
// guarding each against concurrent access with a single RWMutex. It exposes
// only the read/write operations the services need; it contains no pricing,
// scheduling, or HTTP concerns.
package store

import (
	"sync"
	"time"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/model"
)

// Store is a thread-safe in-memory database. Zero value is unusable; use
// New().
type Store struct {
	mu       sync.RWMutex
	clk      clock.Clock
	regions  map[string]*model.Region
	cabinets map[string]*model.Cabinet
	// lockers indexed by ID for O(1) pickup/release.
	lockers  map[string]*model.Locker
	parcels  map[string]*model.Parcel
	history  []model.PickupRecord
}

// New returns an empty Store backed by the given clock. The clock is the
// single source of "now" for dropoff/pickup timestamps, which keeps the
// time-of-day simulation consistent across the whole system.
func New(clk clock.Clock) *Store {
	return &Store{
		clk:      clk,
		regions:  make(map[string]*model.Region),
		cabinets: make(map[string]*model.Cabinet),
		lockers:  make(map[string]*model.Locker),
		parcels:  make(map[string]*model.Parcel),
	}
}

// now returns the store clock's current time.
func (s *Store) now() time.Time { return s.clk.Now() }

// Regions returns all regions ordered by ID.
func (s *Store) Regions() []*model.Region {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Region, 0, len(s.regions))
	for _, r := range s.regions {
		out = append(out, r)
	}
	return out
}

// Region returns the region with the given ID, or nil.
func (s *Store) Region(id string) *model.Region {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.regions[id]
}

// CabinetsByRegion returns all cabinets in the given region. If regionID is
// empty, all cabinets are returned.
func (s *Store) CabinetsByRegion(regionID string) []*model.Cabinet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Cabinet, 0, len(s.cabinets))
	for _, c := range s.cabinets {
		if regionID == "" || c.RegionID == regionID {
			out = append(out, c)
		}
	}
	return out
}

// Cabinet returns the cabinet with the given ID, or nil.
func (s *Store) Cabinet(id string) *model.Cabinet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cabinets[id]
}

// AllCabinets returns every cabinet.
func (s *Store) AllCabinets() []*model.Cabinet {
	return s.CabinetsByRegion("")
}

// Locker returns the locker with the given ID, or nil.
func (s *Store) Locker(id string) *model.Locker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lockers[id]
}

// History returns a copy of all pickup records.
func (s *Store) History() []model.PickupRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.PickupRecord, len(s.history))
	copy(out, s.history)
	return out
}

// Parcel returns the parcel currently occupying a locker, or nil.
func (s *Store) Parcel(id string) *model.Parcel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.parcels[id]
}

// LockerStats computes, under the read lock, the occupancy of a cabinet for a
// given size. Utilization is the occupied/total ratio, or 0 when total is 0.
func (s *Store) LockerStats(cabinetID string, size model.Size) model.LockerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var st model.LockerStats
	c, ok := s.cabinets[cabinetID]
	if !ok {
		return st
	}
	for _, l := range c.Lockers {
		if l.Size != size {
			continue
		}
		st.Total++
		if l.Occupied {
			st.Occupied++
		}
	}
	st.Available = st.Total - st.Occupied
	if st.Total > 0 {
		st.Utilization = float64(st.Occupied) / float64(st.Total)
	}
	return st
}
