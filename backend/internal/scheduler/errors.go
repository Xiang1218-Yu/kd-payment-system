package scheduler

import "errors"

// ErrCabinetNotFound is returned by Decide when the requested cabinet does not
// exist. Services map it to a 404.
var ErrCabinetNotFound = errors.New("cabinet not found")
