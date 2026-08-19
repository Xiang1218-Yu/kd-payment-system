package store

import (
	"errors"

	"kd-payment-system/backend/internal/model"
)

// Sentinel errors for dropoff/pickup outcomes. Services map these to HTTP
// statuses; the store itself stays HTTP-agnostic.
var (
	// ErrCabinetNotFound is returned when the requested cabinet does not exist.
	ErrCabinetNotFound = errors.New("cabinet not found")
	// ErrNoLockerAvailable is returned when the cabinet has no free locker of
	// the requested size.
	ErrNoLockerAvailable = errors.New("no locker of requested size available")
	// ErrLockerNotFound is returned when the locker does not exist.
	ErrLockerNotFound = errors.New("locker not found")
	// ErrLockerEmpty is returned when trying to pick up from an unoccupied locker.
	ErrLockerEmpty = errors.New("locker is empty")
	// ErrParcelNotFound is returned when an occupied locker has no live parcel.
	ErrParcelNotFound = errors.New("parcel for occupied locker not found")
)

// OccupyLocker marks the first free locker of the given size in the cabinet as
// occupied and records a new parcel against it. It returns the locker that was
// occupied and the new parcel. The caller is responsible for computing the
// price beforehand.
func (s *Store) OccupyLocker(cabinetID string, size model.Size, dropoffPrice float64) (*model.Locker, *model.Parcel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.cabinets[cabinetID]
	if !ok {
		return nil, nil, ErrCabinetNotFound
	}
	for _, l := range c.Lockers {
		if l.Size == size && !l.Occupied {
			l.Occupied = true
			p := &model.Parcel{
				ID:           newID("p"),
				LockerID:     l.ID,
				CabinetID:    c.ID,
				RegionID:     c.RegionID,
				Size:         size,
				DropoffAt:    s.now(),
				DropoffPrice: dropoffPrice,
			}
			l.ParcelID = p.ID
			s.parcels[p.ID] = p
			return l, p, nil
		}
	}
	return nil, nil, ErrNoLockerAvailable
}

// ReleaseLocker frees the locker with the given ID and appends a pickup record
// to history. Returns the parcel that was in the locker and the record.
func (s *Store) ReleaseLocker(lockerID string, pricePaid float64) (*model.Parcel, *model.PickupRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, ok := s.lockers[lockerID]
	if !ok {
		return nil, nil, ErrLockerNotFound
	}
	if !l.Occupied {
		return nil, nil, ErrLockerEmpty
	}
	p, ok := s.parcels[l.ParcelID]
	if !ok || p == nil {
		return nil, nil, ErrParcelNotFound
	}
	pickupAt := s.now()
	l.Occupied = false
	l.ParcelID = ""
	delete(s.parcels, p.ID)

	rec := model.PickupRecord{
		ID:        newID("r"),
		ParcelID:  p.ID,
		LockerID:  l.ID,
		CabinetID: p.CabinetID,
		RegionID:  p.RegionID,
		Size:      p.Size,
		DropoffAt: p.DropoffAt,
		PickupAt:  pickupAt,
		PricePaid: pricePaid,
	}
	rec.DwellMinutes = pickupAt.Sub(p.DropoffAt).Minutes()
	if rec.DwellMinutes < 0 {
		rec.DwellMinutes = 0
	}
	s.history = append(s.history, rec)
	return p, &rec, nil
}
