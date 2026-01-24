# DonWeather-ms-weather 🌦️

Микросервис для управления погодными данными в рамках проекта DonWeather.

## Описание

Этот сервис предоставляет API для регистрации и получения погодных данных. Он использует PostgreSQL для хранения данных и поддерживает миграции базы данных. Использует Clean Architecture с разделением на слои: delivery, usecase, repository.

## Требования

* Go 1.19 или выше
* Docker и Docker Compose (для запуска PostgreSQL)
* PostgreSQL (или используйте Docker Compose для локального запуска)
* API ключ от WeatherAPI.com

## Установка

1. Клонируйте репозиторий:

   ```bash
   git clone <repository-url>
   cd DonWeather-ms-weather
   ```

2. Установите зависимости:

   ```bash
   go mod download
   ```

3. Запустите PostgreSQL с помощью Docker Compose:

   ```bash
   docker compose up -d
   ```

## Создание базы данных

Перед запуском микросервиса необходимо создать базу данных и выполнить миграции.

**1. Подключитесь к PostgreSQL:**

```bash
# Если PostgreSQL развернут в Kubernetes
kubectl exec -it <postgres-pod-name> -n <namespace> -- psql -U <db-user> -d postgres

# Или через port-forward
kubectl port-forward -n <namespace> svc/<postgres-service> 5432:5432
psql -h localhost -U <db-user> -d postgres
```

**2. Создайте базу данных:**

```sql
-- (рекомендуется) сначала создайте пользователя
CREATE USER weather WITH PASSWORD 'mypassword';

-- создайте базу и назначьте владельца
CREATE DATABASE weather OWNER weather;
```

**3. Выдайте права на схему `public` (нужно для миграций):**

```sql
-- подключитесь к базе
\c weather

-- важно: миграции создают таблицу public.schema_migrations, поэтому нужен CREATE/USAGE
GRANT USAGE, CREATE ON SCHEMA public TO weather;

-- (опционально) права на будущие таблицы/последовательности в public
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO weather;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO weather;
\q
```

**4. Выполните миграции:**

Миграции находятся в директории `migrations/`. Примените их с помощью инструмента миграций (например, golang-migrate):

```bash
# Пример с golang-migrate
migrate -path migrations -database "postgres://weather:mypassword@localhost:5432/weather?sslmode=disable" up
```

Или используйте psql напрямую:

```bash
psql -h <db-host> -U <db-user> -d weather -f migrations/000001_create_weather.up.sql
```

**5. Проверьте создание таблиц:**

```sql
\c weather
\dt
```

Должна быть создана таблица `weather` со следующей структурой:
- `id` - уникальный идентификатор (BIGINT, автоинкремент)
- `location_name` - название местоположения (TEXT)
- `last_updated` - время последнего обновления (TIMESTAMPTZ)
- `temp_c` - температура в градусах Цельсия (DOUBLE PRECISION)
- `humidity` - влажность (DOUBLE PRECISION)
- `pressure_mb` - давление в миллибарах (DOUBLE PRECISION)
- `wind_kph` - скорость ветра в км/ч (DOUBLE PRECISION)
- `condition_text` - текстовое описание погодных условий (TEXT)
- `created_at` - время создания записи (TIMESTAMPTZ, по умолчанию now())

## Запуск

1. Запустите миграции базы данных (если настроено автоматически).

2. Запустите сервер:

   ```bash
   go run cmd/main.go
   ```

Сервер будет доступен на `http://localhost:8080`.

## API

### Регистрация погодных данных

* **URL**: `/weather/register`
* **Метод**: POST
* **Описание**: Регистрирует новые погодные данные из внешнего API и сохраняет их в базу данных.

Пример запроса с JSON телом:

```bash
curl -X POST "http://localhost:8080/weather/register" \
  -H "Content-Type: application/json" \
  -d '{"q": "Amsterdam", "lang": "nl", "days": "3"}'
```

## Структура проекта

* `cmd/main.go` - Точка входа приложения
* `internal/delivery/http` - HTTP-обработчики
* `internal/usecase` - Бизнес-логика
* `internal/repository` - Репозиторий для работы с БД
* `migrations/` - Миграции базы данных

