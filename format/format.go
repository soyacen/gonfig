package format

import (
	"google.golang.org/protobuf/types/known/structpb"
)

// Formatter interface defines the standard method for parsing configuration data
type Formatter interface {
	// Parse converts byte data into a protobuf Struct object
	//
	// Args:
	//   data ([]byte): Raw configuration data
	//
	// Returns:
	//   *structpb.Struct: Parsed structured data
	//   error: Error if parsing fails
	Parse(data []byte) (*structpb.Struct, error)
}
