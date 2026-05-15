// Package json provides a format.Formatter implementation for JSON content.
package json

import (
	"github.com/soyacen/gonfig/format"
	"google.golang.org/protobuf/types/known/structpb"
)

// init registers the Json formatter with the global format registry.
func init() {
	format.RegisterFormatter("json", Json{})
}

// Json is a Formatter that parses JSON data.
type Json struct{}

// Parse parses JSON data into a protobuf Struct.
func (Json) Parse(data []byte) (*structpb.Struct, error) {
	value, err := structpb.NewStruct(map[string]any{})
	if err != nil {
		return nil, err
	}
	if err := value.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	return value, nil
}
