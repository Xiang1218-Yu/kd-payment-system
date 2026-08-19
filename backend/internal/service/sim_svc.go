package service

import (
	"time"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/store"
)

// shanghai is the timezone the pricing buckets are calibrated for; reset
// re-pins the clock to "now" in this zone so time-of-day pricing stays
// correct regardless of the host/container TZ.
var shanghai = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}()

// SimService exposes the simulated clock so a demo can advance time and watch
// time-of-day pricing change. In production this endpoint would be disabled;
// here it is the whole point of demonstrating dynamic pricing.
type SimService struct {
	store *store.Store
	clk   clock.Clock
}

// SimState is the current simulated-clock status.
type SimState struct {
	Now time.Time `json:"now"`
}

// State returns the current clock time.
func (s *SimService) State() SimState { return SimState{Now: s.clk.Now()} }

// Tick advances the simulated clock by the given duration (e.g. "1h", "30m")
// and returns the new state. If parse fails it advances by one hour.
func (s *SimService) Tick(d time.Duration) SimState {
	if m, ok := s.clk.(*clock.Manual); ok {
		m.Advance(d)
		// Keep the store clock in sync (same instance in practice, but be
		// explicit in case the store holds a different one).
		s.store.SetClock(m)
	}
	return s.State()
}

// Reset snaps the clock back to the current wall time (in the pricing
// timezone) so the demo restarts from "now". The clock stays Manual, so Tick
// continues to work afterwards.
func (s *SimService) Reset() SimState {
	if m, ok := s.clk.(*clock.Manual); ok {
		m.Set(time.Now().In(shanghai))
		s.store.SetClock(m)
	}
	return s.State()
}
