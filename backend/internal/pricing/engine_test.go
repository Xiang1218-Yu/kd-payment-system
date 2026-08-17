package pricing

import (
	"testing"
	"time"

	"kd-payment-system/backend/internal/model"
)

func TestQuoteLatePeakSurcharges(t *testing.T) {
	e := Default()
	stats := model.LockerStats{Total: 10, Occupied: 8, Available: 2, Utilization: 0.8}

	// 09:00 → 平峰 (1.0)
	morning := e.Quote("cab", model.SizeMedium, stats, atHour(9))
	if morning.TimeFactor != 1.0 {
		t.Fatalf("morning timeFactor = %.2f, want 1.0", morning.TimeFactor)
	}

	// 19:00 → 晚高峰 (1.4), final should be larger than morning's.
	evening := e.Quote("cab", model.SizeMedium, stats, atHour(19))
	if evening.TimeFactor != 1.4 {
		t.Fatalf("evening timeFactor = %.2f, want 1.4", evening.TimeFactor)
	}
	if evening.Final <= morning.Final {
		t.Fatalf("evening final %.2f not greater than morning %.2f", evening.Final, morning.Final)
	}

	// 03:00 → 深夜优惠 (0.6), should be cheaper than morning.
	night := e.Quote("cab", model.SizeMedium, stats, atHour(3))
	if night.TimeFactor != 0.6 {
		t.Fatalf("night timeFactor = %.2f, want 0.6", night.TimeFactor)
	}
	if night.Final >= morning.Final {
		t.Fatalf("night final %.2f not less than morning %.2f", night.Final, morning.Final)
	}
}

func TestQuoteUtilizationBands(t *testing.T) {
	e := Default()
	noon := atHour(12) // 平峰, timeFactor 1.0

	// Idle (<50%) → 0.8 promotion.
	idle := e.Quote("cab", model.SizeSmall, model.LockerStats{Total: 10, Occupied: 3, Available: 7, Utilization: 0.3}, noon)
	if idle.UtilizationFactor != 0.8 {
		t.Fatalf("idle utilFactor = %.2f, want 0.8", idle.UtilizationFactor)
	}
	// Crowded (>90%) → 1.5.
	crowded := e.Quote("cab", model.SizeSmall, model.LockerStats{Total: 10, Occupied: 9, Available: 1, Utilization: 0.95}, noon)
	if crowded.UtilizationFactor != 1.5 {
		t.Fatalf("crowded utilFactor = %.2f, want 1.5", crowded.UtilizationFactor)
	}
	if crowded.Final <= idle.Final {
		t.Fatalf("crowded final %.2f not greater than idle %.2f", crowded.Final, idle.Final)
	}
}

func TestQuoteScarcityFactor(t *testing.T) {
	e := Default()
	noon := atHour(12)
	// Available=2 ≤ threshold(3) → scarcity 1.25 stacks on.
	scarce := e.Quote("cab", model.SizeLarge, model.LockerStats{Total: 6, Occupied: 4, Available: 2, Utilization: 0.67}, noon)
	if scarce.StockFactor != 1.25 {
		t.Fatalf("scarce stockFactor = %.2f, want 1.25", scarce.StockFactor)
	}
	// Available=5 > threshold → no scarcity.
	ample := e.Quote("cab", model.SizeLarge, model.LockerStats{Total: 6, Occupied: 1, Available: 5, Utilization: 0.17}, noon)
	if ample.StockFactor != 1.0 {
		t.Fatalf("ample stockFactor = %.2f, want 1.0", ample.StockFactor)
	}
}

func TestQuoteBreakdownEndsAtFinal(t *testing.T) {
	e := Default()
	q := e.Quote("cab", model.SizeMedium, model.LockerStats{Total: 10, Occupied: 8, Available: 2, Utilization: 0.8}, atHour(19))
	if len(q.Breakdown) < 2 {
		t.Fatalf("breakdown too short: %+v", q.Breakdown)
	}
	last := q.Breakdown[len(q.Breakdown)-1]
	if last.Price != q.Final {
		t.Fatalf("last breakdown price %.2f != final %.2f", last.Price, q.Final)
	}
}

func atHour(h int) time.Time {
	return time.Date(2026, 8, 17, h, 0, 0, 0, time.UTC)
}
