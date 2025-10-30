package consul

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/go-leo/gonfig/format"
	"github.com/go-leo/gonfig/format/env"
	"github.com/hashicorp/consul/api"
	_ "golang.org/x/exp/maps"
	_ "golang.org/x/net/http2"
	_ "golang.org/x/sys/unix"
	_ "golang.org/x/text"
	"google.golang.org/protobuf/types/known/structpb"
)

func consulFactory() (*api.Client, error) {
	return api.NewClient(api.DefaultConfig())
}

func TestResource_Load_Consul(t *testing.T) {
	client, err := consulFactory()
	if err != nil {
		t.Errorf("factory() error = %v", err)
		return
	}
	format.RegisterFormatter("env", env.Env{})
	key := "consul.env"

	_, err = client.KV().Put(&api.KVPair{
		Key:   key,
		Value: []byte("TEST_KEY=test_value"),
	}, nil)
	if err != nil {
		t.Errorf("Put() error = %v", err)
		return
	}

	defer func() {
		_, err = client.KV().Delete("consul", nil)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
			return
		}
	}()

	time.Sleep(time.Second)

	r, err := New(client, key)
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

func TestResource_Watch_Consul(t *testing.T) {
	client, err := consulFactory()
	if err != nil {
		t.Errorf("factory() error = %v", err)
		return
	}
	format.RegisterFormatter("env", env.Env{})
	key := "consul.env"

	_, err = client.KV().Put(&api.KVPair{
		Key:   key,
		Value: []byte("TEST_KEY=test_value"),
	}, nil)
	if err != nil {
		t.Errorf("PublishConfig() error = %v", err)
		return
	}

	defer func() {
		_, err = client.KV().Delete(key, nil)
		if err != nil {
			t.Errorf("PublishConfig() error = %v", err)
			return
		}
	}()

	time.Sleep(time.Second)

	r, err := New(client, key)
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

	go func() {
		time.Sleep(time.Second)
		_, err = client.KV().Put(&api.KVPair{
			Key:   key,
			Value: []byte("TEST_KEY_NEW=test_value_new" + time.Now().Format(time.RFC3339)),
		}, nil)
		if err != nil {
			t.Errorf("PublishConfig() error = %v", err)
			return
		}
	}()

	// Wait for the event
	select {
	case err := <-errC:
		if err != nil {
			t.Errorf("Watch() error = %v", err)
		}
	case data := <-notifyC:
		if data == nil {
			t.Error("Expected DataEvent with non-nil data")
		}
	case <-time.After(100 * time.Second):
		t.Error("No event received within the timeout")
	}

	stopFunc(ctx)

	go func() {
		time.Sleep(time.Second)
		_, err = client.KV().Put(&api.KVPair{
			Key:   key,
			Value: []byte("TEST_KEY_NEW=test_value_new" + time.Now().Format(time.RFC3339)),
		}, nil)
		if err != nil {
			t.Errorf("PublishConfig() error = %v", err)
			return
		}
	}()

	select {
	case err := <-errC:
		if err != nil {
			t.Errorf("Watch() error = %v", err)
		}
	case data := <-notifyC:
		if data != nil {
			t.Error("Did not expect to receive an event after stopping the watcher")
		}
	case <-time.After(100 * time.Millisecond):
		// Expected behavior
	}
}
