# DonWeather-ms-weather

Микросервис погодных данных для проекта DonWeather: API регистрации и получения данных, PostgreSQL, миграции. Clean Architecture: delivery, usecase, repository.

---

## Установка (локальная)

### Требования

- Go 1.19+
- Docker и Docker Compose
- PostgreSQL (или через Docker Compose)
- API ключ WeatherAPI.com

### 1. Клонирование и зависимости

```bash
git clone <repository-url>
cd DonWeather-ms-weather
go mod download
```

### 2. Запуск PostgreSQL

```bash
docker compose up -d
```

### 3. База данных

Подключитесь к PostgreSQL (локально или через `kubectl port-forward`), затем:

```sql
CREATE USER weather WITH PASSWORD 'mypassword';
CREATE DATABASE weather OWNER weather;
\c weather
GRANT USAGE, CREATE ON SCHEMA public TO weather;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO weather;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO weather;
\q
```

Миграции:

```bash
migrate -path migrations -database "postgres://weather:mypassword@localhost:5432/weather?sslmode=disable" up
```

Или через psql:

```bash
psql -h localhost -U weather -d weather -f migrations/000001_create_weather.up.sql
```

### 4. Запуск

```bash
go run cmd/main.go
```

Сервис: `http://localhost:8080`.

### 5. API

**POST** `/weather/register` — регистрация погодных данных из внешнего API.

```bash
curl -X POST "http://localhost:8080/weather/register" \
  -H "Content-Type: application/json" \
  -d '{"q": "Amsterdam", "lang": "nl", "days": "3"}'
```

---

## Развёртывание в Dev кластер

Развёртывание через Argo CD. Манифест Application применяется **после** создания секретов в Vault и базы данных.

### Порядок развёртывания

1. **Секреты в Vault** — создать секрет `secret/donweather-ms-weather` (см. подраздел «Секреты в Vault» ниже).
2. **База данных** — создать БД и пользователя, выполнить миграции (см. подраздел «База данных» ниже).
3. **Application** — применить манифест из репозитория:
   ```bash
   export KUBECONFIG=$HOME/kubeconfig-services-cluster.yaml
   kubectl apply -f .argocd/application.yaml
   ```

### Секреты в Vault

Путь в Vault (KV v2): **`secret/donweather-ms-weather`**. VaultStaticSecret создаётся Helm chart при установке (`vaultSecretOperator.enabled: true`).

**Ключи:** `db-password`, `db-user`, `db-host`, `db-port`, `db-name`, `weather-api-key`, `cors-origin` (в поде: `DB_PASSWORD`, `DB_USER`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `WEATHER_API_KEY`, `CORS_ORIGIN`).

**Шаги:**

1. Переменные (Services кластер):
   ```bash
   export KUBECONFIG=$HOME/kubeconfig-services-cluster.yaml
   export VAULT_ADDR="http://127.0.0.1:8200"
   export VAULT_TOKEN=$(cat /tmp/vault-root-token.txt)
   ```

2. Включить KV v2 по пути `secret` (если нужно):
   ```bash
   kubectl exec -it vault-0 -n vault -- sh -c "
   export VAULT_ADDR='http://127.0.0.1:8200'
   export VAULT_TOKEN='$VAULT_TOKEN'
   vault secrets enable -version=2 -path=secret kv 2>&1 || echo 'Уже включен'
   "
   ```

3. Создать секрет (подставить свои значения; пароли в одинарных кавычках):
   ```bash
   kubectl exec -it vault-0 -n vault -- sh -c "
   export VAULT_ADDR='http://127.0.0.1:8200'
   export VAULT_TOKEN='$VAULT_TOKEN'
   vault kv put secret/donweather-ms-weather \
     db-password='<ПАРОЛЬ_БД>' \
     db-user='<USER_БД>' \
     db-host='<ХОСТ_POSTGRESQL>' \
     db-port='5432' \
     db-name='weather' \
     weather-api-key='<API_КЛЮЧ>' \
     cors-origin='*'
   "
   ```

4. Проверить: `vault kv get secret/donweather-ms-weather` (в том же `kubectl exec` с `VAULT_ADDR` и `VAULT_TOKEN`).

5. Обновление — снова `vault kv put secret/donweather-ms-weather ...`. После смены секрета при необходимости: `kubectl rollout restart deployment donweather-ms-weather -n donweather`.

### База данных

PostgreSQL для Dev может быть в Services кластере (доступ по LoadBalancer или port-forward). Пользователь, пароль и имя БД должны совпадать с теми, что записаны в секрете Vault (`db-user`, `db-password`, `db-name`).

1. Подключиться к PostgreSQL (подставьте хост/порт из секрета или из окружения кластера):
   ```bash
   # Например, через port-forward к сервису в Services кластере
   kubectl port-forward -n postgresql svc/postgresql 5432:5432
   psql -h localhost -U postgres -d postgres
   ```

2. Создать пользователя и базу (имя пользователя и пароль — как в ключах `db-user` и `db-password` в Vault; имя базы — как `db-name`, обычно `weather`):
   ```sql
   CREATE USER weather WITH PASSWORD 'ваш_пароль';
   CREATE DATABASE weather OWNER weather;
   \c weather
   GRANT USAGE, CREATE ON SCHEMA public TO weather;
   ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO weather;
   ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO weather;
   \q
   ```

3. Выполнить миграции (хост, порт, пользователь, пароль, база — из секрета Vault):
   ```bash
   migrate -path migrations -database "postgres://<db-user>:<db-password>@<db-host>:<db-port>/weather?sslmode=disable" up
   ```
   Или через psql:
   ```bash
   psql -h <db-host> -p <db-port> -U <db-user> -d weather -f migrations/000001_create_weather.up.sql
   ```
   Пароль можно передать переменной: `PGPASSWORD='<пароль>' psql ...`.

### Application

Применить манифест (из корня клонированного репозитория):

```bash
kubectl apply -f .argocd/application.yaml
```

### Структура проекта

- `cmd/main.go` — точка входа
- `internal/delivery/http` — HTTP-обработчики
- `internal/usecase` — бизнес-логика
- `internal/repository` — работа с БД
- `migrations/` — миграции БД
- `.argocd/application.yaml` — манифест Argo CD Application
- `.helm/donweather-ms-weather/` — Helm chart

### Лицензия

**BSL-1.0** (Business Source License). Подробности в файле `LICENSE`.
