package nacos

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/go-leo/gonfig/format"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/protobuf/types/known/structpb"
)

// Resource represents a configuration resource in Nacos server
type Resource struct {
	// client Nacos config client
	client config_client.IConfigClient
	// group Configuration group in Nacos
	group string
	// dataId Configuration data ID in Nacos
	dataId string
	// extension of the configuration
	ext string
	// formatter for parsing configuration
	formatter format.Formatter
	// data atomic storage for configuration data
	data atomic.Value
}

// Load retrieves configuration from Nacos server and parses it into structpb.Struct
func (r *Resource) Load(ctx context.Context) (*structpb.Struct, error) {
	data, err := r.load(ctx)
	if err != nil {
		return nil, err
	}
	r.data.Store(data)
	return r.formatter.Parse(data)
}

// load is the internal method to get raw configuration data from Nacos
func (r *Resource) load(ctx context.Context) ([]byte, error) {
	content, err := r.client.GetConfig(vo.ConfigParam{
		Group:  r.group,
		DataId: r.dataId,
	})
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

// Watch monitors configuration changes in Nacos and notifies through channels
// Returns a stop function to cancel the watch and any initialization error
func (r *Resource) Watch(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(ctx context.Context) error, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	err := r.client.ListenConfig(vo.ConfigParam{
		Group:  r.group,
		DataId: r.dataId,
		OnChange: func(_, _, _, value string) {
			data := []byte(value)
			preData := r.data.Load()
			// Skip if configuration hasn't changed
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
		},
	})
	if err != nil {
		return nil, err
	}
	stopC := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
		case <-stopC:
		}
		err := r.client.CancelListenConfig(vo.ConfigParam{
			Group:  r.group,
			DataId: r.dataId,
		})
		if err != nil {
			errC <- err
			return
		}
	}()
	stop := func(ctx context.Context) error {
		close(stopC)
		return nil
	}
	return stop, nil
}

// New creates a new Nacos configuration resource
// Validates the dataId extension and finds appropriate formatter
func New(client config_client.IConfigClient, group string, dataId string) (*Resource, error) {
	ext := strings.TrimPrefix(filepath.Ext(dataId), ".")
	if ext == "" {
		return nil, fmt.Errorf("config: dataId extension is empty")
	}
	formatter, ok := format.GetFormatter(ext)
	if !ok {
		return nil, fmt.Errorf("config: not found formatter for %s", ext)
	}
	return &Resource{
		client:    client,
		group:     group,
		dataId:    dataId,
		ext:       ext,
		formatter: formatter,
	}, nil
}
