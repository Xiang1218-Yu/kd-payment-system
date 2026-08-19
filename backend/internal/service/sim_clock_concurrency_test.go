package service

import (
	"sync"
	"testing"
	"time"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/pricing"
	"kd-payment-system/backend/internal/scheduler"
	"kd-payment-system/backend/internal/store"
)

func TestSimClockAndPricingStaySynchronizedUnderLoad(t *testing.T) {
	storeClock := clock.NewManual(time.Date(2026, time.August, 19, 8, 0, 0, 0, time.UTC))
	simClock := clock.NewManual(time.Date(2026, time.August, 19, 9, 0, 0, 0, time.UTC))
	st := store.New(storeClock)
	store.Seed(st)
	engine := pricing.Default()
	svc := New(st, engine, scheduler.New(engine, st), simClock)

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 80; i++ {
			svc.Sim.Tick(time.Minute)
			svc.Sim.Reset()
		}
	}()
	for i := 0; i < 2; i++ {
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 160; j++ {
				if _, ok := svc.Pricing.Quote("cab-cbd-01", model.SizeMedium); !ok {
					t.Error("known cabinet disappeared")
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()

	st.SetClock(storeClock)
	st.SetClock(storeClock)
	state := svc.Sim.Reset()
	if got := st.Clock().Now(); !got.Equal(state.Now) {
		t.Fatalf("store clock = %s, want reset clock %s", got, state.Now)
	}
}
