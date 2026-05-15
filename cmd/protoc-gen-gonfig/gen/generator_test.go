package config

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
)

func TestNewGenerator(t *testing.T) {
	file := &protogen.File{}
	gen := NewGenerator(nil, file)
	if gen == nil {
		t.Fatal("expected non-nil Generator")
	}
	if gen.Plugin != nil {
		t.Error("expected nil Plugin")
	}
	if gen.File != file {
		t.Error("expected File to be set")
	}
}

func TestConfig(t *testing.T) {
	gen := &Generator{}
	msg := &protogen.Message{
		GoIdent: protogen.GoIdent{GoName: "AppConfig"},
	}
	if got := gen.Config(msg); got != "AppConfig" {
		t.Errorf("Config() = %q, want %q", got, "AppConfig")
	}
}

func TestGlobalConfig(t *testing.T) {
	gen := &Generator{}
	msg := &protogen.Message{
		GoIdent: protogen.GoIdent{GoName: "AppConfig"},
	}
	if got := gen.GlobalConfig(msg); got != "_AppConfig" {
		t.Errorf("GlobalConfig() = %q, want %q", got, "_AppConfig")
	}
}

func TestGetConfig(t *testing.T) {
	gen := &Generator{}
	msg := &protogen.Message{
		GoIdent: protogen.GoIdent{GoName: "AppConfig"},
	}
	if got := gen.GetConfig(msg); got != "GetAppConfig" {
		t.Errorf("GetConfig() = %q, want %q", got, "GetAppConfig")
	}
}

func TestSetConfig(t *testing.T) {
	gen := &Generator{}
	msg := &protogen.Message{
		GoIdent: protogen.GoIdent{GoName: "AppConfig"},
	}
	if got := gen.SetConfig(msg); got != "SetAppConfig" {
		t.Errorf("SetConfig() = %q, want %q", got, "SetAppConfig")
	}
}

func TestLoadConfig(t *testing.T) {
	gen := &Generator{}
	msg := &protogen.Message{
		GoIdent: protogen.GoIdent{GoName: "AppConfig"},
	}
	if got := gen.LoadConfig(msg); got != "LoadAppConfig" {
		t.Errorf("LoadConfig() = %q, want %q", got, "LoadAppConfig")
	}
}

func TestWatchConfig(t *testing.T) {
	gen := &Generator{}
	msg := &protogen.Message{
		GoIdent: protogen.GoIdent{GoName: "AppConfig"},
	}
	if got := gen.WatchConfig(msg); got != "WatchAppConfig" {
		t.Errorf("WatchConfig() = %q, want %q", got, "WatchAppConfig")
	}
}

func TestLoadAndWatchConfig(t *testing.T) {
	gen := &Generator{}
	msg := &protogen.Message{
		GoIdent: protogen.GoIdent{GoName: "AppConfig"},
	}
	if got := gen.LoadAndWatchConfig(msg); got != "LoadAndWatchAppConfig" {
		t.Errorf("LoadAndWatchConfig() = %q, want %q", got, "LoadAndWatchAppConfig")
	}
}

func TestGetField(t *testing.T) {
	gen := &Generator{}
	field := &protogen.Field{
		GoName: "Timeout",
	}
	if got := gen.GetField(field); got != "GetTimeout" {
		t.Errorf("GetField() = %q, want %q", got, "GetTimeout")
	}
}

func TestEnabledMessage(t *testing.T) {
	tests := []struct {
		name     string
		messages []*protogen.Message
		want     []string
	}{
		{
			name: "all matching",
			messages: []*protogen.Message{
				{GoIdent: protogen.GoIdent{GoName: "Config"}},
				{GoIdent: protogen.GoIdent{GoName: "Conf"}},
				{GoIdent: protogen.GoIdent{GoName: "Configuration"}},
			},
			want: []string{"Config", "Conf", "Configuration"},
		},
		{
			name: "partial matching",
			messages: []*protogen.Message{
				{GoIdent: protogen.GoIdent{GoName: "Config"}},
				{GoIdent: protogen.GoIdent{GoName: "OtherMessage"}},
				{GoIdent: protogen.GoIdent{GoName: "Configuration"}},
			},
			want: []string{"Config", "Configuration"},
		},
		{
			name: "no matching",
			messages: []*protogen.Message{
				{GoIdent: protogen.GoIdent{GoName: "OtherMessage"}},
				{GoIdent: protogen.GoIdent{GoName: "Request"}},
			},
			want: nil,
		},
		{
			name:     "empty messages",
			messages: []*protogen.Message{},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &protogen.File{
				Messages: tt.messages,
			}
			gen := NewGenerator(nil, file)
			got := gen.EnabledMessage()
			if len(got) != len(tt.want) {
				t.Fatalf("EnabledMessage() returned %d messages, want %d", len(got), len(tt.want))
			}
			for i, msg := range got {
				if msg.GoIdent.GoName != tt.want[i] {
					t.Errorf("message[%d].GoName = %q, want %q", i, msg.GoIdent.GoName, tt.want[i])
				}
			}
		})
	}
}
