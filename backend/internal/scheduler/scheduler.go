// Package scheduler decides which cabinet a courier should use for a dropoff.
// When the requested cabinet is near capacity, it recommends a lower-
// utilization neighbor in the same region, balancing load across the fleet.
package scheduler

import (
	"fmt"
	"math"
	"sort"
	"time"

	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/pricing"
)

// Scheduler depends on the pricing engine (to quote alternatives) and a
// CabinetProvider (so it can query live topology and stats without depending
// on the store package directly — keeping the dependency arrow clean).
type Scheduler struct {
	pricing  *pricing.Engine
	provider CabinetProvider
}

// CabinetProvider is the narrow read interface the scheduler needs. The store
// implements it; the scheduler never imports store.
type CabinetProvider interface {
	Cabinet(id string) *model.Cabinet
	CabinetsByRegion(regionID string) []*model.Cabinet
	LockerStats(cabinetID string, size model.Size) model.LockerStats
}

// New returns a Scheduler.
func New(p *pricing.Engine, prov CabinetProvider) *Scheduler {
	return &Scheduler{pricing: p, provider: prov}
}

// RedirectThreshold is the utilization above which a cabinet is considered
// "too full" and dropoffs should be redirected elsewhere. Tunable.
const RedirectThreshold = 0.85

// ScheduleRequest is the input to Decide.
type ScheduleRequest struct {
	CabinetID string
	Size      model.Size
}

// MaxAlternatives is how many neighbor recommendations to return.
const MaxAlternatives = 3

// Decide produces a ScheduleResult for a dropoff request. It either confirms
// the requested cabinet (when it has capacity and is not too full) or
// redirects to the best nearby cabinet in the same region.
//
// now is passed through so the pricing quotes reflect the caller's notion of
// time (the simulation endpoint can advance the clock).
func (s *Scheduler) Decide(req ScheduleRequest, now time.Time) (model.ScheduleResult, error) {
	res := model.ScheduleResult{
		RequestedCabinetID: req.CabinetID,
		Size:               req.Size,
	}

	reqCab := s.provider.Cabinet(req.CabinetID)
	if reqCab == nil {
		return res, ErrCabinetNotFound
	}
	reqStats := s.provider.LockerStats(req.CabinetID, req.Size)
	reqQuote := s.pricing.Quote(req.CabinetID, req.Size, reqStats, now)

	// Case 1: the requested cabinet has room and is not too crowded → use it.
	if reqStats.Available > 0 && reqStats.Utilization < RedirectThreshold {
		res.RecommendedCabinetID = req.CabinetID
		res.RecommendedQuote = &reqQuote
		res.IsRedirected = false
		res.Reason = "目标柜机有可用格口，建议直接投递"
		res.Alternatives = s.collectAlternatives(reqCab, req.Size, now, nil)
		return res, nil
	}

	// Case 2: redirect to the best nearby cabinet.
	best, alts := s.bestNeighbor(reqCab, req.Size, now)
	if best == nil {
		// No neighbor has room either — fall back to the requested cabinet
		// and let the dropoff service surface the no-capacity error.
		res.RecommendedCabinetID = req.CabinetID
		res.RecommendedQuote = &reqQuote
		res.IsRedirected = false
		res.Reason = "目标柜机及邻近柜机均无可用格口，请稍后再试"
		return res, ErrNoCabinetCapacity
	}
	res.RecommendedCabinetID = best.CabinetID
	res.RecommendedQuote = best.Quote
	res.DistanceMeters = best.DistanceMeters
	res.IsRedirected = true
	res.Reason = redirectReason(reqCab, best)
	res.Alternatives = alts
	return res, nil
}

// neighborCand is an internal carrier for sorting candidates.
type neighborCand struct {
	cab      *model.Cabinet
	distance float64
	stats    model.LockerStats
}

