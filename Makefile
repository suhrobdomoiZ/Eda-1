EXECUTABLE_NAME=eda-one

PROTO_PKG=./pkg/api
PROTO_RESTAURANT_PATH=./services/pb/restaurant.proto
RESTAURANT_PATH=./services/restaurant/cmd/main.go

all: generate

generate:
	@echo "Generating restaurant.pb.go"
	protoc -I ./ \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_RESTAURANT_PATH)