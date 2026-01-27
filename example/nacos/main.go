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
		Type:    "yaml",
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
