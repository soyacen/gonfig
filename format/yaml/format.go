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

// Yaml implements the Formatter interface for environment variables format.
type Yaml struct{}

// Parse converts YAML-formatted byte data into a Protocol Buffer Struct object.
//
// Args:
//
//	data ([]byte): The YAML-formatted byte slice to be parsed
//
// Returns:
//
//	*structpb.Struct: A protobuf Struct object representing the parsed data
//	error: An error if parsing fails (e.g., invalid YAML format or type conversion issues)
func (Yaml) Parse(data []byte) (*structpb.Struct, error) {
	v := make(map[string]any)
	if err := yaml.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return structpb.NewStruct(v)
}
