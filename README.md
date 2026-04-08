# Сущности
## Users
- id UUID
- role VARCHAR(32) ("user", "admin", "restaurant", "deliver")
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
- id UUID
- order_id
- item_id
- count
## Menu
- restaurant_id
- product_id

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