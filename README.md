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

Client (Web) -> (HTTP)
-> API Gateway (:8080) -> (gRPC)
    -> Auth Service (:9001)
    -> Restaurant Service (:9004)
    -> Customer Service (:9005)
    -> Courier Service (:9006)
        -> PostgreSQL (:5432)
        -> Kafka (:9092)

## Microservices and endpoints

## API Gateway (:8080)

## Auth (:9001)

## Restaurant (:9004)

## Customer (:9005)

## Courier (:9006)

# Сущности
## Users
- id UUID
- role VARCHAR(32) ("user", "admin", "restaurant", 'courier')
- username
- password
## Product
- id UUID
- restaurant_id
- name
- description
- price
## Orders
- id UUID
- restaurant_id
- deliver_id
- client_id
- address
- status:**("created", "made by restaurant", "delivered", "cancelled")**
## Ordered_products
- order_id PRIMARY KEY
- item_id
- count


# Микросервисы

## API-gateway port:8080
- Принимает запросы и валидируем их
- перенаправляет куда надо
## Авторизация port:8081
- Собирает данные с формы регистрации/авторизации
- Делает запрос/в БД
- Возвращает ответ

## Ресторан port:8082
#### Управление меню
- CRUD обычный
#### Взаимодействие с заказом
- Получает информацию о заказе(создали, отменили, доставили)

## Клиент port:8083
- Создать заказ/отменить заказ
- получать статус заказа
- список ресторанов

## Курьер port:8084
- Получить заказ
- Отдать заказ


# Глобально
- БД
- Сервер(внутрянка вся)
- Метрики
- Кафка