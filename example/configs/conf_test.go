package configs

import (
	"encoding/json"
	"fmt"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

func TestConfig(t *testing.T) {
	conf := &Config{
		Server: &ServerConfig{
			Addr: "127.0.0.1",
			Port: 8080,
		},
		Db: &DBConfig{
			Dsn: "mysql://user:password@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local",
		},
		Redis: &RedisConfig{
			Addr:     "127.0.0.1:6379",
			Password: "123456",
			Db:       1,
		},
	}
	jsonData, err := protojson.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]any{}
	if err := json.Unmarshal(jsonData, &m); err != nil {
		t.Fatal(err)
	}
	yamlData, err := yaml.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(yamlData))
}
