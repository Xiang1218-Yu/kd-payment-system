package service

import (
	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/store"
)

// StatsService computes aggregate views over the store for the dashboard:
// region-level utilization, top crowded cabinets, and price distribution.
type StatsService struct {
	store *store.Store
}

// Dashboard is the payload for the overview page.
type Dashboard struct {
	Regions         []RegionView        `json:"regions"`
	TopCrowded      []CabinetUtilization `json:"topCrowded"`
	TopIdle         []CabinetUtilization `json:"topIdle"`
	HourlyVolume    [24]int             `json:"hourlyVolume"`
	AvgDwellMinutes float64             `json:"avgDwellMinutes"`
	TotalPickups    int                 `json:"totalPickups"`
}

// CabinetUtilization is one row in the crowded/idle rankings.
type CabinetUtilization struct {
	CabinetID   string  `json:"cabinetId"`
	CabinetName string  `json:"cabinetName"`
	RegionID    string  `json:"regionId"`
	RegionName  string  `json:"regionName"`
	Occupied    int     `json:"occupied"`
	Total       int     `json:"total"`
	Utilization float64 `json:"utilization"`
}

// Dashboard computes the overview. It iterates the topology and history once
// each; for the seed dataset (~15 cabinets, ~3500 records) this is cheap and
// fine to do per-request.
func (s *StatsService) Dashboard() Dashboard {
	var d Dashboard

	// Region aggregation.
	for _, r := range s.store.Regions() {
		rv := RegionView{Region: *r}
		for _, c := range s.store.CabinetsByRegion(r.ID) {
			rv.CabinetCount++
			for _, l := range c.Lockers {
				rv.LockerCount++
				if l.Occupied {
					rv.OccupiedCount++
				}
			}
		}
		if rv.LockerCount > 0 {
			rv.Utilization = float64(rv.OccupiedCount) / float64(rv.LockerCount)
		}
		d.Regions = append(d.Regions, rv)
	}

	// Cabinet-level ranking.
	cabs := s.store.AllCabinets()
	all := make([]CabinetUtilization, 0, len(cabs))
	regionName := func(id string) string {
		if r := s.store.Region(id); r != nil {
			return r.Name
		}
		return id
	}
	for _, c := range cabs {
		occ, total := 0, 0
		for _, l := range c.Lockers {
			total++
			if l.Occupied {
				occ++
			}
		}
		util := 0.0
		if total > 0 {
			util = float64(occ) / float64(total)
		}
		all = append(all, CabinetUtilization{
			CabinetID:   c.ID,
			CabinetName: c.Name,
			RegionID:    c.RegionID,
			RegionName:  regionName(c.RegionID),
			Occupied:    occ,
			Total:       total,
			Utilization: util,
		})
	}
	// Top crowded = highest utilization; Top idle = lowest.
	byUtilDesc(all)
	n := 5
	if n > len(all) {
		n = len(all)
	}
	d.TopCrowded = append(d.TopCrowded, all[:n]...)
	idle := make([]CabinetUtilization, len(all))
	copy(idle, all)
	byUtilAsc(idle)
	d.TopIdle = append(d.TopIdle, idle[:n]...)

	// History aggregates.
	hist := s.store.History()
	d.TotalPickups = len(hist)
	var dwellSum float64
	for _, h := range hist {
		d.HourlyVolume[h.PickupAt.Hour()]++
		dwellSum += h.DwellMinutes
	}
	if len(hist) > 0 {
		d.AvgDwellMinutes = dwellSum / float64(len(hist))
	}
	return d
}

// RegionsView returns the region list with aggregates (used by the region
// listing endpoint).
func (s *StatsService) RegionsView() []RegionView {
	var out []RegionView
	for _, r := range s.store.Regions() {
		rv := RegionView{Region: *r}
		for _, c := range s.store.CabinetsByRegion(r.ID) {
			rv.CabinetCount++
			for _, l := range c.Lockers {
				rv.LockerCount++
				if l.Occupied {
					rv.OccupiedCount++
				}
			}
		}
		if rv.LockerCount > 0 {
			rv.Utilization = float64(rv.OccupiedCount) / float64(rv.LockerCount)
		}
		out = append(out, rv)
	}
	return out
}

// CabinetDetail returns a cabinet with its lockers and per-size stats, for the
// cabinet detail view.
type CabinetDetail struct {
	model.Cabinet
	RegionName string                       `json:"regionName"`
	SizeStats  map[model.Size]model.LockerStats `json:"sizeStats"`
}

// CabinetDetail builds the detail view for one cabinet.
func (s *StatsService) CabinetDetail(id string) (CabinetDetail, bool) {
	c := s.store.Cabinet(id)
	if c == nil {
		return CabinetDetail{}, false
	}
	d := CabinetDetail{Cabinet: *c}
	if r := s.store.Region(c.RegionID); r != nil {
		d.RegionName = r.Name
	}
	d.SizeStats = make(map[model.Size]model.LockerStats)
	for _, sz := range model.AllSizes() {
		d.SizeStats[sz] = s.store.LockerStats(id, sz)
	}
	return d, true
}

// byUtilDesc sorts highest-utilization first.
func byUtilDesc(a []CabinetUtilization) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j].Utilization > a[j-1].Utilization; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

// byUtilAsc sorts lowest-utilization first.
func byUtilAsc(a []CabinetUtilization) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j].Utilization < a[j-1].Utilization; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}
