package json

import (
	"github.com/go-leo/gonfig/format"
	"google.golang.org/protobuf/types/known/structpb"
)

// init registers the Json formatter with the global format registry.
func init() {
	format.RegisterFormatter("json", Json{})
}

// Json implements the Formatter interface for environment variables format.
type Json struct{}

// Parse method converts JSON data into a structpb.Struct object.
// Parameters:
//
// Args:
//
// data []byte: JSON content as a byte slice to be parsed.
//
// Returns:
//
// *structpb.Struct: Pointer to the parsed structure.
// error: Error encountered during parsing, nil if successful.
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
