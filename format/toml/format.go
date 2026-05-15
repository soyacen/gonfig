// Package toml provides a format.Formatter implementation for TOML content.
package toml

import (
	"github.com/BurntSushi/toml"
	"github.com/soyacen/gonfig/format"
	"google.golang.org/protobuf/types/known/structpb"
)

// init registers the Toml formatter with the global format registry.
func init() {
	format.RegisterFormatter("toml", Toml{})
}

// Toml is a Formatter that parses TOML data.
type Toml struct{}

// Parse parses TOML data into a protobuf Struct.
func (Toml) Parse(data []byte) (*structpb.Struct, error) {
	v := make(map[string]any)
	if err := toml.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return structpb.NewStruct(v)
}
