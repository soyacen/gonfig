// Package file provides a resource.Resource implementation that loads
// configuration from local files and watches them for changes.
package file

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/soyacen/gonfig/format"
	"github.com/soyacen/gonfig/resource"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	resource.Register("file", &Factory{})
	resource.Register("", &Factory{})
}

var _ resource.Resource = (*Resource)(nil)

// Resource is a configuration source backed by a local file.
type Resource struct {
	filename  string
	formatter format.Formatter
	pre       atomic.Value
}

// Load reads the configuration file and parses it into a protobuf Struct.
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

// load reads the raw file content.
func (r *Resource) load(ctx context.Context) ([]byte, error) {
	return os.ReadFile(r.filename)
}

// Watch monitors the file for changes using fsnotify and invokes notifyFunc
// when the content changes.
func (r *Resource) Watch(ctx context.Context, notifyFunc resource.NotifyFunc, errFunc resource.ErrFunc) (resource.StopFunc, error) {
	// Validate notify function
	if notifyFunc == nil {
		return nil, fmt.Errorf("gonfig: notifyFunc is nil")
	}

	// Set default error handler if none provided
	if errFunc == nil {
		errFunc = func(err error) {
			slog.Error("gonfig: failed to watch file", slog.String("error", err.Error()))
		}
	}

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Initialize filesystem watcher
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Watch the directory containing the file
	if err := fsWatcher.Add(filepath.Dir(r.filename)); err != nil {
		return nil, err
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
		// Ensure watcher is closed when goroutine exits
		defer func() {
			if err := fsWatcher.Close(); err != nil {
				errFunc(err)
			}
		}()

		// Event loop
		for {
			select {
			case <-ctx.Done():
				// Context cancelled, exit goroutine
				errFunc(ctx.Err())
				return

			case <-stopC:
				// Stop signal received, exit goroutine
				return

			case err, ok := <-fsWatcher.Errors:
				// Error from filesystem watcher
				if !ok {
					return
				}
				errFunc(err)

			case event, ok := <-fsWatcher.Events:
				// File system event received
				if !ok {
					return
				}
				// Only process events for our specific file
				if filepath.Clean(event.Name) != filepath.Clean(r.filename) {
					continue
				}
				// Only react to write/create events
				if !event.Has(fsnotify.Write) && !event.Has(fsnotify.Create) {
					continue
				}
				// Handle file change
				data, err := r.load(ctx)
				if err != nil {
					errFunc(err)
					continue
				}
				// Compare with previous data to avoid unnecessary notifications
				preData := r.pre.Load()
				if preData != nil && bytes.Equal(preData.([]byte), data) {
					continue // Skip if content hasn't changed
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

// New creates a file Resource for the given filename.
func New(filename string) (*Resource, error) {
	// Extract file extension
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	if ext == "" {
		return nil, fmt.Errorf("config: file extension is empty")
	}

	// Find appropriate formatter for the file extension
	formatter, ok := format.GetFormatter(ext)
	if !ok {
		return nil, fmt.Errorf("config: not found formatter for %s", ext)
	}

	// Return new resource instance
	return &Resource{
		filename:  filename,
		formatter: formatter,
	}, nil
}

// Factory creates file Resources from DSN strings.
// DSN format: file:///path/to/file.ext or /path/to/file.ext
type Factory struct{}

// New creates a Resource from the given DSN.
func (Factory) New(ctx context.Context, dsn string) (resource.Resource, error) {
	filename := strings.TrimPrefix(dsn, "file://")
	return New(filename)
}
