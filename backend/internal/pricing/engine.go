package pricing

import (
	"fmt"
	"math"
	"time"

	"kd-payment-system/backend/internal/model"
)

// Engine prices a locker rental. It holds only the policy; the time comes
// from the caller so the simulation endpoint can drive time-of-day changes.
type Engine struct {
	policy Policy
}

// New returns an Engine using the given policy.
func New(p Policy) *Engine { return &Engine{policy: p} }

// Default returns an Engine with DefaultPolicy.
func Default() *Engine { return New(DefaultPolicy()) }

// Quote computes the price for dropping a parcel of the given size into a
// cabinet, given its current stats and the reference time (now).
//
// The result includes a step-by-step breakdown so the UI can show the courier
// exactly which factors moved the price and by how much.
func (e *Engine) Quote(cabinetID string, size model.Size, stats model.LockerStats, now time.Time) model.PriceQuote {
	q := model.PriceQuote{
		CabinetID:    cabinetID,
		Size:         size,
		Base:         e.policy.Base[size],
		Utilization:  stats.Utilization,
		Available:    stats.Available,
	}

	// 1) Time-of-day factor.
	tb := e.timeBucket(now.Hour())
	q.TimeFactor = tb.Factor
	q.TimeOfDayLabel = tb.Label

	// 2) Utilization factor.
	ub := e.utilBand(stats.Utilization)
	q.UtilizationFactor = ub.Factor

	// 3) Scarcity factor (only when stock is genuinely scarce and the
	// cabinet actually carries this size).
	q.StockFactor = e.stockFactor(stats)

	// Final price is the base × the three multipliers, rounded to 0.01 yuan.
	q.Final = round2(q.Base * q.TimeFactor * q.UtilizationFactor * q.StockFactor)

	// Build an explainable breakdown: base → time → utilization → scarcity.
	q.Breakdown = append(q.Breakdown, model.PriceStep{
		Label:  fmt.Sprintf("基准价 %s", size),
		Factor: 0,
		Price:  q.Base,
	})
	q.Breakdown = append(q.Breakdown, model.PriceStep{
		Label:  fmt.Sprintf("%s时段 ×%.2f", tb.Label, tb.Factor),
		Factor: tb.Factor,
		Price:  round2(q.Base * tb.Factor),
	})
	q.Breakdown = append(q.Breakdown, model.PriceStep{
		Label:  fmt.Sprintf("%s(利用率%.0f%%) ×%.2f", ub.Label, stats.Utilization*100, ub.Factor),
		Factor: ub.Factor,
		Price:  round2(q.Base * tb.Factor * ub.Factor),
	})
	if q.StockFactor != 1.0 {
		q.Breakdown = append(q.Breakdown, model.PriceStep{
			Label:  fmt.Sprintf("稀缺加价(剩%d格) ×%.2f", stats.Available, q.StockFactor),
			Factor: q.StockFactor,
			Price:  q.Base * tb.Factor * ub.Factor * q.StockFactor,
		})
	}
	q.Breakdown = append(q.Breakdown, model.PriceStep{
		Label: "最终报价",
		Price: q.Final,
	})
	return q
}

// timeBucket returns the time-of-day bucket matching the given hour.
func (e *Engine) timeBucket(hour int) TimeBucket {
	for _, b := range e.policy.TimeBuckets {
		if hour >= b.FromHour && hour < b.ToHour {
			return b
		}
	}
	// Default to neutral if no bucket matched (shouldn't happen with a 0-24
	// covering policy, but keeps the function total).
	return TimeBucket{Factor: 1.0, Label: "平峰"}
}

// utilBand returns the utilization band whose range contains the ratio.
func (e *Engine) utilBand(util float64) UtilBand {
	for _, b := range e.policy.UtilBands {
		if util < b.Max {
			return b
		}
	}
	last := e.policy.UtilBands[len(e.policy.UtilBands)-1]
	return last
}

// stockFactor returns 1.0 unless the available count is at or below the
// scarcity threshold (and the cabinet actually has lockers of this size).
func (e *Engine) stockFactor(stats model.LockerStats) float64 {
	if stats.Total == 0 {
		return 1.0
	}
	if stats.Available <= e.policy.ScarcityThreshold {
		return e.policy.ScarcityFactor
	}
	return 1.0
}

// round2 rounds to two decimals, in yuan.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
