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

Приложение ожидает следующие секреты в Vault:

1. **DSN** (строка подключения к PostgreSQL)
2. **WEATHER_API_KEY** (API ключ для Weather API)
3. **CORS_ORIGIN** (разрешенные источники для CORS)

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
# Создайте секрет для DSN
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/dsn \
  dsn='postgres://user:password@postgres:5432/weather?sslmode=disable'
"

# Создайте секрет для Weather API Key
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/weather-api-key \
  api-key='your-weather-api-key-here'
"

# Создайте секрет для CORS Origin
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather/cors-origin \
  origin='*'
"
```

**Примечание:** Если ваш pod Vault имеет другое имя или находится в другом namespace, замените `vault-0` и `vault` на соответствующие значения.

### Проверка секретов

После создания секретов проверьте их наличие через kubectl:

```bash
# Получите токен Vault
VAULT_TOKEN=$(kubectl get secret vault-unseal-keys -n vault -o jsonpath='{.data.vault-root}' | base64 -d)

# Проверка секрета DSN
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/dsn
"

# Проверка секрета Weather API Key
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/weather-api-key
"

# Проверка секрета CORS Origin
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather/cors-origin
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
