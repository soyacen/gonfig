package env

import (
	"bytes"
	"fmt"

	"github.com/soyacen/gonfig/format"
	"google.golang.org/protobuf/types/known/structpb"
)

// init registers the Env formatter with the global format registry.
func init() {
	format.RegisterFormatter("env", Env{})
}

// Env implements the Formatter interface for environment variables format.
type Env struct{}

// Parse converts environment variables format data into a protobuf Struct.
// The input data is expected to be a sequence of KEY=VALUE lines separated by newlines.
//
// Args:
//
//	data ([]byte) - Raw byte slice containing key-value pairs in KEY=VALUE format
//
// Returns:
// - *structpb.Struct: Parsed structured data with string values
// - error: Error if parsing fails
func (Env) Parse(data []byte) (*structpb.Struct, error) {
	s := bytes.Split(data, []byte("\n"))
	m := make(map[string]any, len(s))

	for _, v := range s {
		pair := bytes.SplitN(v, []byte("="), 2)
		if len(pair) != 2 {
			return nil, fmt.Errorf("invalid line: %s", v)
		}
		m[string(pair[0])] = string(pair[1])
	}
	return structpb.NewStruct(m)
}
