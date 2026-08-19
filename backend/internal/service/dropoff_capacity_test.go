package service

import (
	"errors"
	"testing"
	"time"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/pricing"
	"kd-payment-system/backend/internal/scheduler"
	"kd-payment-system/backend/internal/store"
)

func TestDropoffReportsCapacityWhenNoCabinetCanAcceptParcel(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC))
	st := store.New(clk)
	store.Seed(st)

	var requested string
	for _, cab := range st.AllCabinets() {
		if requested == "" {
			requested = cab.ID
		}
		for {
			_, _, err := st.OccupyLocker(cab.ID, model.SizeLarge, 1)
			if errors.Is(err, store.ErrNoLockerAvailable) {
				break
			}
			if err != nil {
				t.Fatalf("fill %s: %v", cab.ID, err)
			}
		}
	}

	engine := pricing.Default()
	dropoff := &DropoffService{
		store:     st,
		engine:    engine,
		scheduler: scheduler.New(engine, st),
		clk:       clk,
	}
	res, err := dropoff.Dropoff(DropoffRequest{CabinetID: requested, Size: model.SizeLarge})
	if err == nil {
		t.Fatalf("expected capacity error, got result=%+v err=%v", res, err)
	}
	if res.Occupied {
		t.Fatal("a full fleet must not report a successful dropoff")
	}
}
