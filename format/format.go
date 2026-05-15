package format

import (
	"google.golang.org/protobuf/types/known/structpb"
)

// Formatter parses raw configuration bytes into a protobuf Struct.
type Formatter interface {
	// Parse converts raw configuration data into a protobuf Struct.
	Parse(data []byte) (*structpb.Struct, error)
}
