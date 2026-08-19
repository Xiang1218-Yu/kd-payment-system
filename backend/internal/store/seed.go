package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/model"
)

// newID returns a short random ID with the given prefix. Random bytes keep IDs
// collision-free without a central counter.
func newID(prefix string) string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// rand.Read failing is exceptional; fall back to time-based uniqueness.
		return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(b)
}

// seedRegion is the static description used to materialize the world. Keeping
// the topology here (rather than inline in Seed) makes the demo data easy to
// tune.
type seedRegion struct {
	id, prefix, name, description string
	lat, lng                      float64
	// baseUtil drives how full each cabinet in the region is at startup,
	// simulating the "CBD is crowded, residential B is idle" imbalance.
	baseUtil float64
	// cabinetCount and the per-cabinet name pattern.
	cabinetCount  int
	cabinetPrefix string
}

var seedRegions = []seedRegion{
	{"region-cbd", "cbd", "中央商务区", "写字楼密集，工作日白天利用率极高", 31.2304, 121.4737, 0.92, 5, "CBD"},
	{"region-res-a", "resa", "居民区A", "成熟社区，取件集中在傍晚", 31.2150, 121.4500, 0.66, 6, "社区A"},
	{"region-res-b", "resb", "居民区B", "新交付社区，格口充裕、利用率低", 31.2050, 121.4350, 0.38, 4, "社区B"},
}

// Seed populates the store with a realistic snapshot: 3 regions, cabinets with
// mixed S/M/L lockers, partial occupancy matching each region's baseUtil, and
// 7 days of pickup history with an evening peak distribution.
//
// It uses the store's clock as the reference "now" for history timestamps so
// the demo always shows fresh recent activity.
func Seed(s *Store) {
	now := s.now()

	for _, sr := range seedRegions {
		s.regions[sr.id] = &model.Region{
			ID:          sr.id,
			Name:        sr.name,
			Lat:         sr.lat,
			Lng:         sr.lng,
			Description: sr.description,
		}
		for i := 1; i <= sr.cabinetCount; i++ {
			seedCabinet(s, sr, i, now)
		}
	}
	seedHistory(s, now)
}

// seedCabinet builds one cabinet with ~30 lockers (S/M/L mix), then occupies a
// fraction of them according to the region's baseUtil so the dashboard opens
// to an already-uneven world.
func seedCabinet(s *Store, sr seedRegion, idx int, now time.Time) {
	cid := fmt.Sprintf("cab-%s-%02d", sr.prefix, idx)
	// Spread cabinets in a small grid around the region center so the
	// scheduler has real distances to work with.
	lat := sr.lat + 0.0015*float64(idx%3-1) + 0.0007*float64(idx/3)
	lng := sr.lng + 0.0015*float64(idx%3) + 0.0007*float64(idx/4)

	c := &model.Cabinet{
		ID:       cid,
		RegionID: sr.id,
		Name:     fmt.Sprintf("%s%d号柜", sr.cabinetPrefix, idx),
		Lat:      lat,
		Lng:      lng,
		Address:  fmt.Sprintf("%s %d号", sr.name, idx*100),
	}
	// Locker mix: 12 S, 12 M, 6 L = 30 per cabinet.
	counts := map[model.Size]int{
		model.SizeSmall:  12,
		model.SizeMedium: 12,
		model.SizeLarge:  6,
	}
	for _, sz := range model.AllSizes() {
		for j := 0; j < counts[sz]; j++ {
			lid := fmt.Sprintf("%s-%s%d", cid, sz, j+1)
			l := &model.Locker{
				ID:        lid,
				CabinetID: cid,
				Size:      sz,
			}
			c.Lockers = append(c.Lockers, l)
			s.lockers[lid] = l
		}
	}
	s.cabinets[cid] = c

	// Occupy a fraction of each size to reach roughly baseUtil. We occupy
	// per-size so smaller sizes (more demand) fill first, which is what
	// makes the scheduler redirect S/M dropoffs to neighbors. Each cabinet's
	// effective utilization is jittered ±0.15 around the region baseUtil so
	// every region has both crowded and idle cabinets — making the scheduler's
	// "redirect to a less-loaded neighbor" visibly land on a truly idle one.
	jitter := (pseudoRand(idx, sr.prefix) - 0.5) * 0.30 // ±0.15
	cabinetUtil := sr.baseUtil + jitter
	if cabinetUtil < 0.05 {
		cabinetUtil = 0.05
	}
	if cabinetUtil > 0.98 {
		cabinetUtil = 0.98
	}
	for _, sz := range model.AllSizes() {
		total := counts[sz]
		occ := int(math.Round(float64(total) * cabinetUtil))
		if occ > total {
			occ = total
		}
		placed := 0
		for _, l := range c.Lockers {
			if l.Size != sz {
				continue
			}
			if placed >= occ {
				break
			}
			l.Occupied = true
			p := &model.Parcel{
				ID:           newID("p"),
				LockerID:     l.ID,
				CabinetID:    c.ID,
				RegionID:     c.RegionID,
				Size:         l.Size,
				DropoffAt:    now.Add(-time.Duration(30+placed*5) * time.Minute),
				DropoffPrice: 2.0 + pseudoRand(placed, l.ID)*3.0,
			}
			l.ParcelID = p.ID
			s.parcels[p.ID] = p
			placed++
		}
	}
}

