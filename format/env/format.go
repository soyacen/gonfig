// Package env provides a format.Formatter implementation for environment
// variable files (KEY=VALUE pairs).
package env

import (
	"github.com/joho/godotenv"
	"github.com/soyacen/gonfig/format"
	"google.golang.org/protobuf/types/known/structpb"
)

// init registers the Env formatter with the global format registry.
func init() {
	format.RegisterFormatter("env", Env{})
}

// Env is a Formatter that parses environment variable files.
type Env struct{}

// Parse parses environment variable data into a protobuf Struct.
// The input is expected to be KEY=VALUE lines separated by newlines.
func (Env) Parse(data []byte) (*structpb.Struct, error) {
	m, err := godotenv.UnmarshalBytes(data)
	if err != nil {
		return nil, err
	}
	v := make(map[string]any)
	for key, value := range m {
		v[key] = value
	}
	return structpb.NewStruct(v)
}
