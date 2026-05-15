// Package resource defines the core Resource and Factory interfaces for
// configuration providers, along with callback types for change notifications.
package resource

import (
	"context"

	"google.golang.org/protobuf/types/known/structpb"
)

// NotifyFunc is called when a watched configuration changes.
type NotifyFunc func(value *structpb.Struct)

// ErrFunc is called when an error occurs during watching.
type ErrFunc func(err error)

// StopFunc stops watching and performs cleanup. It accepts a context for
// graceful shutdown and returns any cleanup error.
type StopFunc func(context.Context) error

// Resource is a configuration source that can load data synchronously and
// watch for changes.
type Resource interface {
	// Load retrieves the current configuration.
	Load(ctx context.Context) (*structpb.Struct, error)

	// Watch monitors the configuration for changes and invokes notifyFunc on
	// each update. errFunc receives errors that occur during monitoring.
	// Returns a stop function to cancel watching.
	Watch(ctx context.Context, notifyFunc NotifyFunc, errFunc ErrFunc) (StopFunc, error)
}

// Factory creates Resource instances from a DSN string.
type Factory interface {
	// New creates a new Resource from the given DSN.
	New(ctx context.Context, dsn string) (Resource, error)
}
