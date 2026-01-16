package gonfig

import (
	"context"

	"github.com/soyacen/gonfig/resource"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func Load[Config proto.Message](ctx context.Context, resource resource.Resource) (Config, error) {
	var config Config
	value, err := resource.Load(ctx)
	if err != nil {
		return config, err
	}
	return convert[Config](value)
}

func Watch[Config proto.Message](ctx context.Context, resource resource.Resource, notifyFunc func(conf Config), errFunc resource.ErrFunc) (resource.StopFunc, error) {
	stopFunc, err := resource.Watch(
		ctx,
		func(value *structpb.Struct) {
			conf, err := convert[Config](value)
			if err != nil {
				panic(err)
			}
			notifyFunc(conf)
		},
		errFunc,
	)
	if err != nil {
		return nil, err
	}
	return stopFunc, nil
}

func convert[Config proto.Message](value *structpb.Struct) (Config, error) {
	var config Config
	data, err := value.MarshalJSON()
	if err != nil {
		return config, err
	}
	config = config.ProtoReflect().Type().New().Interface().(Config)
	if err := protojson.Unmarshal(data, config); err != nil {
		return config, err
	}
	return config, nil
}
