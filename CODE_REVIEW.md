# Code Review - DonWeather-ms-weather

## 🔴 Критические проблемы

### 1. Безопасность: API ключ в коде
**Файл:** `internal/usecase/weather_service.go:37`
```go
query.Set("key", "cd3200bf0f914528862150404260801") // ❌ ЗАХАРДКОЖЕН!
```

**Проблема:** API ключ должен быть в переменных окружения.

**Решение:**
```go
apiKey := os.Getenv("WEATHER_API_KEY")
if apiKey == "" {
    return nil, fmt.Errorf("WEATHER_API_KEY environment variable is required")
}
```

### 2. Опечатки в именах
- **Файл:** `internal/domain/weatwer.go` → должно быть `weather.go`
- **Тип:** `WeatherServie` → должно быть `WeatherService` (везде)
- **Метод:** `FeatchAndSaveWeather` → должно быть `FetchAndSaveWeather`
- **Переменная:** `cansel` → должно быть `cancel` (weather_handler.go:24)

### 3. Отсутствие валидации входных данных
**Файл:** `internal/delivery/http/weather_handler.go`

**Проблема:** Нет проверки на пустые значения обязательных полей.

**Решение:** Добавить валидацию:
```go
if reqBody.Q == "" {
    http.Error(w, "field 'q' is required", http.StatusBadRequest)
    return
}
```

---

## 🟡 Важные проблемы

### 4. Отсутствие логирования
**Проблема:** Нет логирования ошибок, запросов, важных событий.

**Рекомендация:** Использовать структурированное логирование (например, `logrus` или `zap`).

### 5. Нет graceful shutdown
**Файл:** `cmd/main.go`

**Проблема:** При остановке приложения соединения с БД не закрываются корректно.

**Решение:** Добавить обработку сигналов (SIGTERM, SIGINT).

### 6. HTTP клиент создается заново
**Файл:** `internal/usecase/weather_service.go:47`

**Проблема:** Каждый раз создается новый `http.Client`, нет переиспользования.

**Решение:** Создать один клиент на уровне сервиса или передавать как зависимость.

### 7. Хардкод в CORS
**Файл:** `internal/delivery/http/weather_handler.go:47`

**Проблема:** Origin захардкожен в коде.

**Решение:** Вынести в переменную окружения или конфиг.

---

## 🟢 Рекомендации по улучшению

### 8. Типизация данных
**Файл:** `internal/delivery/http/weather_http_handler.go`

**Проблема:** `Days` - строка, но должна быть числом.

**Текущий код:**
```go
type WeatherRequest struct {
    Q    string `json:"q"`
    Lang string `json:"lang"`
    Days string `json:"days"` // ❌
}
```

**Рекомендация:**
```go
type WeatherRequest struct {
    Q    string `json:"q" validate:"required"`
    Lang string `json:"lang" validate:"required"`
    Days int    `json:"days" validate:"min=1,max=14"`
}
```

### 9. Улучшение обработки ошибок
**Файл:** `internal/usecase/weather_service.go:56-59`

**Проблема:** Не различаются типы ошибок (4xx vs 5xx).

**Рекомендация:** Создать кастомные типы ошибок для лучшей обработки.

### 10. Отсутствие интерфейса для репозитория
**Файл:** `internal/repository/postgres_weather_repo.go`

**Проблема:** Прямое использование конкретного типа вместо интерфейса.

**Рекомендация:** Создать интерфейс `WeatherRepository` в domain слое.

### 11. Нет timeout для HTTP клиента
**Файл:** `internal/usecase/weather_service.go:47`

**Проблема:** HTTP клиент не имеет timeout.

**Рекомендация:**
```go
client := &http.Client{
    Timeout: 10 * time.Second,
}
```

### 12. Нет проверки на nil
**Файл:** `cmd/main.go:31`

**Проблема:** `db.Close()` в defer может быть излишним, но нет проверки на успешное открытие перед использованием.

### 13. Контекст не используется полностью
**Файл:** `internal/usecase/weather_service.go`

**Проблема:** Контекст передается, но не проверяется на отмену перед сохранением в БД.

---

## 📝 Замечания по стилю кода

1. **Консистентность именования:** Смешаны стили (`NewWeatherHandler` vs `NewWeatherServie`)
2. **Комментарии:** Отсутствуют комментарии к публичным функциям и типам
3. **Форматирование SQL:** В запросах нет пробелов после запятых (строка 22 в `postgres_weather_repo.go`)

---

## ✅ Что хорошо

1. ✅ Используется Clean Architecture
2. ✅ Разделение на слои (delivery, usecase, repository, domain)
3. ✅ Использование миграций БД
4. ✅ Docker-compose для локальной разработки
5. ✅ Использование контекста для отмены операций
6. ✅ Использование prepared statements (через `ExecContext`)

---

## 🎯 Приоритет исправлений

1. **Срочно:**
   - Вынести API ключ в переменные окружения
   - Исправить опечатки в именах

2. **Важно:**
   - Добавить валидацию входных данных
   - Добавить логирование
   - Добавить graceful shutdown

3. **Желательно:**
   - Улучшить обработку ошибок
   - Вынести HTTP клиент на уровень сервиса
   - Вынести CORS origin в конфиг
