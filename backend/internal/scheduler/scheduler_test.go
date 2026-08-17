package scheduler

import (
	"testing"
	"time"

	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/pricing"
)

// fakeProvider is a stand-in CabinetProvider for unit tests. It implements
// just enough topology to exercise Decide and bestNeighbor.
type fakeProvider struct {
	cabinets []*model.Cabinet
	stats    map[string]model.LockerStats
}

func (f fakeProvider) Cabinet(id string) *model.Cabinet {
	for _, c := range f.cabinets {
		if c.ID == id {
			return c
		}
	}
	return nil
}
func (f fakeProvider) CabinetsByRegion(regionID string) []*model.Cabinet {
	var out []*model.Cabinet
	for _, c := range f.cabinets {
		if c.RegionID == regionID {
			out = append(out, c)
		}
	}
	return out
}
func (f fakeProvider) LockerStats(id string, _ model.Size) model.LockerStats {
	return f.stats[id]
}

func newTestScheduler() (*Scheduler, *fakeProvider) {
	prov := &fakeProvider{
		cabinets: []*model.Cabinet{
			{ID: "full", RegionID: "r1", Name: "Full", Lat: 31.23, Lng: 121.47},
			{ID: "near-empty", RegionID: "r1", Name: "NearEmpty", Lat: 31.2308, Lng: 121.4708},
			{ID: "mid", RegionID: "r1", Name: "Mid", Lat: 31.2316, Lng: 121.4716},
			{ID: "other-region", RegionID: "r2", Name: "Other", Lat: 31.0, Lng: 121.0},
		},
		stats: map[string]model.LockerStats{
			"full":        {Total: 10, Occupied: 9, Available: 1, Utilization: 0.90},
			"near-empty":  {Total: 10, Occupied: 2, Available: 8, Utilization: 0.20},
			"mid":         {Total: 10, Occupied: 5, Available: 5, Utilization: 0.50},
			"other-region": {Total: 10, Occupied: 1, Available: 9, Utilization: 0.10},
		},
	}
	return New(pricing.Default(), prov), prov
}

func TestDecideRedirectsWhenTooFull(t *testing.T) {
	sch, _ := newTestScheduler()
	// "full" is at 90% util → above 0.85 threshold → must redirect.
	res, err := sch.Decide(ScheduleRequest{CabinetID: "full", Size: model.SizeMedium}, time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !res.IsRedirected {
		t.Fatalf("expected redirect, got confirmed: %+v", res)
	}
	// Best neighbor by (util asc, dist asc) → near-empty (20% util) wins over mid (50%).
	if res.RecommendedCabinetID != "near-empty" {
		t.Fatalf("recommended = %s, want near-empty", res.RecommendedCabinetID)
	}
	if res.RecommendedQuote == nil {
		t.Fatal("missing recommended quote")
	}
}

func TestDecideConfirmsWhenRoomAvailable(t *testing.T) {
	sch, _ := newTestScheduler()
	// "mid" at 50% util, has room → confirm.
	res, err := sch.Decide(ScheduleRequest{CabinetID: "mid", Size: model.SizeMedium}, time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.IsRedirected {
		t.Fatalf("expected confirm, got redirect: %+v", res)
	}
	if res.RecommendedCabinetID != "mid" {
		t.Fatalf("recommended = %s, want mid", res.RecommendedCabinetID)
	}
}

func TestDecideErrorOnUnknownCabinet(t *testing.T) {
	sch, _ := newTestScheduler()
	_, err := sch.Decide(ScheduleRequest{CabinetID: "nope", Size: model.SizeSmall}, time.Now())
	if err != ErrCabinetNotFound {
		t.Fatalf("err = %v, want ErrCabinetNotFound", err)
	}
}

func TestDecideDoesNotCrossRegions(t *testing.T) {
	sch, _ := newTestScheduler()
	res, _ := sch.Decide(ScheduleRequest{CabinetID: "full", Size: model.SizeMedium}, time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC))
	// Even though "other-region" is very empty, it's in r2 and must never be
	// recommended as a neighbor of an r1 cabinet.
	for _, alt := range res.Alternatives {
		if alt.CabinetID == "other-region" {
			t.Fatalf("scheduler leaked across regions: %s in alternatives", alt.CabinetID)
		}
	}
	if res.RecommendedCabinetID == "other-region" {
		t.Fatal("scheduler recommended a cabinet in another region")
	}
}

func TestHaversineZeroForSamePoint(t *testing.T) {
	d := HaversineMeters(31.23, 121.47, 31.23, 121.47)
	if d != 0 {
		t.Fatalf("distance = %.2f, want 0", d)
	}
	// ~0.0008 deg lat ≈ ~89m; sanity check it's in the right ballpark.
	d2 := HaversineMeters(31.23, 121.47, 31.2308, 121.4708)
	if d2 < 50 || d2 > 200 {
		t.Fatalf("distance = %.1fm, want 50-200m band", d2)
	}
}