## Разработка

Для разработки используйте стандартные инструменты Go. Убедитесь, что база данных запущена перед запуском приложения.

## Управление секретами через Vault

Приложение использует External Secrets Operator для синхронизации секретов из HashiCorp Vault. Для работы необходимо создать секреты в Vault.

### Предварительные требования

1. Установлен и настроен HashiCorp Vault
2. Установлен External Secrets Operator в кластере Kubernetes
3. Создан ClusterSecretStore для подключения к Vault (см. `.helm/donweather-ms-weather/templates/clustersecretstore-vault-example.yaml`)

### Создание секретов в Vault (KV v2)

Микросервис использует следующие секреты из Vault:

1. **DB_PASSWORD** - пароль для подключения к PostgreSQL
2. **DB_USER** - пользователь для подключения к PostgreSQL
3. **DB_HOST** - хост базы данных PostgreSQL
4. **DB_PORT** - порт базы данных PostgreSQL
5. **DB_NAME** - имя базы данных PostgreSQL
6. **CORS_ORIGIN** - разрешенные источники для CORS
7. **WEATHER_API_KEY** - API ключ для Weather API

### Создание секретов через kubectl

Если Vault развернут в Kubernetes кластере, используйте kubectl для выполнения команд Vault CLI:

**1. Получите токен Vault:**

```bash
# Получите root token из Secret
VAULT_TOKEN=$(kubectl get secret vault-unseal-keys -n vault -o jsonpath='{.data.vault-root}' | base64 -d)
```

**2. Включите KV v2 секретный движок (если еще не включен):**

```bash
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault secrets enable -path=secret kv-v2
"
```

**3. Создайте секреты:**

```bash
# Создайте секрет для DB_PASSWORD
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/db-password \
  password='your-db-password'
"

# Создайте секрет для DB_USER
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/db-user \
  user='your-db-user'
"

# Создайте секрет для DB_HOST
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/db-host \
  host='postgres'
"

# Создайте секрет для DB_PORT
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/db-port \
  port='5432'
"

# Создайте секрет для DB_NAME
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/db-name \
  name='weather'
"

# Создайте секрет для CORS_ORIGIN
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/cors-origin \
  origin='*'
"

# Создайте секрет для WEATHER_API_KEY
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/weather-api-key \
  api-key='your-weather-api-key-here'
"
```

**Примечание:** Если ваш pod Vault имеет другое имя или находится в другом namespace, замените `vault-0` и `vault` на соответствующие значения.

### Проверка секретов

После создания секретов проверьте их наличие через kubectl:

```bash
# Получите токен Vault
VAULT_TOKEN=$(kubectl get secret vault-unseal-keys -n vault -o jsonpath='{.data.vault-root}' | base64 -d)

# Проверка секрета DB_PASSWORD
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/db-password
"

# Проверка секрета DB_USER
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/db-user
"

# Проверка секрета DB_HOST
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/db-host
"

# Проверка секрета DB_PORT
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/db-port
"

# Проверка секрета DB_NAME
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/db-name
"

# Проверка секрета CORS_ORIGIN
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/cors-origin
"

# Проверка секрета WEATHER_API_KEY
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/weather-api-key
"
```

### Настройка Helm Chart

После создания секретов в Vault, включите ExternalSecret в `values.yaml`:

```yaml
externalSecret:
  enabled: true
  secretStoreRef:
    name: "vault-backend"  # Имя вашего ClusterSecretStore
    kind: "ClusterSecretStore"
  secrets:
    dsn:
      key: "secret/data/donweather/dsn"
      property: "dsn"
    weatherApiKey:
      key: "secret/data/donweather/weather-api-key"
      property: "api-key"
    corsOrigin:
      key: "secret/data/donweather/cors-origin"
      property: "origin"
```

### Обновление секретов

Для обновления секретов в Vault используйте те же команды, что и для создания. External Secrets Operator автоматически синхронизирует изменения в соответствии с настройкой `refreshInterval` (по умолчанию 1 час).

## Лицензия

Проект используется под лицензией **BSL-1.0** (Business Source License). Подробности в файле `LICENSE`.
