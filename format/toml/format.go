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

// Toml implements the Formatter interface for environment variables format.
type Toml struct{}

// Parse converts TOML-formatted byte data into a Protocol Buffer Struct object.
//
// Args:
//
//	data ([]byte): The TOML-formatted byte slice to be parsed
//
// Returns:
//
//	*structpb.Struct: A protobuf Struct object representing the parsed data
//	error: An error if parsing fails (e.g., invalid TOML format or type conversion issues)
func (Toml) Parse(data []byte) (*structpb.Struct, error) {
	v := make(map[string]any)
	if err := toml.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return structpb.NewStruct(v)
}
