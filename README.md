# Eda Top 1

## Description

Food order and delivery service.
You can register as customer, restaurant or courier.

## Features

- Orders pool with Kafka
- Metrics export using Prometheus and Grafana
- Unite entry point - api_gateway
- Full deployment with Docker-compose
- Separated microservices

## Architecture

Requst route:

    Client (Web) -> (HTTP)
    -> API Gateway (:8080) -> (gRPC)
        -> Auth Service (:9001)
        -> Restaurant Service (:9004)
        -> Customer Service (:9005)
        -> Courier Service (:9006)
            -> PostgreSQL (:5432)
            -> Kafka (:9092)

## Requirements

## Usage

## Microservices and endpoints

The service consists of 5 main microservices:

- API Gateway - entry point for users. Orchestrates all communications between microservices
- Auth - users registration
- Restaurant - menus creation and updating, restaurants listing
- Customer - orders creation, cancelling and listing
- Courier - checking for available orders, accepting orders and delivery confirmation

### API Gateway (:8080)

### Auth (:9001)

### Restaurant (:9004)

### Customer (:9005)

### Courier (:9006)

## Metrics

## DB description

## Testing
