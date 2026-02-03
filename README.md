# DonWeather-ms-weather

Микросервис для управления погодными данными в проекте DonWeather. API для регистрации и получения погодных данных, PostgreSQL, миграции. Clean Architecture: delivery, usecase, repository.

## Требования

- Go 1.19+
- Docker и Docker Compose (для локального PostgreSQL)
- PostgreSQL (или Docker Compose)
- API ключ WeatherAPI.com

## Установка (локальная)

1. Клонируйте репозиторий и перейдите в каталог:
   ```bash
   git clone <repository-url>
   cd DonWeather-ms-weather
   ```

2. Установите зависимости:
   ```bash
   go mod download
   ```

3. Запустите PostgreSQL:
   ```bash
   docker compose up -d
   ```

## Создание базы данных

Нужно для локального запуска и для развёртывания в кластер.

1. Подключитесь к PostgreSQL:
   ```bash
   # Kubernetes
   kubectl exec -it <postgres-pod> -n <namespace> -- psql -U <user> -d postgres
   # или port-forward
   kubectl port-forward -n <namespace> svc/<postgres-svc> 5432:5432
   psql -h localhost -U <user> -d postgres
   ```

2. Создайте пользователя и базу:
   ```sql
   CREATE USER weather WITH PASSWORD 'mypassword';
   CREATE DATABASE weather OWNER weather;
   ```

3. Выдайте права на схему `public` (для миграций):
   ```sql
   \c weather
   GRANT USAGE, CREATE ON SCHEMA public TO weather;
   ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO weather;
   ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO weather;
   \q
   ```

4. Выполните миграции:
   ```bash
   migrate -path migrations -database "postgres://weather:mypassword@localhost:5432/weather?sslmode=disable" up
   ```
   Или через psql:
   ```bash
   psql -h <host> -U <user> -d weather -f migrations/000001_create_weather.up.sql
   ```

5. Проверьте таблицы:
   ```sql
   \c weather
   \dt
   ```
   Должна быть таблица `weather` (id, location_name, last_updated, temp_c, humidity, pressure_mb, wind_kph, condition_text, created_at).

## Запуск (локально)

1. При необходимости выполните миграции (см. выше).
2. Запустите сервер:
   ```bash
   go run cmd/main.go
   ```
   Сервис: `http://localhost:8080`.

## API

- **POST** `/weather/register` — регистрация погодных данных из внешнего API, сохранение в БД.

Пример:
```bash
curl -X POST "http://localhost:8080/weather/register" \
  -H "Content-Type: application/json" \
  -d '{"q": "Amsterdam", "lang": "nl", "days": "3"}'
```

---

## Развёртывание в Dev кластер

Развёртывание через Argo CD. Application применяется **после** создания секретов в Vault и базы данных.

### Предусловия

- Dev кластер добавлен в Argo CD, создан AppProject `dev-microservices`.
- В Dev кластере: Gateway API (NGINX Gateway Fabric), cert-manager, Gateway `dev-gateway`, Vault Secrets Operator (VaultConnection/VaultAuth к Vault в Services).
- В namespace `donweather` есть секрет для Docker Registry (VaultStaticSecret `registry-docker-registry`).
- Vault в Services разблокирован, для Dev настроен Kubernetes auth (mount `kubernetes-dev`, роль `vault-secrets-operator`).

### Порядок развёртывания

1. **Секреты в Vault** — создать секрет `secret/donweather-ms-weather` с ключами (см. подраздел «Секреты в Vault» ниже).
2. **База данных** — создать БД и пользователя, выполнить миграции (раздел «Создание базы данных» выше).
3. **Application** — применить манифест из репозитория:
   ```bash
   export KUBECONFIG=$HOME/kubeconfig-services-cluster.yaml
   kubectl apply -f .argocd/application.yaml
   ```
   Либо указать полный путь к клонированному репозиторию: `kubectl apply -f /path/to/DonWeather-ms-weather/.argocd/application.yaml`.
