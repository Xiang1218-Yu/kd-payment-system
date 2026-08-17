package service

import (
	"errors"

	"kd-payment-system/backend/internal/clock"
	"kd-payment-system/backend/internal/model"
	"kd-payment-system/backend/internal/pricing"
	"kd-payment-system/backend/internal/scheduler"
	"kd-payment-system/backend/internal/store"
)

// DropoffService handles a courier's dropoff request: it asks the scheduler
// where to put the parcel, then occupies a locker in the recommended cabinet
// and records the parcel at the quoted price.
type DropoffService struct {
	store     *store.Store
	engine    *pricing.Engine
	scheduler *scheduler.Scheduler
	clk       clock.Clock
}

// DropoffRequest is the input for a dropoff.
type DropoffRequest struct {
	CabinetID string `json:"cabinetId"`
	Size      model.Size `json:"size"`
}

// DropoffResult is the outcome the courier sees.
type DropoffResult struct {
	Schedule     model.ScheduleResult `json:"schedule"`
	LockerID     string                `json:"lockerId"`
	ParcelID     string                `json:"parcelId"`
	PricePaid    float64               `json:"pricePaid"`
	Occupied     bool                  `json:"occupied"`
}

// Dropoff runs the full flow. It returns the schedule (so the courier sees the
// recommendation even if no locker was ultimately available) and the parcel
// info when a locker was occupied. When the recommended cabinet has no room,
// Occupied is false and the caller surfaces the no-capacity condition.
func (s *DropoffService) Dropoff(req DropoffRequest) (DropoffResult, error) {
	res, err := s.scheduler.Decide(scheduler.ScheduleRequest{
		CabinetID: req.CabinetID,
		Size:      req.Size,
	}, now(s.clk))
	if err != nil {
		return DropoffResult{Schedule: res}, err
	}

	quote := res.RecommendedQuote
	if quote == nil {
		return DropoffResult{Schedule: res}, errors.New("no quote produced")
	}

	locker, parcel, occErr := s.store.OccupyLocker(res.RecommendedCabinetID, req.Size, quote.Final)
	if occErr != nil {
		// The scheduler thought there was room but the occupy failed (e.g. a
		// race). Return the schedule so the courier can pick an alternative,
		// without erroring the whole request.
		return DropoffResult{Schedule: res, Occupied: false}, nil
	}
	return DropoffResult{
		Schedule:  res,
		LockerID:  locker.ID,
		ParcelID:  parcel.ID,
		PricePaid: quote.Final,
		Occupied:  true,
	}, nil
}
