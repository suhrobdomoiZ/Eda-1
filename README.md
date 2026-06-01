# Eda Top 1

## Description

Food order and delivery service built with go microservices.
You can register as **customer**, **restaurant** or **courier**.

## Features

- Microservice architecture (gRPC)
- API Gateway (single entry point)
- JWT authentication + refresh tokens
- Order lifecycle management
- Kafka (orders/event streaming)
- Observability (Prometheus + Grafana)
- Docker Compose deployment
- JWT aunthentification
- Multi-role registration

## Architecture

Requst route:

    Client (Web) -> (HTTP)
    -> API Gateway (:80) -> (gRPC)
        -> Auth Service (:9001)
        -> Restaurant Service (:9004)
        -> Customer Service (:9005)
        -> Courier Service (:9006)
            -> PostgreSQL (:5432)
            -> Kafka (:9092)

## Requirements

- Go 1.26.1+
- Docker 24.0+
- Docker Compose 2.20+
- PostgreSQL 16
- Redis 7
- Kafka latest
- Create `.env` file in your folder (you can copy `.env.example`)
- go libs are listed in go.mod

## Usage

Check Makefile `make help` to see available commands.
All incoming requests go to API Gateway on port 8080 (default)

## Microservices and endpoints

The service consists of 5 main microservices:

- **API Gateway** - entry point, routing, auth, RBAC
- **Auth** - registration, login, JWT, sessions
- **Restaurant** - menu & order management
- **Customer** - orders creation & tracking
- **Courier** - delivery workflow

### API Gateway (:8080)

Single entry point for all clients. Routes requests to internal gRPC services and handles cross-cutting concerns.

#### Gateway Responsibilities

- Routing (HTTP to gRPC)
- Authentication (JWT validation)
- Authorization (role-based access)
- Logging / tracing
- Error mapping (gRPC to HTTP)

---

#### Gateway Architecture

Client - API Gateway - gRPC Services:

- Auth Service
- Customer Service
- Courier Service
- Restaurant Service

---

#### Authentication Flow

1. Client sends request with:

- Authorization: Bearer <access_token>

1. Gateway:

- calls `AuthService.ValidateToken`

1. If valid:

- extracts `user_id`, `role`
- injects into context / headers

1. Forwards request to target service

---

#### Authorization

Role-based access control (RBAC):

- CUSTOMER to Customer service
- COURIER to Courier service
- RESTAURANT to Restaurant service

Gateway must reject unauthorized roles early.

#### Context Propagation

Gateway should pass:

- `user_id`
- `role`

via:

- gRPC metadata

#### Error Handling

- gRPC errors to HTTP status codes
- hide internal details
- standard response format

### Auth (:9001)

Authentication and authorization service (JWT + Refresh Token).

#### Overview

- Access Token - ~15 min
- Refresh Token - ~7 days
- Access - stateless
- Refresh - stored (Redis/DB)

---

### Restaurant (:9004)

Menu management and order processing for restaurant owners.

#### Restaurant Responsibilities

- CRUD operations for menu items (products)
- View incoming orders
- Update order status (cooking → ready)
- Publish order status changes to Kafka
- List all restaurants (public endpoint)

#### Restaurant API Endpoints

| Method | Description | Access |
| ------ | ----------- | ------ |
| POST /menu | Add product | Restaurant only |
| PUT /menu/:id | Update product | Restaurant only |
| DELETE /menu/:id | Delete product | Restaurant only |
| GET /menu/:restaurant_id | List products | Public |
| GET /restaurants | List all restaurants | Public |

#### Order Processing Flow

1. Receives ORDER_CREATED event from Kafka (customer created order)
2. Restaurant views order → changes status to cooking
3. When ready → changes status to ready
4. Publishes ORDER_READY event to Kafka
5. Courier service picks up the order

#### Restaurant Database Tables

- products — menu items with price, description, availability
- restaurant_profiles — extended restaurant info (name, address, phone)

### Customer (:9005)

Order management for customers.

#### Customer Responsibilities

- Create new orders
- View order history
- Cancel orders (only before cooking starts)
- Publish order events to Kafka

#### Order Lifecycle (Customer View)

| Status | Description | Can Cancel? |
| ------ | ----------- | ----------- |
| created | Order placed, waiting for restaurant | Yes |
| confirmed | Restaurant accepted | Yes |
| cooking | Restaurant started cooking | No |
| ready | Waiting for courier | No |
| delivering | Courier on the way | No |
| delivered | Order completed | No |
| cancelled | Order cancelled | Already cancelled |

#### Customer API Endpoints

| Method | Path | Description |
| ------ | ---- | ----------- |
| POST | /orders | Create order |
| GET | /orders | List my orders (with pagination) |
| GET | /orders/:id | Get order details |
| DELETE | /orders/:id | Cancel order |

#### Create Order Flow

1. Customer sends `restaurant_id`, `items[]`, `address`
2. Service calculates total price
3. Creates order with status created
4. Publishes `ORDER_CREATED` event to Kafka
5. Returns `order_id`

#### Cancel Order Flow

1. Checks order belongs to user
2. Verifies order status is cancellable (created or confirmed)
3. Updates status to cancelled
4. Publishes ORDER_CANCELLED event to Kafka
5. Returns refund amount

