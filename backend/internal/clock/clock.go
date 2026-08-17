// Package clock abstracts the source of "now" so the pricing engine's
// time-of-day logic can be tested deterministically and nudged by the
// simulation endpoint during demos. The production clock reads the real
// wall clock; the manual clock holds a value the simulator advances.
package clock

import "time"

// Clock returns the current time. Implementations must be safe for concurrent
// use.
type Clock interface {
	Now() time.Time
}

// Real returns the actual wall-clock time.
type Real struct{}

// Now returns time.Now.
func (Real) Now() time.Time { return time.Now() }

// Manual holds a time that callers can read and advance. It is the seam the
// simulation endpoint uses to demonstrate time-of-day pricing changes.
type Manual struct {
	t time.Time
}

// NewManual creates a Manual clock pinned to the given time.
func NewManual(t time.Time) *Manual { return &Manual{t: t} }

// Now returns the currently pinned time.
func (m *Manual) Now() time.Time { return m.t }

// Set replaces the pinned time.
func (m *Manual) Set(t time.Time) { m.t = t }

// Advance moves the pinned time forward by d.
func (m *Manual) Advance(d time.Duration) { m.t = m.t.Add(d) }
