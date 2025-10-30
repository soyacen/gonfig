package config

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-leo/gonfig/resource"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Watch continuously monitors configuration resources for changes and sends updates
// to the provided channels. It handles:
// - Merging notifications from multiple sources
// - Debouncing changes and periodically reloading configurations
// - Proper cleanup via returned stop function
//
// Parameters:
//
//	ctx      - Context for cancellation and timeout control
//	notifyC  - Channel to receive configuration updates (protobuf messages)
//	errC     - Channel to receive any errors during watching
//	resources - Variadic list of configuration resources to watch
//
// Returns:
//
//	stop function - Call to clean up all watchers (returns combined errors if any)
//	error        - Initial error if watching failed to start
func Watch[Config proto.Message](ctx context.Context, notifyC chan<- Config, errC chan<- error, resources ...resource.Resource) (func(context.Context) error, error) {
	// Channels from individual resource watchers
	var notifyCs []chan *structpb.Struct
	// Stop functions for each watcher
	var stops []func(context.Context) error

	// Combined stop function that stops all individual watchers
	stop := func(context.Context) error {
		var errs []error
		for _, stop := range stops {
			if err := stop(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}

	// Start watching each resource
	for _, watcher := range resources {
		notifyC := make(chan *structpb.Struct, cap(notifyC))
		stop, err := watcher.Watch(ctx, notifyC, errC)
		if err != nil {
			return nil, errors.Join(err, stop(ctx))
		}
		notifyCs = append(notifyCs, notifyC)
		stops = append(stops, stop)
	}

	// Merge notifications from all watchers into single channel
	mergedC := fanIn(
		ctx,
		appendSendChannel(
			make([]<-chan *structpb.Struct, 0, len(notifyCs)),
			notifyCs...,
		)...,
	)

	// Monitor for changes and reload configurations
	go func() {
		ticker := time.NewTicker(time.Second)
		var changed bool
		for {
			select {
			case <-mergedC: // Received change notification
				changed = true
			case <-ticker.C: // Periodic check
				if !changed {
					changed = false
					ticker.Reset(time.Second)
					continue
				}
				// Load and send new configuration
				config, err := Load[Config](ctx, resources...)
				if err != nil {
					errC <- err
					continue
				}
				notifyC <- config
				changed = false
				ticker.Reset(time.Second)
			}
		}
	}()

	return stop, nil
}

// appendSendChannel converts send-only channels to receive-only channels
// for use in fan-in pattern. This is a helper function for Watch.
func appendSendChannel[T any](c []<-chan T, channels ...chan T) []<-chan T {
	for _, ch := range channels {
		c = append(c, ch)
	}
	return c
}

// fanIn combines multiple input channels into a single output channel.
// It runs until all input channels are closed or context is cancelled.
// This is a helper function for Watch.
func fanIn[T any](ctx context.Context, ins ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	for _, ch := range ins {
		wg.Add(1)
		go func(ch <-chan T) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-ch:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case out <- v:
					}
				}
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
