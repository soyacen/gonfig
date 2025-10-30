package file

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/go-leo/gonfig/format"
	"google.golang.org/protobuf/types/known/structpb"
)

// Resource represents a configuration resource loaded from a file
type Resource struct {
	// filename path to the configuration file
	filename string
	// ext File extension (determines format parser)
	ext string
	// formatter for parsing file content
	formatter format.Formatter
	// data atomic storage for configuration data
	data atomic.Value
}

// Load reads and parses the configuration file
func (r *Resource) Load(ctx context.Context) (*structpb.Struct, error) {
	data, err := r.load(ctx)
	if err != nil {
		return nil, err
	}
	r.data.Store(data)
	return r.formatter.Parse(data)
}

// load is an internal helper to read raw file content
func (r *Resource) load(ctx context.Context) ([]byte, error) {
	return os.ReadFile(r.filename)
}

// Watch monitors the file for changes and notifies subscribers
// notifyC: channel to receive parsed configuration when file changes
// errC: channel to receive errors during watching
// Returns a stop function to terminate the watcher and any initialization error
func (r *Resource) Watch(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(ctx context.Context) error, error) {
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

	stopC := make(chan struct{})
	stop := func(ctx context.Context) error {
		close(stopC)
		return nil
	}

	// Start watching in a separate goroutine
	go func() {
		defer func() {
			if err := fsWatcher.Close(); err != nil {
				errC <- err
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-stopC:
				return
			case event, ok := <-fsWatcher.Events:
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
					errC <- err
					continue
				}
				preData := r.data.Load()
				if preData != nil && bytes.Equal(preData.([]byte), data) {
					continue // Skip if content hasn't changed
				}
				newValue, err := r.formatter.Parse(data)
				if err != nil {
					errC <- err
					continue
				}
				notifyC <- newValue
				r.data.Store(data)

			case err, ok := <-fsWatcher.Errors:
				if !ok {
					return
				}
				errC <- err
			}
		}
	}()

	return stop, nil
}

// New creates a new file-based configuration resource
// filename: Path to the configuration file
// Returns the Resource instance or error if initialization fails
func New(filename string) (*Resource, error) {
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	if ext == "" {
		return nil, fmt.Errorf("config: file extension is empty")
	}
	formatter, ok := format.GetFormatter(ext)
	if !ok {
		return nil, fmt.Errorf("config: not found formatter for %s", ext)
	}
	return &Resource{
		filename:  filename,
		ext:       ext,
		formatter: formatter,
	}, nil
}
