// Package env provides environment variable-based implementation of the configuration resource interface
package env

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/exp/slices"

	"github.com/soyacen/gonfig/format"
	"github.com/soyacen/gonfig/resource"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	resource.Register("env", &Factory{})
}

var _ resource.Resource = (*Resource)(nil)

// Resource is a configuration source backed by environment variables.
type Resource struct {
	prefix    string
	interval  time.Duration
	formatter format.Formatter
	pre       atomic.Value
}

// Load collects environment variables matching the configured prefix, sorts
// them, and parses them into a protobuf Struct.
func (r *Resource) Load(ctx context.Context) (*structpb.Struct, error) {
	data, err := r.load(ctx)
	if err != nil {
		return nil, err
	}
	parsed, err := r.formatter.Parse(data)
	if err != nil {
		return nil, err
	}
	r.pre.Store(data)
	return parsed, nil
}

// load filters environment variables by prefix, sorts them, and joins them
// into a single byte slice.
func (r *Resource) load(ctx context.Context) ([]byte, error) {
	var environs [][]byte
	// Filter environment variables by prefix
	for _, environ := range os.Environ() {
		if strings.HasPrefix(environ, r.prefix) {
			environs = append(environs, []byte(environ))
		}
	}
	if len(environs) <= 0 {
		return nil, fmt.Errorf("config: no environment variables found with prefix %s", r.prefix)
	}
	// Sort and join variables for consistent output
	slices.SortFunc(environs, bytes.Compare)
	return bytes.Join(environs, []byte("\n")), nil
}

// Watch polls environment variables at the configured interval and invokes
// notifyFunc when changes are detected.
func (r *Resource) Watch(ctx context.Context, notifyFunc resource.NotifyFunc, errFunc resource.ErrFunc) (resource.StopFunc, error) {
	// Validate notify function
	if notifyFunc == nil {
		return nil, fmt.Errorf("gonfig: notifyFunc is nil")
	}

	// Set default error handler if none provided
	if errFunc == nil {
		errFunc = func(err error) {
			slog.Error("gonfig: failed to watch env", slog.String("error", err.Error()))
		}
	}

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Create stop function with sync.Once to ensure it's only called once
	stopC := make(chan struct{})
	var onceStop sync.Once
	stop := func(ctx context.Context) error {
		onceStop.Do(func() { close(stopC) })
		return nil
	}

	// Start watching in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Context cancelled, exit goroutine
				errFunc(ctx.Err())
				return

			case <-stopC:
				// Stop signal received, exit goroutine
				return

			case <-time.After(r.interval):
				// Check for changes at regular intervals
				data, err := r.load(ctx)
				if err != nil {
					errFunc(err)
					continue
				}
				// Compare with previous data to avoid unnecessary notifications
				preData := r.pre.Load()
				if preData != nil && bytes.Equal(preData.([]byte), data) {
					continue // Skip if no changes
				}
				// Parse new configuration data
				newValue, err := r.formatter.Parse(data)
				if err != nil {
					errFunc(err)
					continue
				}
				// Notify subscribers of the change
				notifyFunc(newValue)
				// Store new data for future comparisons
				r.pre.Store(data)
			}
		}
	}()

	return stop, nil
}

// New creates a Resource that monitors environment variables matching prefix.
// Interval controls the polling frequency; values <= 0 default to 5 seconds.
func New(prefix string, interval time.Duration) (*Resource, error) {
	ext := "env"
	// Find appropriate formatter for environment variables
	formatter, ok := format.GetFormatter(ext)
	if !ok {
		return nil, fmt.Errorf("config: not found formatter for %s", ext)
	}

	// Set default interval if not provided or invalid
	if interval <= 0 {
		interval = 5 * time.Second
	}

	// Return new resource instance
	return &Resource{
		prefix:    prefix,
		interval:  interval,
		formatter: formatter,
	}, nil
}

// Factory creates env Resources from DSN strings.
// DSN format: env://prefix?interval=5s
type Factory struct{}

// New creates a Resource from the given DSN.
func (Factory) New(ctx context.Context, dsn string) (resource.Resource, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("env: parse dsn failed: %w", err)
	}

	if u.Scheme != "env" {
		return nil, fmt.Errorf("env: invalid scheme: %s", u.Scheme)
	}

	// Prefix can be in the hostname or path
	prefix := u.Host
	if prefix == "" {
		prefix = strings.TrimPrefix(u.Path, "/")
	}

	interval := 5 * time.Second
	if v := u.Query().Get("interval"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	return New(prefix, interval)
}
