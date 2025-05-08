# goCalc-v3 - Распределенный Калькулятор с Аутентификацией

<div align="center">
  <img src="https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go 1.20+"/>
  <img src="https://img.shields.io/badge/PostgreSQL-Required-336791?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL Required"/>
  <img src="https://img.shields.io/badge/Docker-Supported-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker Supported"/>
</div>

## 📋 Описание проекта

**goCalc-v3** - это эволюция распределенного калькулятора, реализованного на языке Go. Эта версия добавляет:

*   **Аутентификацию пользователей**: Регистрация и вход с использованием JWT.
*   **Персистентность**: Данные пользователей хранятся в PostgreSQL.
*   **Многопользовательский режим**: (В процессе) Вычисления привязаны к конкретному пользователю.


## 🔧 Требования

*   Go 1.20 или выше
*   Git
*   **PostgreSQL**: Запущенный и доступный экземпляр PostgreSQL.
*   Docker и Docker Compose (рекомендуется для легкого запуска PostgreSQL и самого приложения).

## 🚀 Установка

### 1. Клонирование репозитория

```bash
git clone https://github.com/superlogarifm/goCalc-v2.git
cd goCalc-v2
```

### 2. Установка зависимостей

```bash
go mod download
```

## ⚙️ Конфигурация

Приложение использует переменные окружения для конфигурации. Создайте файл `.env` в корне проекта или установите переменные перед запуском.

| Переменная        | Описание                                      | Значение по умолчанию (Пример!)                         | Обязательно |
| :---------------- | :-------------------------------------------- | :------------------------------------------------------ | :---------: |
| `DATABASE_URL`    | Строка подключения к PostgreSQL               | `postgres://postgres:postgres@localhost:5432/gocalc?sslmode=disable` |     ✅      |
| `JWT_SECRET_KEY`  | Секретный ключ для подписи JWT токенов       | `a-very-insecure-secret-key-replace-me`                 |     ✅      |
| `TOKEN_DURATION`  | Время жизни JWT токена (например, `24h`, `1h30m`) | `24h`                                                   |     ❌      |
| `HOST`            | Хост, на котором будет слушать сервис         | `127.0.0.1`                                             |     ❌      |
| `PORT`            | Порт, на котором будет слушать сервис         | `8080`                                                  |     ❌      |

**⚠️ Важно:**
*   Обязательно **замените** `JWT_SECRET_KEY` на ваш собственный, надежный ключ в производственной среде!
*   Убедитесь, что база данных (`gocalc` в примере `DATABASE_URL`) **существует** в вашем PostgreSQL. Сервис создаст нужные таблицы, но не саму базу данных.

*(Остальные переменные окружения для агентов/оркестратора, если они актуальны)*

| Переменная | Описание | Значение по умолчанию |
|------------|----------|------------------------|
| `TIME_ADDITION_MS` | Время выполнения операции сложения в мс | 5000 |
| `TIME_SUBTRACTION_MS` | Время выполнения операции вычитания в мс | 5000 |
| `TIME_MULTIPLICATIONS_MS` | Время выполнения операции умножения в мс | 5000 |
| `TIME_DIVISIONS_MS` | Время выполнения операции деления в мс | 5000 |


## ▶️ Запуск проекта

### Способ 1: Локальный запуск

1.  **Запустите PostgreSQL**: Убедитесь, что ваш сервер PostgreSQL запущен и доступен.
2.  **Установите переменные окружения**:
    ```bash
    # Linux/macOS
    export DATABASE_URL="postgres://user:pass@host:port/dbname?sslmode=disable"
    export JWT_SECRET_KEY="your_super_secret_key"
    # Windows (PowerShell)
    $env:DATABASE_URL="postgres://user:pass@host:port/dbname?sslmode=disable"
    $env:JWT_SECRET_KEY="your_super_secret_key"
    ```
3.  **Запустите сервис-калькулятор**:
    ```bash
    go run ./cmd/calc_service/start.go
    ```
4.  **Запустите агентов**:
    ```bash
    # Установите ORCHESTRATOR_URL, если необходимо
    go run ./cmd/agent/main.go
    ```

### Способ 2: Запуск с использованием Docker Compose (Рекомендуется)

Файл `docker-compose.yml` настроен для запуска сервиса-калькулятора и базы данных PostgreSQL.

1.  **(Опционально)** Создайте файл `.env` в корне проекта и переопределите переменные окружения при необходимости (особенно `JWT_SECRET_KEY`). Значения из `.env` будут автоматически подхвачены `docker-compose`.
    ```.env
    # Пример .env файла
    POSTGRES_PASSWORD=mysecretpassword # Пароль для пользователя postgres в контейнере БД
    JWT_SECRET_KEY=my_really_strong_and_secret_key_12345
    TOKEN_DURATION=8h
    # DATABASE_URL будет сформирован внутри docker-compose
    ```
2.  **Запустите сервисы**:
    ```bash
    docker-compose up --build -d
    ```
    * `--build` пересобирает образы, если код изменился.
    * `-d` запускает контейнеры в фоновом режиме.

3.  **Остановка сервисов**:
    ```bash
    docker-compose down
    ```

## 📡 API

Базовый URL: `http://localhost:8080` (или другой хост/порт, если настроено).

### Аутентификация

#### Регистрация пользователя

*   **Эндпоинт:** `POST /api/v1/register`
*   **Тело запроса:** `application/json`
    ```json
    {
      "login": "testuser",
      "password": "password123"
    }
    ```
