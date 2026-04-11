EXECUTABLE_NAME=eda-one

PROTO_PKG=./pkg/api
PB_DIR=./services/pb

PROTO_COMMON_PATH=$(PB_DIR)/common.proto
PROTO_CUSTOMER_PATH=$(PB_DIR)/customer.proto
PROTO_COURIER_PATH=$(PB_DIR)/courier.proto

CUSTOMER_PATH=./services/customer/cmd/main.go
COURIER_PATH=./services/courier/cmd/main.go

all: proto

proto: generate-common generate-customer generate-courier
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
