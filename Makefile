PB_DIR=./services/pb

PROTO_COMMON_PATH=$(PB_DIR)/common.proto
PROTO_CUSTOMER_PATH=$(PB_DIR)/customer.proto
PROTO_COURIER_PATH=$(PB_DIR)/courier.proto
PROTO_RESTAURANT_PATH=$(PB_DIR)/restaurant.proto
PROTO_AUTH_PATH=$(PB_DIR)/auth.proto

all: proto

proto: generate-common generate-customer generate-courier generate-restaurant \
	generate-auth
	@echo "All proto generated"

generate-common:
	@echo "Generating common.pb.go"
	protoc -I $(PB_DIR) \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		--go-grpc_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		$(PROTO_COMMON_PATH)

generate-customer:
	@echo "Generating customer.pb.go"
	protoc -I $(PB_DIR) \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		--go-grpc_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		$(PROTO_CUSTOMER_PATH)

generate-courier:
	@echo "Generating courier.pb.go"
	protoc -I $(PB_DIR) \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		--go-grpc_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		$(PROTO_COURIER_PATH)

generate-restaurant:
	@echo "Generating restaurant.pb.go"
	protoc -I $(PB_DIR) \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		--go-grpc_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		$(PROTO_RESTAURANT_PATH)

generate-auth:
	@echo "Generating auth.pb.go"
	protoc -I $(PB_DIR) \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		--go-grpc_opt=module=github.com/suhrobdomoiZ/Eda-1 \
		$(PROTO_AUTH_PATH)
