package resource

import (
	"context"
	"testing"
)

func TestRegister(t *testing.T) {
	t.Run("register successfully", func(t *testing.T) {
		var f Factory = &testFactory{}
		Register("testscheme", f)
		got, ok := Get("testscheme")
		if !ok {
			t.Fatal("expected factory to be registered")
		}
		if got != f {
			t.Error("expected same factory instance")
		}
	})

	t.Run("nil factory panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for nil factory")
			}
		}()
		Register("nilscheme", nil)
	})

	t.Run("duplicate registration panics", func(t *testing.T) {
		Register("dupscheme", &testFactory{})
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for duplicate registration")
			}
		}()
		Register("dupscheme", &testFactory{})
	})
}

func TestGet(t *testing.T) {
	Register("getscheme", &testFactory{})

	tests := []struct {
		name   string
		scheme string
		wantOk bool
	}{
		{"registered scheme", "getscheme", true},
		{"unregistered scheme", "notfound", false},
		{"case insensitive", "GETSCHEME", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := Get(tt.scheme)
			if ok != tt.wantOk {
				t.Errorf("Get(%q) ok = %v, want %v", tt.scheme, ok, tt.wantOk)
			}
		})
	}
}

type testFactory struct{}

func (testFactory) New(ctx context.Context, dsn string) (Resource, error) {
	return nil, nil
}
