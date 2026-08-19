package scheduler

import "errors"

// ErrCabinetNotFound is returned by Decide when the requested cabinet does not
// exist. Services map it to a 404.
var ErrCabinetNotFound = errors.New("cabinet not found")

// ErrNoCapacity is returned by Decide when the requested cabinet is full (or
// above the redirect threshold) and no same-region neighbor has a free locker
// of the requested size either — i.e. the whole region is out of capacity for
// this size. Services map it to a 409 so the caller can tell a genuine capacity
// exhaustion apart from a normal "not occupied" result. The returned
// ScheduleResult is still populated (Reason, Alternatives) so the courier can
// surface context alongside the error.
var ErrNoCapacity = errors.New("no locker capacity available in the requested region")
