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

## Лицензия

Проект используется под лицензией **BSL-1.0** (Business Source License). Подробности в файле `LICENSE`.
