PB_DIR=./services/pb

PROTO_COMMON_PATH=$(PB_DIR)/common.proto
PROTO_CUSTOMER_PATH=$(PB_DIR)/customer.proto
PROTO_COURIER_PATH=$(PB_DIR)/courier.proto
PROTO_RESTAURANT_PATH=$(PB_DIR)/restaurant.proto
PROTO_AUTH_PATH=$(PB_DIR)/auth.proto

DOCKER_COMPOSE=docker-compose
GO=go
GOFMT=gofmt

all: proto mod-tidy mod-download

# Proto
.PHONY: all proto

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

# Docker
.PHONY: up down restart logs ps clean

up:
	@echo "Starting all services..."
	$(DOCKER_COMPOSE) up -d
	@echo "Services started"
	@make ps

down:
	@echo "Stopping all services..."
	$(DOCKER_COMPOSE) down
	@echo "Services stopped"

restart: down up

logs:
	$(DOCKER_COMPOSE) logs -f

logs-gateway:
	$(DOCKER_COMPOSE) logs -f api_gateway

logs-auth:
	$(DOCKER_COMPOSE) logs -f auth

logs-customer:
	$(DOCKER_COMPOSE) logs -f customer

logs-restaurant:
	$(DOCKER_COMPOSE) logs -f restaurant

logs-courier:
	$(DOCKER_COMPOSE) logs -f courier

logs-kafka:
	$(DOCKER_COMPOSE) logs -f kafka

logs-db:
	$(DOCKER_COMPOSE) logs -f database

ps:
	$(DOCKER_COMPOSE) ps

clean:
	@echo "Cleaning all containers, volumes, and images..."
	$(DOCKER_COMPOSE) down -v --rmi all
	@echo "Cleaned"

# Build
.PHONY: build build-gateway build-auth build-customer build-restaurant build-courier

build:
	@echo "Building all services..."
	$(DOCKER_COMPOSE) build

build-gateway:
	$(DOCKER_COMPOSE) build api_gateway

build-auth:
	$(DOCKER_COMPOSE) build auth

build-customer:
	$(DOCKER_COMPOSE) build customer

build-restaurant:
	$(DOCKER_COMPOSE) build restaurant

build-courier:
	$(DOCKER_COMPOSE) build courier

# Code
.PHONY: mod-tidy mod-download

mod-tidy:
	@echo "Tidying go.mod..."
	$(GO) mod tidy

mod-download:
	@echo "Downloading dependencies..."
	$(GO) mod download

# CI/CD
.PHONY: ci-test

ci-test:
	go test ./services/restaurant/... -v
	go test ./services/courier/... -v
	go test ./services/customer/... -v

# Help
.PHONY: help

help:
	@echo "Eda Top 1 - Available Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Proto:"
	@echo "  make proto              - Generate all protobuf files"
	@echo "  make generate-common    - Generate common.pb.go"
	@echo "  make generate-customer  - Generate customer.pb.go"
	@echo "  make generate-courier   - Generate courier.pb.go"
	@echo "  make generate-restaurant- Generate restaurant.pb.go"
	@echo "  make generate-auth      - Generate auth.pb.go"
	@echo ""
	@echo "Docker:"
	@echo "  make up                 - Start all services"
	@echo "  make down               - Stop all services"
	@echo "  make restart            - Restart all services"
	@echo "  make logs               - View all logs"
	@echo "  make logs-gateway       - View API Gateway logs"
	@echo "  make logs-auth          - View Auth service logs"
	@echo "  make logs-customer      - View Customer service logs"
	@echo "  make logs-restaurant    - View Restaurant service logs"
	@echo "  make logs-courier       - View Courier service logs"
	@echo "  make logs-kafka         - View Kafka logs"
	@echo "  make logs-db            - View Database logs"
	@echo "  make ps                 - Show running containers"
	@echo "  make clean              - Remove all containers, volumes, images"
	@echo ""
	@echo "Build:"
	@echo "  make build              - Build all Docker images"
	@echo "  make build-gateway      - Build API Gateway"
	@echo "  make build-auth         - Build Auth service"
	@echo "  make build-customer     - Build Customer service"
	@echo "  make build-restaurant   - Build Restaurant service"
	@echo "  make build-courier      - Build Courier service"
	@echo ""
	@echo "Code:"
	@echo "  make mod-tidy           - Tidy go.mod"
	@echo "  make mod-download       - Download all dependencies"
	@echo ""
	@echo "Default:"
	@echo "  make all                - Generate proto + tidy + download deps"
