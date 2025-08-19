package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-leo/config/example/configs"
	"github.com/go-leo/config/resource/env"
	"github.com/go-leo/config/resource/file"
)

func main() {
	// prepare config

	// 模拟环境变量
	os.Setenv("LEO_RUN_ENV", "dev")
	defer os.Unsetenv("LEO_RUN_ENV")

	tmpDir := os.TempDir()

	// 模拟json文件
	jsonFilename := tmpDir + "/config.json"
	if err := os.WriteFile(jsonFilename, genConfigJSON(), 0o644); err != nil {
		panic(err)
	}
	defer os.Remove(jsonFilename)

	// 模拟yaml文件
	yamlFilename := tmpDir + "/config.yaml"
	if err := os.WriteFile(yamlFilename, genConfigYaml(), 0o644); err != nil {
		panic(err)
	}
	defer os.Remove(yamlFilename)

	// 创建环境变量资源
	envRsc, err := env.New("LEO_")
	if err != nil {
		panic(err)
	}
	// 创建json文件资源
	jsonRsc, err := file.New(jsonFilename)
	if err != nil {
		panic(err)
	}
	// 创建yaml文件资源
	yamlRsc, err := file.New(yamlFilename)
	if err != nil {
		panic(err)
	}
	// 加载配置
	if err := configs.LoadApplicationConfig(context.TODO(), envRsc, jsonRsc, yamlRsc); err != nil {
		panic(err)
	}
	// 获取配置
	fmt.Println(configs.GetApplicationConfig())
	// 监听配置
	// sigC 当有配置更新时，会发送通知。
	// stop 用于停止监听。
	sigC, stop, err := configs.WatchApplicationConfig(context.TODO(), envRsc, jsonRsc, yamlRsc)
	if err != nil {
		panic(err)
	}
	go func() {
		time.Sleep(10 * time.Second)
		stop(context.TODO())
	}()

	go func() {
		for range sigC {
			fmt.Println(configs.GetApplicationConfig())
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			if err := os.WriteFile(jsonFilename, genConfigJSON(), 0o644); err != nil {
				panic(err)
			}
			if err := os.WriteFile(yamlFilename, genConfigYaml(), 0o644); err != nil {
				panic(err)
			}
		}
	}()

	time.Sleep(11 * time.Second)
}

func genConfigJSON() []byte {
	return []byte(fmt.Sprintf(`{"grpc":{"addr":"127.0.0.1","port":%d}}`, time.Now().Unix()))
}

func genConfigYaml() []byte {
	return []byte(fmt.Sprintf(`
redis:
  addr: 127.0.0.1:6379
  network: tcp
  password: oqnevaqm
  db: %d`, time.Now().Unix()))
}