// seedHistory generates 7 days of pickup records with an evening peak (17:00-
// 21:00) so the stats dashboard shows realistic dwell-time and volume
// patterns. Records are written directly to history (no live parcels).
func seedHistory(s *Store, now time.Time) {
	// hourlyWeights models relative pickup volume across a day; the evening
	// peak is the whole reason dynamic pricing exists.
	hourlyWeights := [24]float64{
		0.2, 0.1, 0.05, 0.05, 0.05, 0.1, // 0-5
		0.3, 0.6, 0.9, 1.0, 0.8, 0.7, // 6-11
		0.9, 0.7, 0.6, 0.6, 0.8, 1.6, // 12-17
		1.8, 1.5, 1.1, 0.8, 0.5, 0.3, // 18-23
	}
	var totalWeight float64
	for _, w := range hourlyWeights {
		totalWeight += w
	}

	cabinets := s.AllCabinets()
	// ~240 records per cabinet over 7 days → enough signal without bloat.
	const recordsPerCabinet = 240

	for _, c := range cabinets {
		for n := 0; n < recordsPerCabinet; n++ {
			// pick an hour weighted by hourlyWeights
			target := pseudoRand(n, c.ID) * totalWeight
			hour := 0
			acc := hourlyWeights[0]
			for hour < 23 && acc < target {
				hour++
				acc += hourlyWeights[hour]
			}
			dayOffset := int(pseudoRand(n+1000, c.ID) * 7)
			if dayOffset > 6 {
				dayOffset = 6
			}
			dropoff := now.Add(-time.Duration(dayOffset) * 24 * time.Hour).
				Add(-time.Duration(hour) * time.Hour).
				Add(-time.Duration(int(pseudoRand(n+2000, c.ID)*60)) * time.Minute)
			// Dwell 30min–24h: evening dropoffs sit longer.
			dwell := time.Duration(30+pseudoRand(n+3000, c.ID)*1380) * time.Minute
			pickup := dropoff.Add(dwell)
			if pickup.After(now) {
				pickup = now
			}
			sz := model.AllSizes()[int(pseudoRand(n+4000, c.ID)*3)%3]
			s.history = append(s.history, model.PickupRecord{
				ID:           newID("r"),
				LockerID:     fmt.Sprintf("%s-%s1", c.ID, sz),
				CabinetID:    c.ID,
				RegionID:     c.RegionID,
				Size:         sz,
				DropoffAt:    dropoff,
				PickupAt:     pickup,
				DwellMinutes: dwell.Minutes(),
				PricePaid:    2.0 + pseudoRand(n+5000, c.ID)*3.0,
			})
		}
	}
}

// pseudoRand returns a deterministic float in [0,1) derived from an integer
// and a salt string. Determinism makes the seed reproducible across runs
// without a global RNG seed; it is good enough for volume distributions.
func pseudoRand(n int, salt string) float64 {
	h := uint64(0xcbf29ce484222325)
	for _, b := range []byte(salt) {
		h ^= uint64(b)
		h *= 0x100000001b3
	}
	h ^= uint64(n) * 0x9e3779b97f4a7c15
	h ^= h >> 31
	// map to [0,1)
	return float64(h%100000) / 100000.0
}

// SetClock swaps the store's clock. Used by the simulation endpoint to advance
// time so time-of-day pricing changes become visible during a demo. It takes
// the write lock because the clock field is read on every mutation.
func (s *Store) SetClock(c clock.Clock) {
	s.mu.Lock()
	s.clk = c
	s.mu.Unlock()
}

// Clock returns the store's current clock, primarily so services can read
// "now" through the same source as the store.
func (s *Store) Clock() clock.Clock {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clk
}
