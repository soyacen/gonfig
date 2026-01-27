gen:
	protoc \
	--proto_path=. \
	--go_out=. \
	--go_opt=paths=source_relative \
	--gonfig_out=. \
	--gonfig_opt=paths=source_relative \
	example/configs/*.proto

run:
	go run ./cmd/main.go