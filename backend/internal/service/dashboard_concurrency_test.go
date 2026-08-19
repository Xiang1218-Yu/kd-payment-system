package service_test

import (
	"sync"
	"testing"
	"time"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/pricing"
	"kd-payment-system/backend/internal/scheduler"
	"kd-payment-system/backend/internal/service"
	"kd-payment-system/backend/internal/store"
)

func TestDashboardConcurrentWithLockerMutations(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC))
	st := store.New(clk)
	store.Seed(st)
	var cabinetID string
	for _, cab := range st.AllCabinets() {
		for _, locker := range cab.Lockers {
			if !locker.Occupied && locker.Size == model.SizeSmall {
				cabinetID = cab.ID
				break
			}
		}
		if cabinetID != "" {
			break
		}
	}
	if cabinetID == "" {
		t.Fatal("seed did not provide a free locker")
	}
	locker, _, err := st.OccupyLocker(cabinetID, model.SizeSmall, 2.0)
	if err != nil {
		t.Fatalf("seed locker setup failed: %v", err)
	}
	lockerID := locker.ID
	engine := pricing.Default()
	svc := service.New(st, engine, scheduler.New(engine, st), clk)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 300; i++ {
			_ = svc.Stats.Dashboard()
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		occupied := true
		for i := 0; i < 300; i++ {
			if occupied {
				_, _, _ = st.ReleaseLocker(lockerID, 2.0)
				occupied = false
			} else {
				locker, _, err := st.OccupyLocker(cabinetID, model.SizeSmall, 2.0)
				if err != nil {
					t.Fatalf("re-occupy failed: %v", err)
				}
				lockerID = locker.ID
				occupied = true
			}
		}
	}()
	close(start)
	wg.Wait()
}
