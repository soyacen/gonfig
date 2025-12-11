gen:
	protoc \
	--proto_path=. \
	--proto_path=../proto \
	--go_out=. \
	--go_opt=paths=source_relative \
	--gonfig_out=. \
	--gonfig_opt=paths=source_relative \
	configs/*.proto

run:
	go run ./cmd/main.go