// bestNeighbor finds the lowest-utilization, closest cabinet with capacity,
// excluding the requested cabinet. It returns the winner plus the full ranked
// alternative list (minus the winner), capped at MaxAlternatives.
func (s *Scheduler) bestNeighbor(req *model.Cabinet, size model.Size, now time.Time) (*model.ScheduleAlternative, []model.ScheduleAlternative) {
	neighbors := s.collectNeighbors(req, size)
	if len(neighbors) == 0 {
		return nil, nil
	}

	// Rank by utilization ascending, then distance ascending. Lower
	// utilization first balances load; distance breaks ties so the courier
	// walks less.
	sort.Slice(neighbors, func(i, j int) bool {
		if neighbors[i].stats.Utilization != neighbors[j].stats.Utilization {
			return neighbors[i].stats.Utilization < neighbors[j].stats.Utilization
		}
		return neighbors[i].distance < neighbors[j].distance
	})

	winner := neighbors[0]
	winnerQuote := s.pricing.Quote(winner.cab.ID, size, winner.stats, now)
	winnerAlt := &model.ScheduleAlternative{
		CabinetID:      winner.cab.ID,
		CabinetName:    winner.cab.Name,
		DistanceMeters: winner.distance,
		Utilization:    winner.stats.Utilization,
		Available:      winner.stats.Available,
		Quote:          &winnerQuote,
	}

	var alts []model.ScheduleAlternative
	for _, n := range neighbors[1:] {
		q := s.pricing.Quote(n.cab.ID, size, n.stats, now)
		alts = append(alts, model.ScheduleAlternative{
			CabinetID:      n.cab.ID,
			CabinetName:    n.cab.Name,
			DistanceMeters: n.distance,
			Utilization:    n.stats.Utilization,
			Available:      n.stats.Available,
			Quote:          &q,
		})
		if len(alts) >= MaxAlternatives {
			break
		}
	}
	return winnerAlt, alts
}

// collectAlternatives returns ranked alternatives for the "confirm" case,
// where the requested cabinet is itself excluded from the list.
func (s *Scheduler) collectAlternatives(req *model.Cabinet, size model.Size, now time.Time, _ interface{}) []model.ScheduleAlternative {
	neighbors := s.collectNeighbors(req, size)
	sort.Slice(neighbors, func(i, j int) bool {
		if neighbors[i].stats.Utilization != neighbors[j].stats.Utilization {
			return neighbors[i].stats.Utilization < neighbors[j].stats.Utilization
		}
		return neighbors[i].distance < neighbors[j].distance
	})
	out := make([]model.ScheduleAlternative, 0, MaxAlternatives)
	for _, n := range neighbors {
		q := s.pricing.Quote(n.cab.ID, size, n.stats, now)
		out = append(out, model.ScheduleAlternative{
			CabinetID:      n.cab.ID,
			CabinetName:    n.cab.Name,
			DistanceMeters: n.distance,
			Utilization:    n.stats.Utilization,
			Available:      n.stats.Available,
			Quote:          &q,
		})
		if len(out) >= MaxAlternatives {
			break
		}
	}
	return out
}

// collectNeighbors gathers same-region cabinets (excluding the requested one)
// that actually have a free locker of the requested size, with their
// haversine distance attached.
func (s *Scheduler) collectNeighbors(req *model.Cabinet, size model.Size) []neighborCand {
	peers := s.provider.CabinetsByRegion(req.RegionID)
	out := make([]neighborCand, 0, len(peers))
	for _, c := range peers {
		if c.ID == req.ID {
			continue
		}
		st := s.provider.LockerStats(c.ID, size)
		if st.Available <= 0 {
			continue
		}
		out = append(out, neighborCand{
			cab:      c,
			distance: HaversineMeters(req.Lat, req.Lng, c.Lat, c.Lng),
			stats:    st,
		})
	}
	return out
}

// redirectReason writes a short, courier-facing explanation of the redirect.
func redirectReason(req *model.Cabinet, best *model.ScheduleAlternative) string {
	if best == nil {
		return "未找到可用邻近柜机"
	}
	if best.DistanceMeters < 1 {
		return "目标柜机接近满载，已为您调度至同区域邻近柜机"
	}
	return fmt.Sprintf("目标柜机接近满载，已为您调度至 %s（约%d米）", best.CabinetName, int(best.DistanceMeters))
}

// HaversineMeters returns the great-circle distance in meters between two
// lat/lng points. Used by the scheduler to rank neighbors.
func HaversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371000.0 // earth radius, meters
	la1 := lat1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180
	dla := (lat2 - lat1) * math.Pi / 180
	dlo := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dla/2)*math.Sin(dla/2) +
		math.Cos(la1)*math.Cos(la2)*math.Sin(dlo/2)*math.Sin(dlo/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}
