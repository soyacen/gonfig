package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-leo/gonfig/example/configs"
	"github.com/go-leo/gonfig/resource/env"
	"github.com/go-leo/gonfig/resource/file"
)

func main() {
	envExample()
	jsonExample()
	yamlExample()
}

func envExample() {
	errFunc := func(err error) {
		fmt.Println(err)
	}

	// 模拟环境变量
	os.Setenv("LEO_RUN_ENV", "dev")
	defer os.Unsetenv("LEO_RUN_ENV")

	// 创建环境变量资源
	envRsc, err := env.New("LEO_", time.Second)
	if err != nil {
		panic(err)
	}
	// 加载配置
	if err := configs.LoadEnvConfigConfig(context.TODO(), envRsc); err != nil {
		panic(err)
	}
	fmt.Println(configs.GetEnvConfigConfig().LEO_RUN_ENV)

	envStop, err := configs.WatchEnvConfigConfig(context.TODO(), envRsc, errFunc)
	if err != nil {
		panic(err)
	}
	defer envStop(context.TODO())

	os.Setenv("LEO_RUN_ENV", "prod")
	time.Sleep(3 * time.Second)
	fmt.Println(configs.GetEnvConfigConfig().LEO_RUN_ENV)
}

func jsonExample() {
	errFunc := func(err error) {
		fmt.Println(err)
	}
	tmpDir := os.TempDir()

	// 模拟json文件
	jsonFilename := tmpDir + "/config.json"
	if err := os.WriteFile(jsonFilename, genConfigJSON(), 0o644); err != nil {
		panic(err)
	}
	defer os.Remove(jsonFilename)
	// 创建json文件资源
	jsonRsc, err := file.New(jsonFilename)
	if err != nil {
		panic(err)
	}
	// 加载配置
	if err := configs.LoadJSONFileConfigConfig(context.TODO(), jsonRsc); err != nil {
		panic(err)
	}
	fmt.Println(configs.GetJSONFileConfigConfig())
	jsonStop, err := configs.WatchJSONFileConfigConfig(context.TODO(), jsonRsc, errFunc)
	if err != nil {
		panic(err)
	}
	defer jsonStop(context.TODO())
	go func() {
		time.Sleep(time.Second)
		if err := os.WriteFile(jsonFilename, genConfigJSON(), 0o644); err != nil {
			panic(err)
		}
	}()
	time.Sleep(3 * time.Second)
	fmt.Println(configs.GetJSONFileConfigConfig())
}

func yamlExample() {
	errFunc := func(err error) {
		fmt.Println(err)
	}
	tmpDir := os.TempDir()

	// 模拟yaml文件
	yamlFilename := tmpDir + "/config.yaml"
	if err := os.WriteFile(yamlFilename, genConfigYaml(), 0o644); err != nil {
		panic(err)
	}
	defer os.Remove(yamlFilename)
	// 创建yaml文件资源
	yamlRsc, err := file.New(yamlFilename)
	if err != nil {
		panic(err)
	}
	// 加载配置
	if err := configs.LoadYAMLFileConfigConfig(context.TODO(), yamlRsc); err != nil {
		panic(err)
	}
	fmt.Println(configs.GetYAMLFileConfigConfig())
	yamlStop, err := configs.WatchYAMLFileConfigConfig(context.TODO(), yamlRsc, errFunc)
	if err != nil {
		panic(err)
	}
	defer yamlStop(context.TODO())
	go func() {
		time.Sleep(time.Second)
		if err := os.WriteFile(yamlFilename, genConfigYaml(), 0o644); err != nil {
			panic(err)
		}
	}()
	time.Sleep(3 * time.Second)
	fmt.Println(configs.GetYAMLFileConfigConfig())
}

func genConfigJSON() []byte {
	return []byte(fmt.Sprintf(`{"addr":"127.0.0.1","port":%d}`, time.Now().Unix()))
}

func genConfigYaml() []byte {
	return []byte(fmt.Sprintf(`addr: 127.0.0.1:6379
network: tcp
password: oqnevaqm
db: %d`, time.Now().Unix()))
}
