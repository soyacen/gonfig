package config

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-leo/gonfig/test"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockResource implements resource.Resource for testing
type mockResource struct {
	watchFunc func(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(context.Context) error, error)
	loadFunc  func(ctx context.Context) (*structpb.Struct, error)
}

func (m *mockResource) Load(ctx context.Context) (*structpb.Struct, error) {
	return m.loadFunc(ctx)
}

func (m *mockResource) Watch(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(context.Context) error, error) {
	return m.watchFunc(ctx, notifyC, errC)
}

func TestWatch(t *testing.T) {
	t.Run("SuccessWatch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		v, _ := structpb.NewStruct(map[string]any{"field1": "value1"})

		// Setup test channels
		notifyC := make(chan *test.Config, 1)
		errC := make(chan error, 1)

		// Create mock resource
		mockRes := &mockResource{
			watchFunc: func(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(context.Context) error, error) {
				<-time.After(time.Second)
				notifyC <- v
				return func(ctx context.Context) error { return nil }, nil
			},
			loadFunc: func(ctx context.Context) (*structpb.Struct, error) {
				return v, nil
			},
		}

		// Call Watch
		stop, err := Watch[*test.Config](ctx, notifyC, errC, mockRes)
		if err != nil {
			t.Fatalf("Watch failed: %v", err)
		}
		defer stop(ctx)

		// Test notification
		select {
		case <-notifyC:
			// Success
		case <-time.After(2 * time.Second):
			t.Error("Timeout waiting for notification")
		}
	})

	t.Run("WatchError", func(t *testing.T) {
		ctx := context.Background()
		notifyC := make(chan *test.Config, 1)
		errC := make(chan error, 1)

		expectedErr := errors.New("watch error")
		mockRes := &mockResource{
			watchFunc: func(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(context.Context) error, error) {
				return func(ctx context.Context) error { return nil }, expectedErr
			},
		}

		_, err := Watch[*test.Config](ctx, notifyC, errC, mockRes)
		if err == nil || !errors.Is(err, expectedErr) {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("ContextCancel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		notifyC := make(chan *test.Config, 1)
		errC := make(chan error, 1)

		mockRes := &mockResource{
			watchFunc: func(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(context.Context) error, error) {
				return func(ctx context.Context) error { return nil }, nil
			},
		}

		_, err := Watch[*test.Config](ctx, notifyC, errC, mockRes)
		if err != nil {
			t.Fatalf("Watch failed: %v", err)
		}

		// Cancel context and verify behavior
		cancel()
		time.Sleep(100 * time.Millisecond) // Allow goroutines to exit
	})

	t.Run("StopAllResources", func(t *testing.T) {
		ctx := context.Background()
		notifyC := make(chan *test.Config, 1)
		errC := make(chan error, 1)

		stopCalled := false
		mockRes := &mockResource{
			watchFunc: func(ctx context.Context, notifyC chan<- *structpb.Struct, errC chan<- error) (func(context.Context) error, error) {
				return func(ctx context.Context) error {
					stopCalled = true
					return nil
				}, nil
			},
		}

		stop, err := Watch[*test.Config](ctx, notifyC, errC, mockRes)
		if err != nil {
			t.Fatalf("Watch failed: %v", err)
		}

		if err := stop(ctx); err != nil {
			t.Errorf("Stop failed: %v", err)
		}

		if !stopCalled {
			t.Error("Expected resource stop to be called")
		}
	})
}

func TestAppendSendChannel(t *testing.T) {
	inCh := make(chan int)
	outCh := appendSendChannel([]<-chan int{}, inCh)

	if len(outCh) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(outCh))
	}
}

func TestFanIn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)

	out := fanIn(ctx, ch1, ch2)

	ch1 <- 1
	ch2 <- 2

	results := make(map[int]bool)
	for i := 0; i < 2; i++ {
		select {
		case v := <-out:
			results[v] = true
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for fan-in results")
		}
	}

	if !results[1] || !results[2] {
		t.Error("Did not receive all expected values")
	}
}