### Courier (:9006)

Delivery management for couriers.

#### Courier Responsibilities

- View available orders (status = ready)
- Accept order (assign to courier)
- Mark order as picked up from restaurant
- Mark order as delivered to customer
- View delivery history and earnings
- Publish delivery status updates to Kafka

#### Courier Order Flow

| Step | Action | Status Change | Kafka Event |
| ---- | ------ | ------------- | ----------- |
| 1 | View available | ready | — |
| 2 | Accept order | ready → delivering | ORDER_DELIVERING |
| 3 | Pick up from restaurant | delivering → picked_up | ORDER_PICKED_UP |
| 4 | Deliver to customer | picked_up → delivered | ORDER_DELIVERED |

#### Active Orders Limit

- Maximum 3 active orders per courier
- Active = delivering
- Prevents overloading

#### Courier API Endpoints
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/orders/available` | List available orders (ready) |
| `POST` | `/orders/:id/accept` | Accept order |
| `GET` | `/orders` | My orders (history + active) |
| `POST` | `/orders/:id/pickup` | Mark as picked up |
| `POST` | `/orders/:id/deliver` | Mark as delivered (returns earnings) |

#### Courier Database Tables

- Uses shared orders table
- courier_id field links order to courier
- courier_profiles — extended courier info (name, phone)

## Metrics

Technical and buisiness metrics are collected and can be observed in Grafana (:3000)
For now the metrics are:

- RequestsTotal    - CounterVec
- RequestDuration  - HistogramVec
- RequestsInFlight - Gauge
- OccusedErrors    - CounterVec
- OrdersCreated    - Counter
- OrdersDelivered  - Counter
- OrdersCancelled  - Counter
- OrdersInProgress - Gauge
- OrdersWaiting    - Gauge
- OrdersDelivering - Gauge
- OrdersPrices     - Histogram

And more can be easily added on your purpose.

## Testing

Tested via CI/CD Gitlab feature (check `.gitlab-cy.yml`)

## End-to-End Api Flow

Full system interaction scenarios.

## 1. Registration

### Customer

`POST /api/v1/auth/register`

```
json
{
  "username": "alice",
  "password": "secret123",
  "role": "customer"
}
```

### Restaurant

```json
{
  "username": "pizza_place",
  "password": "secret123",
  "role": "restaurant",
  "restaurant_name": "Pizza Place",
  "restaurant_address": "ул. Ленина 1",
  "restaurant_phone": "+7 777 000 00 00"
}
```

### Courier

```json
{
  "username": "courier_ivan",
  "password": "secret123",
  "role": "courier",
  "courier_name": "Иван Петров",
  "courier_phone": "+7 777 111 22 33"
}
```

## 2. Authentication

### Login

`POST /api/v1/auth/login`

```json
{
  "username": "alice",
  "password": "secret123"
}
```

### Invalid password = 401 status

```json
{
  "username": "alice",
  "password": "wrongpassword"
}
```

## Get profile

`GET /api/v1/auth/profile`

### Header: Authorization: Bearer <access_token>

## Refresh token

`POST /api/v1/auth/refresh`

```json
{
    "refresh_token": "<refresh_token>"
}
```

## 3. Restaurant Flow (Menu Management)

### Add product

`POST /api/v1/restaurant/menu`

```json
{
  "name": "Маргарита",
  "description": "Томат, моцарелла, базилик",
  "price": 59000
}
```

### Update product

`PUT /api/v1/restaurant/menu/{product_id}`

### Delete product

`DELETE /api/v1/restaurant/menu/{product_id}`

## 4. Public endpoints

`GET /api/v1/customer/restaurants`
`GET /api/v1/customer/restaurants/{id}/menu`
`GET /api/v1/restaurant/menu/{restaurant_id}`
`GET /api/v1/restaurant/menu/{restaurant_id}/product/{product_id}`

## 5. Customer flow (Orders)

### Create order

`POST /api/v1/customer/orders`

```json
{
  "restaurant_id": "<restaurant_user_id>",
  "items": [
    { "product_id": "<product_1_id>", "quantity": 2 },
    { "product_id": "<product_2_id>", "quantity": 1 }
  ],
  "address": "ул. Пушкина 10, кв. 5"
}
```

### List orders

`GET /api/v1/customer/orders`

### Get order

`GET /api/v1/customer/orders/{order_id}`

## 6. Restaurant Order Flow

### View Orders

`GET /api/v1/restaurant/orders`

### Update Status

`PUT /api/v1/restaurant/orders/<order_id>/status`

``` json
{
  "status": 1
}
// ORDER_STATUS_COOKING
```

## 7. Courier Flow

### Available Orders

`GET /api/v1/courier/orders/available`

### Accept Order

`POST /api/v1/courier/orders/{order_id}/accept`

### Pickup Order

`POST /api/v1/courier/orders/{order_id}/pickup`

### Deliver Order

`POST /api/v1/courier/orders/{order_id}/deliver`

## 8. Cancel Flow

### Cancel (CREATED only)

`DELETE /api/v1/customer/orders/{order_id}`

## 9. Logout

### Logout

`POST /api/v1/auth/logout`

```json
{
  "refresh_token": "<refresh_token>"
}
```
