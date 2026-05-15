// Package nacos provides a resource.Resource implementation backed by Nacos.
package nacos

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/soyacen/gonfig/format"
	"github.com/soyacen/gonfig/resource"
	"google.golang.org/protobuf/types/known/structpb"
)

func init() {
	resource.Register("nacos", Factory{})
}

var _ resource.Resource = (*Resource)(nil)

// Resource is a configuration source backed by Nacos.
type Resource struct {
	client    config_client.IConfigClient
	group     string
	dataId    string
	ext       string
	formatter format.Formatter
	pre       atomic.Value
}

// Load fetches the configuration from Nacos and parses it into a protobuf Struct.
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

// load fetches raw configuration content from Nacos.
func (r *Resource) load(ctx context.Context) ([]byte, error) {
	content, err := r.client.GetConfig(vo.ConfigParam{Group: r.group, DataId: r.dataId})
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

// Watch registers a Nacos listener for the configured group and dataId and
// invokes notifyFunc when the configuration changes.
func (r *Resource) Watch(ctx context.Context, notifyFunc resource.NotifyFunc, errFunc resource.ErrFunc) (resource.StopFunc, error) {
	// Validate notify function
	if notifyFunc == nil {
		return nil, fmt.Errorf("gonfig: notifyFunc is nil")
	}

	// Set default error handler if none provided
	if errFunc == nil {
		errFunc = func(err error) {
			slog.Error("gonfig: failed to watch nacos", slog.String("error", err.Error()))
		}
	}

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Set up handler for configuration change events
	onChange := func(_, _, _, value string) {
		data := []byte(value)
		preData := r.pre.Load()
		// Skip if configuration hasn't changed
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

	// Register listener with Nacos client
	if err := r.client.ListenConfig(vo.ConfigParam{Group: r.group, DataId: r.dataId, OnChange: onChange}); err != nil {
		return nil, err
	}

	// Create stop function with sync.Once to ensure it's only called once
	stopC := make(chan struct{})
	var onceStop sync.Once
	stop := func(ctx context.Context) error {
		onceStop.Do(func() { close(stopC) })
		return nil
	}

	// Start a goroutine to handle context cancellation and cleanup
	go func() {
		defer func() {
			// Cancel listener when goroutine exits
			if err := r.client.CancelListenConfig(vo.ConfigParam{Group: r.group, DataId: r.dataId}); err != nil {
				errFunc(err)
				return
			}
		}()
		select {
		case <-ctx.Done():
			// Context cancelled, report error
			errFunc(ctx.Err())
			return

		case <-stopC:
			// Stop signal received, exit goroutine
			return
		}
	}()

	return stop, nil
}

// New creates a Nacos Resource for the given client, group, and dataId.
func New(client config_client.IConfigClient, group string, dataId string) (*Resource, error) {
	// Extract dataId extension
	ext := strings.TrimPrefix(filepath.Ext(dataId), ".")
	if ext == "" {
		return nil, fmt.Errorf("config: dataId extension is empty")
	}

	// Find appropriate formatter for the dataId extension
	formatter, ok := format.GetFormatter(ext)
	if !ok {
		return nil, fmt.Errorf("config: not found formatter for %s", ext)
	}

	// Return new resource instance
	return &Resource{
		client:    client,
		group:     group,
		dataId:    dataId,
		ext:       ext,
		formatter: formatter,
	}, nil
}

// Factory creates nacos Resources from DSN strings.
// DSN format: nacos://username:password@ip:port/dataId.ext?namespace=ns&group=g
type Factory struct{}

// New creates a Resource from the given DSN.
func (Factory) New(ctx context.Context, dsn string) (resource.Resource, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("nacos: parse dsn failed: %w", err)
	}

	if u.Scheme != "nacos" {
		return nil, fmt.Errorf("nacos: invalid scheme: %s", u.Scheme)
	}

	// Extract dataId from path
	// path: /dataId.ext
	path := strings.TrimPrefix(u.Path, "/")
	if path == "" {
		return nil, fmt.Errorf("nacos: dataId is required in path")
	}
	dataId := path

	// Extract namespace and group from query parameters
	query := u.Query()
	namespace := query.Get("namespace")
	group := query.Get("group")

	// Set default values if not provided
	if namespace == "" {
		namespace = "public"
	}
	if group == "" {
		group = "DEFAULT_GROUP"
	}

	// Extract server config
	host := u.Hostname()
	portStr := u.Port()
	if portStr == "" {
		portStr = "8848"
	}
	port, err := strconv.ParseUint(portStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("nacos: invalid port: %s", portStr)
	}

	serverConfigs := []constant.ServerConfig{*constant.NewServerConfig(host, port)}

	// Extract client config
	clientConfig := constant.NewClientConfig()
	clientConfig.NamespaceId = namespace
	if u.User != nil {
		clientConfig.Username = u.User.Username()
		clientConfig.Password, _ = u.User.Password()
	}

	// Parse query parameters for additional configurations
	if v := query.Get("timeoutMs"); v != "" {
		if t, err := strconv.ParseUint(v, 10, 64); err == nil {
			clientConfig.TimeoutMs = t
		}
	}
	if v := query.Get("logDir"); v != "" {
		clientConfig.LogDir = v
	}
	if v := query.Get("cacheDir"); v != "" {
		clientConfig.CacheDir = v
	}
	if v := query.Get("logLevel"); v != "" {
		clientConfig.LogLevel = v
	}
	if v := query.Get("notLoadCacheAtStart"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			clientConfig.NotLoadCacheAtStart = b
		}
	}
	if v := query.Get("appName"); v != "" {
		clientConfig.AppName = v
	}
	if v := query.Get("logDir"); v != "" {
		clientConfig.LogDir = v
	} else {
		clientConfig.LogDir = "/tmp/nacos/log"
	}
	if v := query.Get("cacheDir"); v != "" {
		clientConfig.CacheDir = v
	} else {
		clientConfig.CacheDir = "/tmp/nacos/cache"
	}

	client, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, fmt.Errorf("nacos: create client failed: %w", err)
	}

	return New(client, group, dataId)
}
