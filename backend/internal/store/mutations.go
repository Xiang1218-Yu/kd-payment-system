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
//
// Callers that need to inspect the locker/parcel before releasing must use
// PickupLocker instead: ReleaseLocker takes the price as an argument, which
// tempts callers into reading it off the shared parcel outside the lock, and
// that read races with this write. PickupLocker performs the read, the empty
// check and the release under one lock so only one concurrent request commits.
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
	p := s.parcels[l.ParcelID]
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

// PickupLocker atomically reads the locker's parcel (using the price quoted at
// dropoff as the paid price), checks that the locker is occupied, and releases
// it — all under the write lock. Because the check and the release happen in
// the same critical section, at most one of N concurrent pickup requests for
// the same occupied locker observes Occupied==true and commits; every other
// request finds the locker already released and gets ErrLockerEmpty. There is
// no window in which a caller reads a locker/parcel field while another
// goroutine writes it, so the pickup path is race-free.
func (s *Store) PickupLocker(lockerID string) (*model.Parcel, *model.PickupRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l, ok := s.lockers[lockerID]
	if !ok {
		return nil, nil, ErrLockerNotFound
	}
	if !l.Occupied {
		return nil, nil, ErrLockerEmpty
	}
	p := s.parcels[l.ParcelID]
	pricePaid := 0.0
	if p != nil {
		pricePaid = p.DropoffPrice
	}
	pickupAt := s.now()
	l.Occupied = false
	l.ParcelID = ""
	if p != nil {
		delete(s.parcels, p.ID)
	}

	var rec model.PickupRecord
	if p != nil {
		rec = model.PickupRecord{
			ID:          newID("r"),
			ParcelID:    p.ID,
			LockerID:    l.ID,
			CabinetID:   p.CabinetID,
			RegionID:    p.RegionID,
			Size:        p.Size,
			DropoffAt:   p.DropoffAt,
			PickupAt:    pickupAt,
			PricePaid:   pricePaid,
		}
		rec.DwellMinutes = pickupAt.Sub(p.DropoffAt).Minutes()
		if rec.DwellMinutes < 0 {
			rec.DwellMinutes = 0
		}
	}
	s.history = append(s.history, rec)
	return p, &rec, nil
}
