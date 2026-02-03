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
2. **База данных** — создать БД и пользователя в PostgreSQL, выполнить миграции (аналогично п. 3 раздела «Установка (локальная)», с учётом хоста/порта для кластера).
3. **Application** — применить манифест из репозитория:
   ```bash
   export KUBECONFIG=$HOME/kubeconfig-services-cluster.yaml
   kubectl apply -f .argocd/application.yaml
   ```
4. **Проверка:**
   ```bash
   kubectl get application donweather-ms-weather-dev -n argocd
   export KUBECONFIG=$HOME/kubeconfig-dev-cluster.yaml
   kubectl get pods,svc -n donweather -l app.kubernetes.io/name=donweather-ms-weather
   kubectl get httproute -n donweather
   ```
   После синхронизации сервис доступен по адресу: **https://api.donweather.dev.buildbyte.ru** (при настроенных DNS и Gateway).

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

### Структура проекта

- `cmd/main.go` — точка входа
- `internal/delivery/http` — HTTP-обработчики
- `internal/usecase` — бизнес-логика
- `internal/repository` — работа с БД
- `migrations/` — миграции БД
- `.argocd/application.yaml` — манифест Argo CD Application
- `.helm/donweather-ms-weather/` — Helm chart

### Разработка

Перед запуском убедитесь, что PostgreSQL доступен и миграции применены.

### Лицензия

**BSL-1.0** (Business Source License). Подробности в файле `LICENSE`.
