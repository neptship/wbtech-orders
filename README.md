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
make docker up
```

### 3. Примените миграции к базе данных

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest

make migrate up
```

> Начальная миграция: `000001_init_orders.sql` (создаёт таблицу `orders`).


### 4. Настройте переменные окружения

Скопируйте пример файла окружения и при необходимости измените значения:

```sh
cp .env.example .env
```

### 5. Запустите приложение

```sh
go run cmd/main.go
```

### 6. Откройте веб-интерфейс

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
