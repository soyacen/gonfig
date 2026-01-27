module github.com/soyacen/gonfig/resource/consul

go 1.25.5

replace github.com/soyacen/gonfig => ../../

require (
	github.com/hashicorp/consul/api v1.33.2
	github.com/hashicorp/go-hclog v1.6.3
	github.com/soyacen/gonfig v0.0.6
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96
	golang.org/x/net v0.43.0
	golang.org/x/sys v0.40.0
	golang.org/x/text v0.28.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.4 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/serf v0.10.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
)