*   **Ответ (Успех):** `200 OK`
*   **Ответ (Ошибка):**
    *   `400 Bad Request` (Неверное тело запроса, короткий пароль и т.д.)
    *   `409 Conflict` (Пользователь с таким логином уже существует)
    *   `500 Internal Server Error`

*   **Пример `curl`:**
    ```bash
    curl --location 'localhost:8080/api/v1/register' \
    --header 'Content-Type: application/json' \
    --data '{
        "login": "myuser",
        "password": "mypassword"
    }'
    ```

#### Вход пользователя

*   **Эндпоинт:** `POST /api/v1/login`
*   **Тело запроса:** `application/json`
    ```json
    {
      "login": "testuser",
      "password": "password123"
    }
    ```
*   **Ответ (Успех):** `200 OK` с телом:
    ```json
    {
      "token": "your_jwt_token_here"
    }
    ```
*   **Ответ (Ошибка):**
    *   `400 Bad Request` (Неверное тело запроса)
    *   `401 Unauthorized` (Неверный логин или пароль)
    *   `500 Internal Server Error`

*   **Пример `curl`:**
    ```bash
    curl --location 'localhost:8080/api/v1/login' \
    --header 'Content-Type: application/json' \
    --data '{
        "login": "myuser",
        "password": "mypassword"
    }'
    ```
    *(Скопируйте полученный `token` для следующих запросов)*

### Вычисления (Требуется аутентификация)

#### Отправка выражения на вычисление

*   **Эндпоинт:** `POST /api/v1/calculate`
*   **Заголовок:** `Authorization: Bearer <your_jwt_token_here>`
*   **Тело запроса:** `application/json`
    ```json
    {
      "expression": "2+2*2"
    }
    ```
*   **Ответ (Успех):** `200 OK` с телом:
    ```json
    {
      "result": 6
    }
    ```
*   **Ответ (Ошибка):**
    *   `400 Bad Request` (Неверное тело запроса, пустое выражение)
    *   `401 Unauthorized` (Нет токена, неверный токен, истекший токен)
    *   `422 Unprocessable Entity` (Ошибка вычисления выражения, например, деление на ноль)
    *   `500 Internal Server Error`

*   **Пример `curl` (замените `YOUR_TOKEN`):**
    ```bash
    TOKEN="YOUR_TOKEN" # Вставьте сюда токен, полученный при логине

    curl --location 'localhost:8080/api/v1/calculate' \
    --header "Authorization: Bearer $TOKEN" \
    --header 'Content-Type: application/json' \
    --data '{
        "expression": "(10+5)*2-3/1.5"
    }'
    ```

#### Получение статуса и результата выражения

После отправки выражения на вычисление с помощью эндпоинта `POST /api/v1/calculate`, вы получите `expression_id`. Используйте этот ID для запроса статуса и результата вычисления.

*   **Эндпоинт:** `GET /api/v1/expressions/{id}`
    *   Замените `{id}` на фактический `expression_id`.
*   **Заголовок:** `Authorization: Bearer <your_jwt_token_here>`
*   **Ответ (Успех):** `200 OK` с телом, содержащим детали выражения:
    ```json
    {
      "expression": {
        "id": "123", // ID вашего выражения
        "expression": "(10+5)*2-3/1.5", // Исходное выражение
        "status": "completed", // Статус: pending, processing, completed, error
        "result": 28.0,        // Результат вычисления (если status="completed")
        "error": null          // Сообщение об ошибке (если status="error")
      }
    }
    ```
    *   Поле `result` будет присутствовать и заполнено, если `status` равен `"completed"`.
    *   Поле `error` будет содержать сообщение об ошибке, если `status` равен `"error"`.
    *   Поле `expression` (внутри объекта expression) содержит исходное выражение и может отсутствовать в некоторых ответах или быть `omitempty`.

*   **Ответ (Ошибка):**
    *   `401 Unauthorized` (Нет токена, неверный токен, истекший токен)
    *   `404 Not Found` (Выражение с указанным ID не найдено)
    *   `500 Internal Server Error`

*   **Пример `curl` (замените `YOUR_TOKEN` и `EXPRESSION_ID`):**
    ```bash
    TOKEN="YOUR_TOKEN" 
    EXPRESSION_ID="YOUR_EXPRESSION_ID" # ID, полученный от /api/v1/calculate

    curl --location "localhost:8080/api/v1/expressions/$EXPRESSION_ID" \
    --header "Authorization: Bearer $TOKEN"
    ```

*(Остальные эндпоинты, если они актуальны, например, /api/v1/expressions/{id})*

## ⚠️ Устранение неполадок

### Ошибки при запуске

*   **`Failed to connect to database`**: Проверьте переменную `DATABASE_URL` и доступность сервера PostgreSQL. Убедитесь, что пользователь и пароль верны, и база данных существует.
*   **`JWT_SECRET_KEY environment variable not set`**: Установите переменную окружения `JWT_SECRET_KEY`.

### Проблемы с Docker Compose

*   **Контейнер `db` не запускается**: Проверьте логи контейнера (`docker-compose logs db`). Возможно, проблема с паролем или нехваткой ресурсов.
*   **Контейнер `calc_service` не запускается**: Проверьте логи (`docker-compose logs calc_service`). Часто это связано с невозможностью подключиться к БД (контейнер `db` еще не успел запуститься или неверный `DATABASE_URL` внутри docker-compose). `docker-compose.yml` обычно настраивает `depends_on`, но иногда требуется дополнительная логика ожидания.

### Логи контейнеров

```bash
docker-compose logs             # Показать логи всех сервисов
docker-compose logs -f db       # Показать логи БД и следить за ними
docker-compose logs calc_service # Показать логи сервиса калькулятора
```

 
