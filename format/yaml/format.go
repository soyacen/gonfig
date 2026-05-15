// Package yaml provides a format.Formatter implementation for YAML content.
package yaml

import (
	"github.com/soyacen/gonfig/format"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// init registers the Yaml formatter with the global format registry.
func init() {
	format.RegisterFormatter("yaml", Yaml{})
	format.RegisterFormatter("yml", Yaml{})
}

// Yaml is a Formatter that parses YAML data.
type Yaml struct{}

// Parse parses YAML data into a protobuf Struct.
func (Yaml) Parse(data []byte) (*structpb.Struct, error) {
	v := make(map[string]any)
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return structpb.NewStruct(v)
}
