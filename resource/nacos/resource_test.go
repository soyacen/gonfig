package nacos

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/go-leo/gonfig/format"
	"github.com/go-leo/gonfig/format/env"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	_ "golang.org/x/crypto/chacha20"
	_ "golang.org/x/net/http2"
	_ "golang.org/x/sync/singleflight"
	_ "golang.org/x/sys/unix"
	"google.golang.org/protobuf/types/known/structpb"
)

func nacosFactory() (config_client.IConfigClient, error) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}
	cc := *constant.NewClientConfig(
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
		constant.WithLogDir("/tmp/nacos.log"),
	)
	return clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
}

func TestResource_Load_Nacos(t *testing.T) {
	configClient, err := nacosFactory()
	if err != nil {
		t.Errorf("factory() error = %v", err)
		return
	}
	format.RegisterFormatter("env", env.Env{})

	dataId := "nacos.env"
	group := "test"
	_, err = configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: "TEST_KEY=test_value",
	})
	if err != nil {
		t.Errorf("PublishConfig() error = %v", err)
		return
	}

	defer func() {
		_, err = configClient.DeleteConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  group,
		})
		if err != nil {
			t.Errorf("DeleteConfig() error = %v", err)
			return
		}
	}()

	time.Sleep(time.Second)

	r, err := New(configClient, group, dataId)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}
	ctx := context.Background()
	content, err := r.Load(ctx)
	if err != nil {
		t.Errorf("Load() error = %v", err)
		return
	}

	if !reflect.DeepEqual(content.AsMap(), map[string]any{"TEST_KEY": "test_value"}) {
		t.Errorf("Load() data = %v, want data to contain 'TEST_KEY=test_value'", content.AsMap())
	}

	time.Sleep(time.Second)
}

func TestResource_Watch_Nacos(t *testing.T) {
	configClient, err := nacosFactory()
	if err != nil {
		t.Errorf("factory() error = %v", err)
		return
	}
	format.RegisterFormatter("env", env.Env{})

	dataId := "nacos.env"
	group := "test"
	_, err = configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: "TEST_KEY=test_value",
	})
	if err != nil {
		t.Errorf("PublishConfig() error = %v", err)
		return
	}

	defer func() {
		_, err = configClient.DeleteConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  group,
		})
		if err != nil {
			t.Errorf("DeleteConfig() error = %v", err)
			return
		}
	}()

	time.Sleep(time.Second)

	r, err := New(configClient, group, dataId)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}
	notifyC := make(chan *structpb.Struct, 1)
	errC := make(chan error, 1)
	// Start watching
	ctx := context.Background()
	stopFunc, err := r.Watch(ctx, notifyC, errC)
	if err != nil {
		t.Errorf("Watch() error = %v", err)
		return
	}

	// Give some time for the watcher to detect the change
	go func() {
		time.Sleep(time.Second)
		ok, err := configClient.PublishConfig(vo.ConfigParam{
			DataId:  dataId,
			Group:   group,
			Content: "TEST_KEY_NEW=test_value_new" + time.Now().Format(time.RFC3339),
		})
		if err != nil {
			t.Errorf("PublishConfig() error = %v", err)
			return
		}
		t.Log(ok)
	}()

	// Wait for the event
	select {
	case value := <-notifyC:
		if value == nil {
			t.Error("Expected DataEvent with non-nil data")
		}
	case <-time.After(100 * time.Second):
		t.Error("No event received within the timeout")
	}

	stopFunc(ctx)

	// Give some time for the watcher to detect the change
	go func() {
		time.Sleep(time.Second)
		ok, err := configClient.PublishConfig(vo.ConfigParam{
			DataId:  dataId,
			Group:   group,
			Content: "TEST_KEY_NEW=test_value_new" + time.Now().Format(time.RFC3339),
		})
		if err != nil {
			t.Errorf("PublishConfig() error = %v", err)
			return
		}
		t.Log(ok)
	}()

	select {
	case data := <-notifyC:
		if data != nil {
			t.Error("Did not expect to receive an event after stopping the watcher")
		}
	case <-time.After(2 * time.Millisecond):
	}
}
