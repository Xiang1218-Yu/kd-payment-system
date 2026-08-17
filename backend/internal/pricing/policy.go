// Package pricing computes dynamic per-locker rental prices. It is a pure
// function of (cabinet utilization, available stock, time of day) against a
// fixed policy — it has no knowledge of HTTP, persistence, or scheduling.
//
// finalPrice = base[size] × timeFactor × utilizationFactor × stockFactor
package pricing

import "kd-payment-system/backend/internal/model"

// Policy holds the tunable parameters of the pricing model. Keeping them in
// one struct makes the model easy to reason about and adjust.
type Policy struct {
	// Base price (yuan) per size, before multipliers.
	Base map[model.Size]float64

	// Time-of-day buckets: [fromHour, toHour) → multiplier. The first match
	// by hour wins. The label is shown in the UI.
	TimeBuckets []TimeBucket

	// Utilization bands: utilization falls into the first band whose Max
	// (exclusive) is greater than the ratio.
	UtilBands []UtilBand

	// StockFactor applies on top of utilization when available lockers of the
	// requested size drop below the scarcity threshold.
	ScarcityThreshold int     // when available <= this, apply ScarcityFactor
	ScarcityFactor    float64 // multiplier added (multiplicatively)
}

// TimeBucket is one pricing window in the day.
type TimeBucket struct {
	FromHour int     // inclusive, 0-23
	ToHour   int     // exclusive, 1-24
	Factor   float64
	Label    string // e.g. "晚高峰"
}

// UtilBand maps a utilization ratio range to a multiplier.
type UtilBand struct {
	Max    float64 // exclusive upper bound on utilization ratio
	Factor float64
	Label  string
}

// DefaultPolicy returns the policy used in production demos:
//   - late evening peak (17:00-21:00) carries a 1.40 surcharge;
//   - early morning (7:00-9:00) a mild 1.10;
//   - late night (22:00-6:00) a 0.60 discount;
//   - idle cabinets (<50% util) get a 0.80 promotion, crowded ones (>90%)
//     a 1.50 surcharge;
//   - when fewer than 3 lockers remain, a 1.25 scarcity surcharge stacks on.
func DefaultPolicy() Policy {
	return Policy{
		Base: map[model.Size]float64{
			model.SizeSmall:  2.0,
			model.SizeMedium: 3.0,
			model.SizeLarge:  5.0,
		},
		TimeBuckets: []TimeBucket{
			{FromHour: 0, ToHour: 6, Factor: 0.6, Label: "深夜优惠"},
			{FromHour: 6, ToHour: 7, Factor: 0.8, Label: "清晨"},
			{FromHour: 7, ToHour: 9, Factor: 1.1, Label: "早高峰"},
			{FromHour: 9, ToHour: 17, Factor: 1.0, Label: "平峰"},
			{FromHour: 17, ToHour: 21, Factor: 1.4, Label: "晚高峰"},
			{FromHour: 21, ToHour: 22, Factor: 1.1, Label: "夜间"},
			{FromHour: 22, ToHour: 24, Factor: 0.6, Label: "深夜优惠"},
		},
		UtilBands: []UtilBand{
			{Max: 0.5, Factor: 0.8, Label: "空闲促销"},
			{Max: 0.8, Factor: 1.0, Label: "正常"},
			{Max: 0.9, Factor: 1.2, Label: "繁忙加价"},
			{Max: 1.01, Factor: 1.5, Label: "爆满加价"},
		},
		ScarcityThreshold: 3,
		ScarcityFactor:    1.25,
	}
}
