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

## Заведение секретов в Vault

Инструкция по созданию и обновлению секретов микросервиса в HashiCorp Vault. Основана на [DonInfrastructure README](https://github.com/opusdvs/DonInfrastructure/blob/main/README.md). Приложение использует **Vault Secrets Operator** (VSO) для синхронизации секретов из Vault в Kubernetes.

### Предварительные условия

- Vault развёрнут в **Services кластере** и разблокирован (unsealed).
- Vault Secrets Operator в **Dev кластере** настроен и подключён к Vault по адресу `https://vault.buildbyte.ru` (раздел 8 DonInfrastructure).
- Kubernetes auth в Vault для Dev кластера настроен (mount `kubernetes-dev`, роль `vault-secrets-operator`).
- У вас есть root token Vault или токен с правами на запись в `secret/`.

### Структура секрета в Vault

Микросервис использует **один** секрет в Vault (KV v2) по пути **`donweather`** (mount `secret`).  
Vault Secrets Operator синхронизирует все ключи из этого пути в Kubernetes Secret с тем же набором ключей.

**Обязательные ключи в секрете Vault:**

| Ключ в Vault      | Назначение                          | Переменная в поде   |
|-------------------|-------------------------------------|---------------------|
| `db-password`     | Пароль пользователя БД              | `DB_PASSWORD`       |
| `db-user`         | Имя пользователя БД                 | `DB_USER`           |
| `db-host`         | Хост PostgreSQL                     | `DB_HOST`           |
| `db-port`         | Порт PostgreSQL (например `5432`)   | `DB_PORT`           |
| `db-name`         | Имя базы данных                     | `DB_NAME`           |
| `weather-api-key` | API-ключ внешнего сервиса погоды    | `WEATHER_API_KEY`   |
| `cors-origin`     | Разрешённый CORS origin (например `*`) | `CORS_ORIGIN`   |

Имена ключей в Vault должны совпадать с указанными — их ожидает Helm chart и Deployment.

### 1. Подготовка переменных

```bash
# Переключиться на Services кластер (здесь установлен Vault)
export KUBECONFIG=$HOME/kubeconfig-services-cluster.yaml

# Токен Vault (root или с правами на запись в secret/)
export VAULT_ADDR="http://127.0.0.1:8200"
export VAULT_TOKEN=$(cat /tmp/vault-root-token.txt)
# Если root token в другом месте — укажите свой путь
```

### 2. Проверка KV v2

Убедитесь, что секретный движок KV v2 включён по пути `secret`:

```bash
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault secrets enable -version=2 -path=secret kv 2>&1 || echo 'Секретный движок уже включен'
"
```

### 3. Создание секрета donweather

Подставьте свои значения вместо плейсхолдеров. Для паролей и токенов используйте одинарные кавычки, чтобы избежать интерпретации спецсимволов shell.

```bash
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather \
  db-password='<ПАРОЛЬ_ПОЛЬЗОВАТЕЛЯ_БД>' \
  db-user='<ИМЯ_ПОЛЬЗОВАТЕЛЯ_БД>' \
  db-host='<ХОСТ_POSTGRESQL>' \
  db-port='5432' \
  db-name='weather' \
  weather-api-key='<API_КЛЮЧ_ПОГОДЫ>' \
  cors-origin='*'
"
```

**Пример для dev (PostgreSQL в Services по LoadBalancer):**

```bash
# Предполагается, что внешний IP PostgreSQL уже известен
POSTGRES_HOST='<ВНЕШНИЙ_IP_POSTGRESQL>'  # из kubectl get svc postgresql-external -n postgresql

kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather \
  db-password='<ПАРОЛЬ>' \
  db-user='myuser' \
  db-host='$POSTGRES_HOST' \
  db-port='5432' \
  db-name='weather' \
  weather-api-key='<ВАШ_WEATHER_API_KEY>' \
  cors-origin='https://api.donweather.dev.buildbyte.ru'
"
```

### 4. Проверка записи

```bash
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get secret/donweather
"
```

В выводе должны быть все семь ключей (значения показываются в открытом виде — не логируйте вывод с реальными паролями).

Проверка в формате JSON (для отладки):

```bash
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv get -format=json secret/donweather | jq '.data.data | keys'
"
```

Должен вывести список ключей: `["cors-origin", "db-host", "db-name", "db-password", "db-port", "db-user", "weather-api-key"]`.

### 5. Синхронизация в Kubernetes (Dev кластер)

Секрет в Kubernetes создаётся **Vault Secrets Operator** на основе ресурса `VaultStaticSecret` из Helm chart. После записи в Vault:

1. Убедитесь, что приложение развёрнуто в Dev кластере с включённым `vaultSecretOperator.enabled: true` и корректным `vaultAuthRef` (например, `vault-secrets-operator/default` или `vault-auth` в namespace приложения).
2. Оператор периодически синхронизирует секрет (по умолчанию `refreshAfter: 1h`). Чтобы не ждать:
   - можно перезапустить поды Vault Secrets Operator в namespace `vault-secrets-operator`, или
   - временно изменить `refreshAfter` в values и пересинхронизировать приложение через Argo CD / Helm.

Проверка в Dev кластере:

```bash
export KUBECONFIG=$HOME/kubeconfig-dev-cluster.yaml

# Статус VaultStaticSecret (namespace — тот, куда ставите чарт, например donweather)
kubectl get vaultstaticsecret -n donweather
kubectl describe vaultstaticsecret donweather-ms-weather-vault-secret -n donweather

# Имя итогового Secret задаётся в chart: <release>-secret
kubectl get secret donweather-ms-weather-secret -n donweather
kubectl get secret donweather-ms-weather-secret -n donweather -o jsonpath='{.data}' | jq 'keys'
```

### 6. Обновление секрета

При изменении данных в Vault достаточно перезаписать путь `secret/donweather`:

```bash
# Те же переменные VAULT_ADDR и VAULT_TOKEN
kubectl exec -it vault-0 -n vault -- sh -c "
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_TOKEN='$VAULT_TOKEN'
vault kv put secret/donweather \
  db-password='<НОВЫЙ_ПАРОЛЬ>' \
  db-user='myuser' \
  db-host='<ХОСТ>' \
  db-port='5432' \
  db-name='weather' \
  weather-api-key='<КЛЮЧ>' \
  cors-origin='*'
"
```

После обновления Vault оператор подтянет новые значения в течение интервала обновления. Чтобы поды приложения получили новые переменные окружения, перезапустите Deployment:

```bash
kubectl rollout restart deployment donweather-ms-weather -n donweather
```

### 7. Политика доступа Vault (Dev кластер)

По инструкции DonInfrastructure для Dev кластера используется политика `vault-secrets-operator-dev-policy`, разрешающая чтение `secret/data/*` и `secret/metadata/*`. Путь `secret/donweather` попадает под эту политику — отдельная политика для donweather-ms-weather не требуется.

Если вы настраиваете отдельную политику (например, только для пути donweather), минимальный фрагмент:

```hcl
path "secret/data/donweather" {
  capabilities = ["read"]
}
path "secret/metadata/donweather" {
  capabilities = ["read", "list"]
}
```

### Краткий чек-лист

- [ ] Vault разблокирован, KV v2 включён по пути `secret`
- [ ] Секрет создан: `vault kv put secret/donweather` с ключами `db-password`, `db-user`, `db-host`, `db-port`, `db-name`, `weather-api-key`, `cors-origin`
- [ ] Проверка: `vault kv get secret/donweather`
- [ ] В Dev кластере развёрнут Helm chart с `vaultSecretOperator.enabled: true` и правильным `vaultAuthRef`
- [ ] VaultStaticSecret в статусе Synced, Secret создан и поды читают актуальные переменные

### Ссылки

- [DonInfrastructure README](https://github.com/opusdvs/DonInfrastructure/blob/main/README.md) — общий порядок установки, Vault, VSO, Dev кластер
- Раздел 8 DonInfrastructure — установка и настройка Vault Secrets Operator в Dev кластере, VaultConnection и VaultAuth
- Раздел 8.1 DonInfrastructure — Kubernetes Auth в Vault для VSO (Services)
- Разделы 8.2–8.4 DonInfrastructure — Kubernetes Auth для Dev кластера и создание VaultConnection/VaultAuth

## Лицензия

Проект используется под лицензией **BSL-1.0** (Business Source License). Подробности в файле `LICENSE`.
