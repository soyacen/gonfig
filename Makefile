.PHONY: install
install:
	go install ./cmd/protoc-gen-gonfig

.PHONY: test
test:
	go test -v ./...

.PHONY: example
example:
	protoc \
	--proto_path=. \
	--go_out=. \
	--go_opt=paths=source_relative \
	--gonfig_out=. \
	--gonfig_opt=paths=source_relative \
	example/configs/*.proto

run-example:
	go run ./example/cmd/main.go

all: install compile-example run-example

