package httpapi

import (
	"context"
	"sync/atomic"
	"time"
)

// Dependency is the minimum database contract required by readiness.
type Dependency interface {
	Ping(context.Context) error
}

// Readiness controls admission and verifies the database on each readiness probe.
type Readiness struct {
	accepting atomic.Bool
	database  Dependency
	timeout   time.Duration
}

// NewReadiness creates a bounded readiness gate for the database dependency.
func NewReadiness(database Dependency, timeout time.Duration) *Readiness {
	return &Readiness{database: database, timeout: timeout}
}

// SetReady changes whether the process admits readiness checks.
func (r *Readiness) SetReady(ready bool) {
	r.accepting.Store(ready)
}

// IsReady reports whether the process currently admits readiness checks.
func (r *Readiness) IsReady() bool {
	return r.accepting.Load()
}

// Check verifies admission and the database dependency within the configured timeout.
func (r *Readiness) Check(ctx context.Context) error {
	if !r.IsReady() {
		return errNotAccepting
	}

	checkCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	err := r.database.Ping(checkCtx)
	if !r.IsReady() {
		return errNotAccepting
	}
	return err
}
