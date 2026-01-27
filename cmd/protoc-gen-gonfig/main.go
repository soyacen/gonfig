package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	config "github.com/soyacen/gonfig/cmd/protoc-gen-gonfig/gen"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var Version = "v0.0.5"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Fprintf(os.Stdout, "%v %v\n", filepath.Base(os.Args[0]), Version)
		return
	}

	var flags flag.FlagSet
	options := &protogen.Options{ParamFunc: flags.Set}
	options.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		return generate(plugin)
	})
}

func generate(plugin *protogen.Plugin) error {
	for _, file := range plugin.Files {
		if !file.Generate {
			continue
		}

		// 配置生成
		configGenerator := config.NewGenerator(plugin, file)
		if err := configGenerator.Generate(); err != nil {
			return err
		}
	}
	return nil
}
