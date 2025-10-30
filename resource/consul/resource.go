package consul

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/go-leo/gonfig/format"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/hashicorp/go-hclog"

	"google.golang.org/protobuf/types/known/structpb"
)

// Resource represents a configuration resource stored in Consul KV store
type Resource struct {
	// client Consul API client
	client *api.Client
	// key path in Consul KV store
	key string
	// ext extension of the config (determines format)
	ext string
	// formatter for parsing config data
	formatter format.Formatter
	// data atomic storage for the configuration data
	data atomic.Value
}

// Load retrieves and parses the configuration from Consul
func (r *Resource) Load(ctx context.Context) (*structpb.Struct, error) {
	data, err := r.load(ctx)
	if err != nil {
		return nil, err
	}
	r.data.Store(data)
	return r.formatter.Parse(data)
}

// load is an internal helper to fetch raw data from Consul
func (r *Resource) load(ctx context.Context) ([]byte, error) {
	pair, _, err := r.client.KV().Get(r.key, new(api.QueryOptions).WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return pair.Value, nil
}

// Watch sets up a watcher for configuration changes in Consul
// notifyC: channel to receive new configuration when changed
// errC: channel to receive errors during watching
// Returns a stop function to terminate the watcher and any initialization error
func (r *Resource) Watch(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(ctx context.Context) error, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	params := map[string]any{
		"type": "key",
		"key":  r.key,
	}
	plan, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}
	plan.Handler = func(idx uint64, raw interface{}) {
		if raw == nil {
			return
		}
		pair, ok := raw.(*api.KVPair)
		if !ok {
			return
		}
		data := pair.Value
		preData := r.data.Load()
		if preData != nil && bytes.Equal(preData.([]byte), data) {
			return
		}
		newValue, err := r.formatter.Parse(data)
		if err != nil {
			errC <- err
			return
		}
		notifyC <- newValue
		r.data.Store(data)
	}
	go func() {
		_ = plan.RunWithClientAndHclog(
			r.client,
			&consuleLogger{
				Logger: hclog.NewNullLogger(),
				errC:   errC,
			})
	}()
	stopC := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-stopC:
		}
		plan.Stop()
	}()
	stop := func(ctx context.Context) error {
		close(stopC)
		return nil
	}
	return stop, nil
}

// consuleLogger is a custom logger that forwards errors to error channel
type consuleLogger struct {
	hclog.Logger
	errC chan<- error // Channel to forward errors
}

// Error implements the hclog.Logger interface and forwards errors to errC
func (l *consuleLogger) Error(msg string, args ...interface{}) {
	l.errC <- fmt.Errorf(msg, args...)
}

// New creates a new Consul configuration resource
// client: Consul API client
// key: Path to the configuration in Consul KV store
// Returns the Resource instance or error if initialization fails
func New(client *api.Client, key string) (*Resource, error) {
	ext := strings.TrimPrefix(filepath.Ext(key), ".")
	if ext == "" {
		return nil, fmt.Errorf("config: key extension is empty")
	}
	formatter, ok := format.GetFormatter(ext)
	if !ok {
		return nil, fmt.Errorf("config: not found formatter for %s", ext)
	}
	return &Resource{
		client:    client,
		key:       key,
		ext:       ext,
		formatter: formatter,
	}, nil
}
