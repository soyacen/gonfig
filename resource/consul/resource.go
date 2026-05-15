// Package consul provides a resource.Resource implementation backed by Consul KV.
package consul

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"
	"github.com/soyacen/gonfig/format"
	"github.com/soyacen/gonfig/resource"

	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	resource.Register("consul", Factory{})
}

var _ resource.Resource = (*Resource)(nil)

// Resource is a configuration source backed by Consul KV.
type Resource struct {
	client    *api.Client
	key       string
	formatter format.Formatter
	pre       atomic.Value
}

// Load fetches the configuration value from Consul KV and parses it into a
// protobuf Struct.
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

// load fetches raw data from Consul KV.
func (r *Resource) load(ctx context.Context) ([]byte, error) {
	pair, _, err := r.client.KV().Get(r.key, new(api.QueryOptions).WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, fmt.Errorf("gonfig: consul key %q not found", r.key)
	}
	return pair.Value, nil
}

// Watch monitors the Consul KV key for changes using Consul's watch plan and
// invokes notifyFunc when the value changes.
func (r *Resource) Watch(ctx context.Context, notifyFunc resource.NotifyFunc, errFunc resource.ErrFunc) (resource.StopFunc, error) {
	// Validate notify function
	if notifyFunc == nil {
		return nil, fmt.Errorf("gonfig: notifyFunc is nil")
	}

	// Set default error handler if none provided
	if errFunc == nil {
		errFunc = func(err error) {
			slog.Error("gonfig: failed to watch consul", slog.String("error", err.Error()))
		}
	}

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Prepare watch parameters for key monitoring
	params := map[string]any{
		"type": "key",
		"key":  r.key,
	}

	// Create watch plan
	plan, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}

	// Set up handler for watch events
	plan.Handler = func(idx uint64, raw interface{}) {
		// Validate the received data
		if raw == nil {
			errFunc(fmt.Errorf("gonfig: consul watch returned unexpected type %T", raw))
			return
		}

		// Type assert to KVPair
		pair, ok := raw.(*api.KVPair)
		if !ok {
			return
		}

		// Get the data value
		data := pair.Value

		// Compare with previous data to avoid unnecessary notifications
		preData := r.pre.Load()
		if preData != nil && bytes.Equal(preData.([]byte), data) {
			return
		}

		// Parse new configuration data
		newValue, err := r.formatter.Parse(data)
		if err != nil {
			errFunc(err)
			return
		}

		// Notify subscribers of the change
		notifyFunc(newValue)

		// Store new data for future comparisons
		r.pre.Store(data)
	}

	// Start watching in a separate goroutine
	go func() {
		// Create custom logger that forwards errors to errFunc
		logger := &consulLogger{
			Logger:  hclog.NewNullLogger(),
			errFunc: errFunc,
		}

		// Run the watch plan with the Consul client
		if err := plan.RunWithClientAndHclog(r.client, logger); err != nil {
			errFunc(err)
		}
	}()

	// Create stop function with sync.Once to ensure it's only called once
	stopC := make(chan struct{})
	var onceStop sync.Once
	stop := func(ctx context.Context) error {
		onceStop.Do(func() { close(stopC) })
		return nil
	}

	// Start a goroutine to handle context cancellation
	go func() {
		// Ensure plan is stop when goroutine exits
		defer plan.Stop()
		for {
			select {
			case <-ctx.Done():
				// Context cancelled, exit goroutine
				errFunc(ctx.Err())
				return
			case <-stopC:
				// Stop signal received, exit goroutine
				return
			}
		}
	}()

	return stop, nil
}

// consulLogger forwards hclog errors to the configured errFunc.
type consulLogger struct {
	hclog.Logger
	errFunc resource.ErrFunc
}

// Error implements hclog.Logger by forwarding the formatted message to errFunc.
func (l *consulLogger) Error(msg string, args ...interface{}) {
	buf := bytes.NewBufferString(msg)
	for i := 0; i < len(args); i += 2 {
		buf.WriteString(fmt.Sprintf(" %v=%v", args[i], args[i+1]))
	}
	l.errFunc(errors.New(buf.String()))
}

// New creates a Consul Resource for the given client and KV key.
func New(client *api.Client, key string) (*Resource, error) {
	// Extract key extension
	ext := strings.TrimPrefix(filepath.Ext(key), ".")
	if ext == "" {
		return nil, fmt.Errorf("config: key extension is empty")
	}

	// Find appropriate formatter for the key extension
	formatter, ok := format.GetFormatter(ext)
	if !ok {
		return nil, fmt.Errorf("config: not found formatter for %s", ext)
	}

	// Return new resource instance
	return &Resource{
		client:    client,
		key:       key,
		formatter: formatter,
	}, nil
}

// Factory creates consul Resources from DSN strings.
// DSN format: consul://[token@]ip:port/key.ext?wait=5s&partition=p1
type Factory struct{}

// New creates a Resource from the given DSN.
func (Factory) New(ctx context.Context, dsn string) (resource.Resource, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("consul: parse dsn failed: %w", err)
	}

	if u.Scheme != "consul" {
		return nil, fmt.Errorf("consul: invalid scheme: %s", u.Scheme)
	}

	// The entire path (without leading /) is treated as the key
	key := strings.TrimPrefix(u.Path, "/")
	if key == "" {
		return nil, fmt.Errorf("consul: key is empty")
	}

	config := api.DefaultConfig()
	if u.Host != "" {
		config.Address = u.Host
	}

	// Token can be in user info or query param
	if u.User != nil {
		config.Token = u.User.Username()
	}

	query := u.Query()
	if v := query.Get("token"); v != "" {
		config.Token = v
	}

	// Query parameters for Consul configuration
	if v := query.Get("datacenter"); v != "" {
		config.Datacenter = v
	}
	if v := query.Get("partition"); v != "" {
		config.Partition = v
	}
	if v := query.Get("namespace"); v != "" {
		config.Namespace = v
	}
	if v := query.Get("wait"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			config.WaitTime = d
		}
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("consul: create client failed: %w", err)
	}

	return New(client, key)
}
