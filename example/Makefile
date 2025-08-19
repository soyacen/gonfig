gen:
	protoc \
	--proto_path=. \
	--proto_path=../proto \
	--go_out=. \
	--go_opt=paths=source_relative \
	--config_out=. \
	--config_opt=paths=source_relative \
	configs/*.proto

run:
	go run ./cmd/main.go