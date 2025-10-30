package config

import (
	"context"

	"github.com/go-leo/gonfig/merge"
	"github.com/go-leo/gonfig/resource"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Load loads and merges configurations from multiple resources into a target protobuf message type.
//
// This is a generic function that supports loading any configuration type implementing proto.Message.
//
// Parameters:
//
//	ctx context.Context - Context for controlling the loading process
//	resources ...resource.Resource - Variadic parameter for multiple configuration resource loaders
//
// Returns:
//
//	Config - Successfully loaded and merged configuration object
//	error - Any error encountered during loading or processing
func Load[Config proto.Message](ctx context.Context, resources ...resource.Resource) (Config, error) {
	// 1. Sequentially load from all resources (return on first error)
	var config Config
	var values []*structpb.Struct
	for _, loader := range resources {
		value, err := loader.Load(ctx)
		if err != nil {
			return config, err
		}
		values = append(values, value)
	}

	// 2. Merge all loaded configurations using configured merger
	value := merge.GetMerger().Merge(values...)

	// 3. Convert merged structpb.Struct to JSON format
	data, err := value.MarshalJSON()
	if err != nil {
		return config, err
	}

	// 4. Unmarshal JSON into target protobuf message
	config = config.ProtoReflect().Type().New().Interface().(Config)
	if err := protojson.Unmarshal(data, config); err != nil {
		return config, err
	}
	return config, nil
}
