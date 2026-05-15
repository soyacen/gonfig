package format

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestRegisterFormatter(t *testing.T) {
	t.Run("register successfully", func(t *testing.T) {
		var f Formatter = &testFormatter{}
		RegisterFormatter("testfmt", f)
		got, ok := GetFormatter("testfmt")
		if !ok {
			t.Fatal("expected formatter to be registered")
		}
		if got != f {
			t.Error("expected same formatter instance")
		}
	})

	t.Run("nil formatter panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for nil formatter")
			}
		}()
		RegisterFormatter("nilfmt", nil)
	})

	t.Run("duplicate registration panics", func(t *testing.T) {
		RegisterFormatter("dupfmt", &testFormatter{})
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for duplicate registration")
			}
		}()
		RegisterFormatter("dupfmt", &testFormatter{})
	})
}

func TestGetFormatter(t *testing.T) {
	RegisterFormatter("getfmt", &testFormatter{})

	tests := []struct {
		name   string
		ext    string
		wantOk bool
	}{
		{"registered extension", "getfmt", true},
		{"unregistered extension", "notfound", false},
		{"case insensitive", "GETFMT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := GetFormatter(tt.ext)
			if ok != tt.wantOk {
				t.Errorf("GetFormatter(%q) ok = %v, want %v", tt.ext, ok, tt.wantOk)
			}
		})
	}
}

type testFormatter struct{}

func (testFormatter) Parse(data []byte) (*structpb.Struct, error) {
	return nil, nil
}
