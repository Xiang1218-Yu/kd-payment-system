package scheduler

import "errors"

var (
	// ErrCabinetNotFound is returned by Decide when the requested cabinet does not
	// exist. Services map it to a 404.
	ErrCabinetNotFound = errors.New("cabinet not found")
	// ErrNoCabinetCapacity indicates that no same-region cabinet can accept
	// the requested parcel size.
	ErrNoCabinetCapacity = errors.New("no cabinet has capacity for requested size")
)