4. **Проверка:**
   ```bash
   # Services кластер
   kubectl get application donweather-ms-weather-dev -n argocd
   kubectl get application donweather-ms-weather-dev -n argocd -o jsonpath='{.status.sync.status}' && echo
   kubectl get application donweather-ms-weather-dev -n argocd -o jsonpath='{.status.health.status}' && echo

   # Dev кластер
   export KUBECONFIG=$HOME/kubeconfig-dev-cluster.yaml
   kubectl get pods,svc -n donweather -l app.kubernetes.io/name=donweather-ms-weather
   kubectl get httproute -n donweather
   ```
   После синхронизации сервис доступен по адресу: **https://api.donweather.dev.buildbyte.ru** (при настроенных DNS и Gateway).

### Секреты в Vault

Секрет в Vault (KV v2), путь: **`secret/donweather-ms-weather`**. Vault Secrets Operator синхронизирует его в Kubernetes. VaultStaticSecret создаётся Helm chart при установке (`vaultSecretOperator.enabled: true`), отдельно применять манифест не нужно.

**Обязательные ключи:**

| Ключ в Vault      | Переменная в поде |
|-------------------|-------------------|
| `db-password`     | `DB_PASSWORD`      |
| `db-user`         | `DB_USER`          |
| `db-host`         | `DB_HOST`          |
| `db-port`         | `DB_PORT`          |
| `db-name`         | `DB_NAME`          |
| `weather-api-key`  | `WEATHER_API_KEY`  |
| `cors-origin`     | `CORS_ORIGIN`      |

**Предусловия:** Vault в Services разблокирован; Vault Secrets Operator в Dev настроен; Kubernetes auth для Dev настроен; есть root token или токен с записью в `secret/`.

**Шаги:**

1. Подготовка переменных:
   ```bash
   export KUBECONFIG=$HOME/kubeconfig-services-cluster.yaml
   export VAULT_ADDR="http://127.0.0.1:8200"
   export VAULT_TOKEN=$(cat /tmp/vault-root-token.txt)
   ```

2. Включить KV v2 по пути `secret` (если ещё не включён):
   ```bash
   kubectl exec -it vault-0 -n vault -- sh -c "
   export VAULT_ADDR='http://127.0.0.1:8200'
   export VAULT_TOKEN='$VAULT_TOKEN'
   vault secrets enable -version=2 -path=secret kv 2>&1 || echo 'Уже включен'
   "
   ```

3. Создать секрет (подставьте свои значения; для паролей используйте одинарные кавычки):
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

4. Проверить запись:
   ```bash
   kubectl exec -it vault-0 -n vault -- sh -c "
   export VAULT_ADDR='http://127.0.0.1:8200'
   export VAULT_TOKEN='$VAULT_TOKEN'
   vault kv get secret/donweather-ms-weather
   "
   ```
   Должны отображаться все семь ключей.

5. Обновление секрета: снова выполнить `vault kv put secret/donweather-ms-weather ...`. После обновления при необходимости перезапустить поды:
   ```bash
   kubectl rollout restart deployment donweather-ms-weather -n donweather
   ```

Политика доступа: для Dev обычно используется политика `vault-secrets-operator-dev-policy` (чтение `secret/data/*`, `secret/metadata/*`). Отдельная политика для `donweather-ms-weather` не требуется.

**Чек-лист:** Vault разблокирован, KV v2 включён → секрет создан и проверен → приложение развёрнуто с `vaultSecretOperator.enabled: true` → VaultStaticSecret в статусе Synced, Secret в namespace создан.

---

## Структура проекта

- `cmd/main.go` — точка входа
- `internal/delivery/http` — HTTP-обработчики
- `internal/usecase` — бизнес-логика
- `internal/repository` — работа с БД
- `migrations/` — миграции БД
- `.argocd/application.yaml` — манифест Argo CD Application
- `.helm/donweather-ms-weather/` — Helm chart

## Разработка

Используйте стандартные инструменты Go. Перед запуском убедитесь, что PostgreSQL доступен и миграции применены.

## Лицензия

**BSL-1.0** (Business Source License). Подробности в файле `LICENSE`.
