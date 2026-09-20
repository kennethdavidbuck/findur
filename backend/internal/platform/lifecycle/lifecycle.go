// Package lifecycle owns the ordered process drain sequence.
package lifecycle

import "context"

type readinessSetter interface {
	SetReady(bool)
}

type serverShutdown interface {
	Shutdown(context.Context) error
}

// Drain fails readiness, drains HTTP, and only then closes PostgreSQL.
func Drain(ctx context.Context, readiness readinessSetter, server serverShutdown, closeDatabase func()) error {
	readiness.SetReady(false)
	err := server.Shutdown(ctx)
	closeDatabase()
	return err
}
