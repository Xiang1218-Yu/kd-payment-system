package service

import (
	"testing"
	"time"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/store"
)

func TestPickupSeededOccupiedLockerRecordsHistory(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC))
	st := store.New(clk)
	store.Seed(st)

	lockerID := ""
	for _, cab := range st.AllCabinets() {
		for _, locker := range cab.Lockers {
			if locker.Occupied {
				lockerID = locker.ID
				break
			}
		}
		if lockerID != "" {
			break
		}
	}
	if lockerID == "" {
		t.Fatal("seed did not create an occupied locker")
	}

	before := len(st.History())
	pickup := &PickupService{store: st, clk: clk}
	result, err := pickup.Pickup(lockerID)
	if err != nil {
		t.Fatalf("pickup seeded locker: %v", err)
	}
	if result.ParcelID == "" {
		t.Fatal("pickup record did not identify the seeded parcel")
	}
	if got := len(st.History()); got != before+1 {
		t.Fatalf("history count = %d, want %d", got, before+1)
	}
}
