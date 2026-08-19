package service_test

import (
	"errors"
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

func TestConcurrentPickupSingleCommit(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC))
	st := store.New(clk)
	store.Seed(st)
	var cabinetID string
	for _, cab := range st.AllCabinets() {
		for _, locker := range cab.Lockers {
			if !locker.Occupied && locker.Size == model.SizeMedium {
				cabinetID = cab.ID
				break
			}
		}
		if cabinetID != "" {
			break
		}
	}
	if cabinetID == "" {
		t.Fatal("seed did not provide a free medium locker")
	}
	locker, _, err := st.OccupyLocker(cabinetID, model.SizeMedium, 2.0)
	if err != nil {
		t.Fatalf("seed locker setup failed: %v", err)
	}
	lockerID := locker.ID

	engine := pricing.Default()
	svc := service.New(st, engine, scheduler.New(engine, st), clk).Pickup
	initialHistory := len(st.History())
	start := make(chan struct{})
	results := make(chan error, 64)
	var wg sync.WaitGroup
	for i := 0; i < cap(results); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := svc.Pickup(lockerID)
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	successes := 0
	empties := 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		if errors.Is(err, store.ErrLockerEmpty) {
			empties++
			continue
		}
		t.Fatalf("unexpected pickup error: %v", err)
	}
	if successes != 1 || empties != cap(results)-1 {
		t.Fatalf("successes=%d empties=%d, want one commit and remaining empty", successes, empties)
	}
	if got := len(st.History()) - initialHistory; got != 1 {
		t.Fatalf("new history entries=%d, want 1", got)
	}
}
