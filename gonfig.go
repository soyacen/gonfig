package gonfig

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/soyacen/gonfig/resource"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func Load[Config proto.Message](ctx context.Context, dsn string) (Config, error) {
	var config Config
	u, err := url.Parse(dsn)
	if err != nil {
		return config, fmt.Errorf("gonfig: parse dsn failed: %w", err)
	}
	factory, ok := resource.Get(u.Scheme)
	if !ok {
		return config, errors.New("gonfig: resource factory not found for dsn " + dsn)
	}
	resource, err := factory.New(ctx, dsn)
	if err != nil {
		return config, err
	}
	value, err := resource.Load(ctx)
	if err != nil {
		return config, err
	}
	return convert[Config](value)
}

func Watch[Config proto.Message](ctx context.Context, dsn string, notifyFunc func(conf Config), errFunc resource.ErrFunc) (resource.StopFunc, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("gonfig: parse dsn failed: %w", err)
	}
	factory, ok := resource.Get(u.Scheme)
	if !ok {
		return nil, errors.New("gonfig: resource factory not found for dsn " + dsn)
	}
	resource, err := factory.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
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
	unmarshalOptions := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	if err := unmarshalOptions.Unmarshal(data, config); err != nil {
		return config, err
	}
	return config, nil
}
