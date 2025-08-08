# WBTech Orders Demo Service

Демонстрационный микросервис для хранения и отображения данных о заказах.

## Быстрый старт

### 1. Клонируйте репозиторий

```sh
git clone https://github.com/neptship/wbtech-orders.git
cd wbtech-orders
```

### 2. Запустите сервисы через Docker Compose

```sh
docker-compose up -d
```

### 3. Примените миграции к базе данных

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest

goose -dir ./internal/storage/migrations postgres "host=localhost user=postgres 
dbname=orders_data password=postgres sslmode=disable" up
```

### 4. Запустите приложение

```sh
go run cmd/main.go
```

### 5. Откройте веб-интерфейс

Перейдите в браузере по адресу:  
[http://localhost:8081/](http://localhost:8081/)

---

## Эндпоинты

- **Получить заказ по ID (JSON):**
  ```
  GET /order/<order_uid>
  ```

- **Добавить заказ по ID (JSON):**
  ```
  POST /order_add
  ```
