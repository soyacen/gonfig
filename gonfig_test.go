package gonfig

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/soyacen/gonfig/resource"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockResource struct {
	loadValue *structpb.Struct
	loadErr   error
	stopFunc  resource.StopFunc
	watchErr  error
}

func (m *mockResource) Load(ctx context.Context) (*structpb.Struct, error) {
	return m.loadValue, m.loadErr
}

func (m *mockResource) Watch(ctx context.Context, notifyFunc resource.NotifyFunc, errFunc resource.ErrFunc) (resource.StopFunc, error) {
	return m.stopFunc, m.watchErr
}

type mockFactory struct {
	resource resource.Resource
	newErr   error
}

func (m *mockFactory) New(ctx context.Context, dsn string) (resource.Resource, error) {
	return m.resource, m.newErr
}

func TestLoad(t *testing.T) {
	resource.Register("mock", &mockFactory{
		resource: &mockResource{
			loadValue: mustNewStruct(map[string]any{"key": "value"}),
		},
	})

	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{"valid mock dsn", "mock://test", false},
		{"invalid dsn", "://invalid", true},
		{"unknown scheme", "unknown://test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := Load[*structpb.Struct](ctx, tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad_FactoryNewError(t *testing.T) {
	resource.Register("mockloadnewerr", &mockFactory{
		resource: nil,
		newErr:   fmt.Errorf("factory error"),
	})

	ctx := context.Background()
	_, err := Load[*structpb.Struct](ctx, "mockloadnewerr://config")
	if err == nil {
		t.Error("expected error when factory.New fails")
	}
}

func TestLoad_ResourceLoadError(t *testing.T) {
	resource.Register("mockloaderr", &mockFactory{
		resource: &mockResource{loadErr: fmt.Errorf("load error")},
	})

	ctx := context.Background()
	_, err := Load[*structpb.Struct](ctx, "mockloaderr://config")
	if err == nil {
		t.Error("expected error when resource.Load fails")
	}
}

func TestLoad_Success(t *testing.T) {
	ctx := context.Background()
	expected := mustNewStruct(map[string]any{"hello": "world"})

	f := &mockFactory{
		resource: &mockResource{loadValue: expected},
	}
	resource.Register("mockload", f)

	val, err := Load[*structpb.Struct](ctx, "mockload://config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val == nil {
		t.Fatal("expected non-nil value")
	}
	got := val.AsMap()
	want := expected.AsMap()
	if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stopCalled := false
	stop := func(c context.Context) error {
		stopCalled = true
		return nil
	}

	resource.Register("mockwatch", &mockFactory{
		resource: &mockResource{stopFunc: stop},
	})

	stopFunc, err := Watch[*structpb.Struct](ctx, "mockwatch://config", func(conf *structpb.Struct) {}, func(err error) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stopFunc == nil {
		t.Fatal("expected non-nil stopFunc")
	}

	if err := stopFunc(ctx); err != nil {
		t.Errorf("stopFunc error: %v", err)
	}
	if !stopCalled {
		t.Error("expected stop to be called")
	}
}

func TestWatch_FactoryNewError(t *testing.T) {
	resource.Register("mockwatchnewerr", &mockFactory{
		resource: nil,
		newErr:   fmt.Errorf("factory error"),
	})

	ctx := context.Background()
	_, err := Watch[*structpb.Struct](ctx, "mockwatchnewerr://config",
		func(conf *structpb.Struct) {},
		func(err error) {},
	)
	if err == nil {
		t.Error("expected error when factory.New fails")
	}
}

func TestWatch_ResourceWatchError(t *testing.T) {
	resource.Register("mockwatcherr", &mockFactory{
		resource: &mockResource{
			watchErr: fmt.Errorf("watch error"),
		},
	})

	ctx := context.Background()
	_, err := Watch[*structpb.Struct](ctx, "mockwatcherr://config",
		func(conf *structpb.Struct) {},
		func(err error) {},
	)
	if err == nil {
		t.Error("expected error when resource.Watch fails")
	}
}

func TestConvert(t *testing.T) {
	tests := []struct {
		name    string
		value   *structpb.Struct
		wantErr bool
	}{
		{"valid struct", mustNewStruct(map[string]any{"key": "value"}), false},
		{"empty struct", mustNewStruct(map[string]any{}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convert[*structpb.Struct](tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got == nil {
				t.Fatal("expected non-nil result")
			}
		})
	}
}

func TestConvert_UnmarshalError(t *testing.T) {
	badValue := mustNewStruct(map[string]any{"key": "value"})

	_, err := convert[*structpb.ListValue](badValue)
	if err == nil {
		t.Error("expected error when unmarshal fails")
	}
}

func mustNewStruct(v map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(v)
	if err != nil {
		panic(err)
	}
	return s
}
