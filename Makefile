.PHONY: install
install:
	go install ./cmd/protoc-gen-gonfig

.PHONY: test
test:
	go test -v ./...

.PHONY: compile-proto
compile-proto:
	protoc \
	--proto_path=. \
	--go_out=. \
	--go_opt=paths=source_relative \
	proto/gonfig/*.proto

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

all: compile-proto install compile-example run-example

