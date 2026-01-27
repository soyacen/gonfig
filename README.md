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
go get github.com/soyacen/gonfig
```

## 安装 protoc-gen-gonfig 插件

```bash
make install
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

### 2. 使用环境变量配置

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    
    "github.com/soyacen/gonfig/example/configs"
    "github.com/soyacen/gonfig/resource/env"
)

func main() {
    // 设置环境变量
    os.Setenv("ADDR", "localhost")
    os.Setenv("PORT", "8080")
    os.Setenv("ENVIRONMENT", "development")
    
    // 创建环境变量资源配置
    envResource, err := env.New("", time.Second)
    if err != nil {
        panic(err)
    }
    
    // 加载配置
    if err := configs.LoadConfig(context.TODO(), envResource); err != nil {
        panic(err)
    }
    
    // 获取配置
    config := configs.GetConfig()
    fmt.Printf("Address: %s\n", config.Addr)
    fmt.Printf("Port: %d\n", config.Port)
    fmt.Printf("Environment: %s\n", config.Environment)
    
    // 也可以直接获取字段值
    fmt.Printf("Direct field access - Address: %s\n", configs.GetAddr())
    fmt.Printf("Direct field access - Port: %d\n", configs.GetPort())
    fmt.Printf("Direct field access - Environment: %s\n", configs.GetEnvironment())
}
```

### 3. 监听配置变化

```go
errFunc := func(err error) {
    fmt.Printf("配置监听错误: %v\n", err)
}

// 开始监听配置变化
stop, err := configs.WatchConfig(context.TODO(), envResource, errFunc)
if err != nil {
    panic(err)
}
defer stop(context.TODO())
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
- **ENV**: 环境变量格式（键值对）

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