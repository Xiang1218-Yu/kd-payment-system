package model

// PriceQuote is the output of the pricing engine for a single locker size at
// a single cabinet. It carries the final price plus a full breakdown so the
// frontend can explain why a price is what it is.
type PriceQuote struct {
	CabinetID string `json:"cabinetId"`
	Size      Size   `json:"size"`

	// Base price for the size, before any dynamic multipliers.
	Base float64 `json:"base"`
	// Multipliers that combine multiplicatively into Final.
	TimeFactor         float64 `json:"timeFactor"`
	UtilizationFactor  float64 `json:"utilizationFactor"`
	StockFactor        float64 `json:"stockFactor"`
	// Final quoted price = Base × TimeFactor × UtilizationFactor × StockFactor.
	Final float64 `json:"final"`

	// Human-readable label for the current time-of-day bucket, e.g. "晚高峰".
	TimeOfDayLabel string `json:"timeOfDayLabel"`
	// Current utilization ratio of the cabinet (for the requested size),
	// echoed back so the caller need not re-fetch it.
	Utilization float64 `json:"utilization"`
	// Number of available lockers of this size right now.
	Available int `json:"available"`

	// Breakdown is the ordered list of factors shown in the UI, from base to
	// final. Each step shows the running price after applying that factor.
	Breakdown []PriceStep `json:"breakdown"`
}

// PriceStep is one factor in the price composition chain.
type PriceStep struct {
	Label  string  `json:"label"`  // e.g. "晚高峰时段系数 ×1.40"
	Factor float64 `json:"factor"` // 0 when this step only sets the base
	Price  float64 `json:"price"`  // running price after this step
}

// ScheduleResult is the output of the scheduling engine for a dropoff
// request. It tells the courier which cabinet to use and why.
type ScheduleResult struct {
	RequestedCabinetID string  `json:"requestedCabinetId"`
	Size               Size    `json:"size"`

	// Recommended is the cabinet the system actually suggests using. It may
	// equal RequestedCabinetID when the original cabinet has capacity.
	RecommendedCabinetID string      `json:"recommendedCabinetId"`
	RecommendedQuote     *PriceQuote `json:"recommendedQuote"`
	DistanceMeters       float64     `json:"distanceMeters"`
	IsRedirected         bool        `json:"isRedirected"`
	// Reason explains the scheduling decision in one short sentence.
	Reason string `json:"reason"`

	// Alternatives are other nearby cabinets the courier could pick instead,
	// ranked by (utilization asc, distance asc). Top 3.
	Alternatives []ScheduleAlternative `json:"alternatives"`
}

// ScheduleAlternative is one neighbor cabinet recommendation.
type ScheduleAlternative struct {
	CabinetID      string      `json:"cabinetId"`
	CabinetName    string      `json:"cabinetName"`
	DistanceMeters float64     `json:"distanceMeters"`
	Utilization    float64     `json:"utilization"`
	Available      int         `json:"available"`
	Quote          *PriceQuote `json:"quote"`
}
