package env

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/exp/slices"

	"github.com/go-leo/gonfig/format"
	"google.golang.org/protobuf/types/known/structpb"
)

// Resource represents a configuration resource loaded from environment variables
type Resource struct {
	// prefix to filter environment variables
	prefix string
	// formatter for parsing environment variables
	formatter format.Formatter
	// Atomic storage for the configuration data
	data atomic.Value
}

// Load retrieves and parses environment variables with the specified prefix
func (r *Resource) Load(ctx context.Context) (*structpb.Struct, error) {
	data, err := r.load(ctx)
	if err != nil {
		return nil, err
	}
	r.data.Store(data)
	return r.formatter.Parse(data)
}

// load collects and prepares environment variables data
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

// Watch monitors environment variables for changes
// notifyC: channel to receive new configuration when variables change
// errC: channel to receive errors during watching
// Returns a stop function to terminate the watcher and any initialization error
func (r *Resource) Watch(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(ctx context.Context) error, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	stopC := make(chan struct{})
	stop := func(ctx context.Context) error {
		close(stopC)
		return nil
	}
	// Start watching in a separate goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-stopC:
				return
			case <-time.After(time.Second):
				// Check for changes every second
				data, err := r.load(ctx)
				if err != nil {
					errC <- err
					continue
				}
				preData := r.data.Load()
				if preData != nil && bytes.Equal(preData.([]byte), data) {
					continue // Skip if no changes
				}
				newValue, err := r.formatter.Parse(data)
				if err != nil {
					errC <- err
					continue
				}
				notifyC <- newValue
				r.data.Store(data)
			}
		}
	}()
	return stop, nil
}

// New creates a new environment variable configuration resource
// prefix: The prefix used to filter environment variables (e.g., "APP_")
// Returns the Resource instance or error if initialization fails
func New(prefix string) (*Resource, error) {
	ext := "env"
	formatter, ok := format.GetFormatter(ext)
	if !ok {
		return nil, fmt.Errorf("config: not found formatter for %s", ext)
	}
	return &Resource{
		prefix:    prefix,
		formatter: formatter,
	}, nil
}
