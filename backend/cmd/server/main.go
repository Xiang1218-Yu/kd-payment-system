// Command server is the entry point for the locker pricing & scheduling
// system. It wires the clock, store, pricing engine, scheduler and services
// together, seeds the store, and serves the HTTP API + embedded frontend.
//
// Responsibilities here are intentionally limited to composition: every piece
// of behavior lives in its own package, so this file reads as a wiring
// diagram.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/handler"
	"kd-payment-system/backend/internal/pricing"
	"kd-payment-system/backend/internal/scheduler"
	"kd-payment-system/backend/internal/service"
	"kd-payment-system/backend/internal/store"
)

// shanghai is the timezone the pricing model is calibrated for: the time-of-
// day buckets (晚高峰 17:00-21:00 etc.) are Beijing time. Anchoring the
// clock here keeps the demo consistent regardless of the container's TZ.
var shanghai = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}()

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// Start the simulated clock pinned to Shanghai time, a few minutes in the
	// past so the first quote reflects "right now". The /api/sim/tick
	// endpoint advances it to demonstrate time-of-day pricing changes.
	now := time.Now().In(shanghai).Truncate(time.Second).Add(-time.Minute * 5)
	clk := clock.NewManual(now)

	st := store.New(clk)
	store.Seed(st)

	engine := pricing.Default()
	// *store.Store satisfies scheduler.CabinetProvider (Cabinet,
	// CabinetsByRegion, LockerStats), so no adapter is needed — the
	// scheduler depends on the interface, not the store package.
	sch := scheduler.New(engine, st)
	svc := service.New(st, engine, sch, clk)

	log.Printf("locker pricing & scheduling system listening on %s (clock: %s)", addr, clk.Now().Format("2006-01-02 15:04 MST"))
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.New(svc),
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
