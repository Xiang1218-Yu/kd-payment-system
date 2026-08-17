// Package service orchestrates use cases by composing the store, pricing
// engine and scheduler. It is the only layer allowed to cross those
// boundaries, so handler code never touches pricing/scheduling/store
// internals directly — keeping the HTTP layer thin and the dependency arrows
// pointing inward.
package service

import (
	"time"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/pricing"
	"kd-payment-system/backend/internal/scheduler"
	"kd-payment-system/backend/internal/store"
)

// Services bundles the use cases so the handler layer can depend on one
// aggregate. Each field is a small, focused service with one responsibility.
type Services struct {
	Pricing  *PricingService
	Dropoff  *DropoffService
	Pickup   *PickupService
	Stats    *StatsService
	Sim      *SimService
}

// New wires the services together from their dependencies.
func New(s *store.Store, p *pricing.Engine, sch *scheduler.Scheduler, clk clock.Clock) *Services {
	return &Services{
		Pricing: &PricingService{store: s, engine: p, clk: clk},
		Dropoff: &DropoffService{store: s, engine: p, scheduler: sch, clk: clk},
		Pickup:  &PickupService{store: s, clk: clk},
		Stats:   &StatsService{store: s},
		Sim:     &SimService{store: s, clk: clk},
	}
}

// now is a tiny helper shared by the services to read the injected clock.
func now(clk clock.Clock) time.Time { return clk.Now() }

// RegionView is the region as seen by the dashboard: topology plus aggregate
// utilization across all its cabinets.
type RegionView struct {
	model.Region
	CabinetCount  int     `json:"cabinetCount"`
	LockerCount   int     `json:"lockerCount"`
	OccupiedCount int     `json:"occupiedCount"`
	Utilization   float64 `json:"utilization"`
}
