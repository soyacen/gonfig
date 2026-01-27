# gonfig - Go 配置管理库

gonfig 是一个功能强大的 Go 语言配置管理库，支持多种配置源和格式，提供了统一的配置加载和监听机制。

## 特性

- **多源支持**：支持从环境变量、文件、Consul、Nacos 等多种来源加载配置
- **多格式支持**：内置支持 JSON、YAML、TOML、ENV 等常见配置格式
- **热更新**：支持配置变更监听和自动重新加载
- **Protobuf 集成**：与 Protobuf 深度集成，可自动生成配置管理代码
- **类型安全**：基于结构化数据和强类型配置对象
- **易于扩展**：提供清晰的接口用于添加新的配置源和格式

## 安装

```bash
go get github.com/soyacen/gonfig@latest
```

## 安装 protoc-gen-gonfig 插件

```bash
go install github.com/soyacen/gonfig/cmd/protoc-gen-gonfig@latest
```

## 快速开始

### 1. 定义配置结构

首先使用 Protobuf 定义配置结构，消息名称必须是 `Config`、`Conf` 或 `Configuration` 之一：

```protobuf
syntax = "proto3";
package example;
option go_package = "github.com/soyacen/gonfig/example/configs;configs";

message Config {
  string addr = 1;
  int32 port = 2;
  string environment = 3;
}
```

运行以下命令生成代码：

```bash
protoc --go_out=. --gonfig_out=. configs/*.proto
```

### 2. 使用 Nacos 配置中心

```go
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/soyacen/gonfig/example/configs"
	"github.com/soyacen/gonfig/resource/nacos"
)

func main() {
	configClient, err := nacosFactory()
	if err != nil {
		log.Fatal(err)
	}

	errFunc := func(err error) {
		fmt.Println(err)
	}

	dataId := "config.yaml"
	group := "example"
	_, err = configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: string(genConfigYaml()),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_, err = configClient.DeleteConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  group,
		})
		if err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(5 * time.Second)
	rsc, err := nacos.New(configClient, group, dataId)
	if err != nil {
		log.Fatal(err)
	}
	// 加载配置
	if err := configs.LoadConfig(context.TODO(), rsc); err != nil {
		panic(err)
	}
	// 输出配置
	fmt.Println(configs.GetConfig())

	yamlStop, err := configs.WatchConfig(context.Background(), rsc, errFunc)
	if err != nil {
		panic(err)
	}
	defer yamlStop(context.Background())
	go func() {
		time.Sleep(time.Second)
		content := string(genConfigYaml())
		content = strings.ReplaceAll(content, "123456", "654321")
		_, err = configClient.PublishConfig(vo.ConfigParam{
			DataId:  dataId,
			Group:   group,
			Content: content,
		})
		if err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(5 * time.Second)
	fmt.Println(configs.GetConfig())
}

func genConfigYaml() []byte {
	return []byte(`db:
    dsn: mysql://user:password@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local
redis:
    addr: 127.0.0.1:6379
    db: 1
    password: "123456"
server:
    addr: 127.0.0.1
    port: 8080`)
}

func nacosFactory() (config_client.IConfigClient, error) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}
	cc := *constant.NewClientConfig(
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("debug"),
		constant.WithLogDir("/tmp/nacos.log"),
	)
	return clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
}
```

## 支持的配置源

### 1. 环境变量 (env)

```go
import "github.com/soyacen/gonfig/resource/env"

resource, err := env.New("PREFIX_", time.Second)
```

### 2. 文件 (file)

```go
import "github.com/soyacen/gonfig/resource/file"

resource, err := file.New("/path/to/config.json")
```

### 3. Consul

```go
import "github.com/soyacen/gonfig/resource/consul"

client, _ := api.NewClient(api.DefaultConfig())
resource, err := consul.New(client, "config/key")
```

### 4. Nacos

```go
import "github.com/soyacen/gonfig/resource/nacos"

resource, err := nacos.New(client, "group", "dataId")
```

## 支持的配置格式

- **JSON**: `.json` 文件扩展名
- **YAML**: `.yaml` 或 `.yml` 文件扩展名
- **TOML**: `.toml` 文件扩展名
- **ENV**:  `.env` 文件扩展名或环境变量格式（键值对）

## Protobuf 消息命名约定

代码生成器通过检查 Protobuf 消息的名称来决定是否为其生成配置管理代码。只要消息名称是 `Config`、`Conf` 或 `Configuration` 之一，就会自动生成相应的配置管理代码。

不需要使用任何特殊的注解或选项。

这会生成以下辅助函数：

- [GetConfig()](file:///home/soyacen/Workspace/github.com/soyacen/gonfig/cmd/protoc-gen-gonfig/gen/generator.go#L87-L89) - 获取当前配置实例
- [LoadConfig()](file:///home/soyacen/Workspace/github.com/soyacen/gonfig/cmd/protoc-gen-gonfig/gen/generator.go#L95-L97) - 从资源加载配置
- [WatchConfig()](file:///home/soyacen/Workspace/github.com/soyacen/gonfig/cmd/protoc-gen-gonfig/gen/generator.go#L99-L101) - 监听配置变化
- `GetFieldName()` - 直接获取字段值的函数（例如 `GetAddr()`、`GetPort()`）

## 生成的代码结构

代码生成器会为每个匹配的消息名称（`Config`、`Conf` 或 `Configuration`）生成以下内容：

1. 一个全局变量存储配置（使用 `sync/atomic.Value` 类型）
2. `init()` 函数初始化全局配置变量
3. [LoadConfig()](file:///home/soyacen/Workspace/github.com/soyacen/gonfig/cmd/protoc-gen-gonfig/gen/generator.go#L95-L97) 函数用于从指定资源加载配置
4. [WatchConfig()](file:///home/soyacen/Workspace/github.com/soyacen/gonfig/cmd/protoc-gen-gonfig/gen/generator.go#L99-L101) 函数用于监听配置变化
5. [GetConfig()](file:///home/soyacen/Workspace/github.com/soyacen/gonfig/cmd/protoc-gen-gonfig/gen/generator.go#L87-L89) 函数用于获取当前配置的副本
6. 每个字段的独立获取函数（如 `GetAddr()`、`GetPort()` 等）

## 注意事项

- Protobuf 消息名称必须是 `Config`、`Conf` 或 `Configuration` 之一才能被识别并生成代码
- 不支持 `oneof` 字段类型
- 所有生成的函数都是线程安全的
- 使用 `google.golang.org/protobuf/proto.Clone` 来确保配置的深拷贝
- 代码生成完全基于消息名称，不依赖任何 Protobuf 注解

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。