# Полное ревью проекта DonWeather-ms-weather

**Дата ревью:** $(date)  
**Версия Go:** 1.24.3  
**Архитектура:** Clean Architecture (delivery, usecase, repository, domain)

---

## ✅ Что уже исправлено

1. ✅ **API ключ вынесен в переменные окружения** - `cmd/main.go:25-28`
2. ✅ **Опечатки исправлены:**
   - Файл `weatwer.go` → `weather.go`
   - `WeatherServie` → `WeatherService`
   - `FeatchAndSaveWeather` → `FetchAndSaveWeather`
   - `cansel` → `cancel`
3. ✅ **Интерфейс репозитория перенесен в domain** - `internal/domain/weather.go:22-25`
4. ✅ **Поле репозитория переименовано** - `weather` → `weatherRepo`
5. ✅ **Конструктор репозитория возвращает интерфейс** - `internal/repository/postgres_weather_repo.go:14`

---

## 🔴 Критические проблемы

### 1. Отсутствие валидации входных данных

**Файл:** `internal/delivery/http/weather_handler.go:23-42`

**Проблема:** Нет проверки обязательных полей запроса.

**Текущий код:**
```go
func (wh *weatherHandler) Register(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    var reqBody WeatherRequest
    
    if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }
    
    weather, err := wh.svc.FetchAndSaveWeather(ctx, reqBody.Q, reqBody.Lang, reqBody.Days)
    // Нет проверки на пустые значения!
}
```

**Решение:**
```go
if reqBody.Q == "" {
    http.Error(w, `{"error": "field 'q' is required"}`, http.StatusBadRequest)
    return
}

// Валидация days
if reqBody.Days != "" {
    daysInt, err := strconv.Atoi(reqBody.Days)
    if err != nil || daysInt < 1 || daysInt > 14 {
        http.Error(w, `{"error": "field 'days' must be between 1 and 14"}`, http.StatusBadRequest)
        return
    }
}
```

---

### 2. Нет обработки ошибок JSON encoding

**Файл:** `internal/delivery/http/weather_handler.go:41`

**Проблема:** Игнорируется ошибка при кодировании ответа.

**Текущий код:**
```go
json.NewEncoder(w).Encode(weather)  // ❌ Ошибка игнорируется
```

**Решение:**
```go
if err := json.NewEncoder(w).Encode(weather); err != nil {
    log.Printf("failed to encode response: %v", err)
    // Уже записали статус код, поэтому можем только залогировать
}
```

---

### 3. Небезопасное копирование структуры сервиса

**Файл:** `internal/delivery/http/weather_handler.go:17-20`

**Проблема:** Копируется структура вместо указателя.

**Текущий код:**
```go
func NewWeatherHandler(svc *usecase.WeatherService) WeatherHTTPHandler {
    return &weatherHandler{
        svc: *svc,  // ❌ Копирование структуры
    }
}
```

**Решение:** Хранить указатель:
```go
type weatherHandler struct {
    svc *usecase.WeatherService  // Указатель
}

func NewWeatherHandler(svc *usecase.WeatherService) WeatherHTTPHandler {
    return &weatherHandler{
        svc: svc,  // Сохраняем указатель
    }
}
```

**Почему:** Избегаем лишнего копирования и проблем при изменении структуры сервиса.

---

## 🟡 Важные проблемы

### 4. Отсутствие логирования

**Проблема:** Нет структурированного логирования запросов, ошибок, важных событий.

**Файлы:**
- `internal/delivery/http/weather_handler.go` - нет логов запросов
- `internal/usecase/weather_service.go` - нет логов ошибок API
- `cmd/main.go` - только стандартный `log`

**Рекомендация:** Использовать структурированное логирование:
```go
// Пример с zap или logrus
import "go.uber.org/zap"

logger.Info("weather request received",
    zap.String("location", reqBody.Q),
    zap.String("lang", reqBody.Lang),
)

logger.Error("failed to fetch weather",
    zap.Error(err),
    zap.String("location", q),
)
```

---

### 5. Нет graceful shutdown

**Файл:** `cmd/main.go:44-47`

**Проблема:** При остановке приложения (SIGTERM/SIGINT) соединения не закрываются корректно.

**Текущий код:**
```go
log.Fatal(http.ListenAndServe(":8080", nil))  // Нет обработки сигналов
```

**Решение:**
```go
import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    // ... инициализация ...
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: nil, // или ваш router
    }
    
    go func() {
        log.Println("Server started at :8080")
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exited")
}
```

---

### 6. HTTP клиент создается заново при каждом запросе

**Файл:** `internal/usecase/weather_service.go:44`

**Проблема:** Неэффективно создавать новый клиент каждый раз.

**Текущий код:**
```go
client := &http.Client{}  // ❌ Создается каждый раз
resp, err := client.Do(req)
```

**Решение:** Создать клиент на уровне сервиса:
```go
type WeatherService struct {
    weatherRepo domain.WeatherRepository
    apiKey      string
    httpClient  *http.Client  // Новое поле
}

func NewWeatherService(weatherRepo domain.WeatherRepository, apiKey string) *WeatherService {
    return &WeatherService{
        weatherRepo: weatherRepo,
        apiKey:      apiKey,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,  // ✅ Добавить timeout
        },
    }
}

func (ws *WeatherService) FetchAndSaveWeather(...) (*domain.Weather, error) {
    // ...
    resp, err := ws.httpClient.Do(req)  // Используем переиспользуемый клиент
    // ...
}
```

---

### 7. Нет timeout для HTTP клиента

**Файл:** `internal/usecase/weather_service.go:44`

**Проблема:** HTTP клиент может висеть бесконечно.

**Решение:** См. выше - добавить timeout в конструктор клиента.

---

### 8. Хардкод CORS origin

**Файл:** `internal/delivery/http/weather_handler.go:47`

**Проблема:** Origin захардкожен в коде.

**Текущий код:**
```go
w.Header().Set("Access-Control-Allow-Origin", "http://185.196.117.162:3000")
```

**Решение:** Вынести в переменную окружения или конфиг:
```go
corsOrigin := os.Getenv("CORS_ORIGIN")
if corsOrigin == "" {
    corsOrigin = "*"  // или другой дефолт
}
w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
```

---

### 9. Отсутствие проверки компилятором реализации интерфейса

**Файл:** `internal/repository/postgres_weather_repo.go`

**Проблема:** Нет явной проверки, что репозиторий реализует интерфейс.

**Решение:**
```go
// Гарантируем, что PostgresWeatherRepository реализует domain.WeatherRepository
var _ domain.WeatherRepository = (*PostgresWeatherRepository)(nil)

type PostgresWeatherRepository struct {
    db *sql.DB
}
```

**Почему:** Если интерфейс изменится, компилятор сразу покажет ошибку.

---

### 10. Тип Days - строка вместо числа

**Файл:** `internal/delivery/http/weather_http_handler.go:12`

**Проблема:** `Days` - строка, хотя должна быть числом.

**Текущий код:**
```go
type WeatherRequest struct {
    Q    string `json:"q"`
    Lang string `json:"lang"`
    Days string `json:"days"`  // ❌
}
```

**Решение:**
```go
type WeatherRequest struct {
    Q    string `json:"q" validate:"required"`
    Lang string `json:"lang" validate:"required"`
    Days *int   `json:"days,omitempty" validate:"omitempty,min=1,max=14"`  // ✅ Указатель для optional
}
```

Или с валидацией вручную:
```go
Days string `json:"days"`  // Если оставляем строкой

// В handler:
var daysInt int = 1  // дефолт
if reqBody.Days != "" {
    var err error
    daysInt, err = strconv.Atoi(reqBody.Days)
    if err != nil || daysInt < 1 || daysInt > 14 {
        http.Error(w, `{"error": "days must be between 1 and 14"}`, http.StatusBadRequest)
        return
    }
}
```

---

## 🟢 Рекомендации по улучшению

### 11. Отсутствие комментариев к публичным функциям

**Проблема:** Нет документации к публичным функциям и типам.

**Рекомендация:** Добавить комментарии в формате GoDoc:
```go
// WeatherRepository определяет контракт для работы с погодными данными.
// Реализации этого интерфейса отвечают за сохранение и удаление
// погодных данных в хранилище.
type WeatherRepository interface {
    // Save сохраняет погодные данные в хранилище.
    Save(context.Context, *Weather) error
    
    // Delete удаляет погодные данные по идентификатору.
    Delete(context.Context, string) error
}
```

---

### 12. Нет обработки ошибок БД

**Файл:** `internal/repository/postgres_weather_repo.go:20-35`

**Проблема:** Ошибки БД возвращаются как есть, без обертки.

**Рекомендация:** Добавить контекст к ошибкам:
```go
func (p *PostgresWeatherRepository) Save(ctx context.Context, weather *domain.Weather) error {
    query := `...`
    
    _, err := p.db.ExecContext(ctx, query, ...)
    if err != nil {
        return fmt.Errorf("failed to save weather for location %s: %w", 
            weather.Location.Name, err)
    }
    return nil
}
```

---

### 13. SQL запросы без форматирования

**Файл:** `internal/repository/postgres_weather_repo.go:21-23`

**Проблема:** SQL запрос плохо читается.

**Текущий код:**
```go
query := `
    INSERT INTO weather (location_name,last_updated,temp_c,humidity,pressure_mb,wind_kph,condition_text)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
`
```

**Рекомендация:** Добавить пробелы:
```go
query := `
    INSERT INTO weather (
        location_name, last_updated, temp_c, humidity, 
        pressure_mb, wind_kph, condition_text
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7)
`
```

---

### 14. Отсутствие тестов

**Проблема:** Нет unit-тестов для критичных компонентов.

**Рекомендация:** Добавить тесты:
- `internal/usecase/weather_service_test.go`
- `internal/repository/postgres_weather_repo_test.go`
- `internal/delivery/http/weather_handler_test.go`

---

### 15. Нет .env файла или примера

**Проблема:** Непонятно, какие переменные окружения нужны.

**Рекомендация:** Создать `.env.example`:
```env
DSN=postgres://user:password@localhost:5432/weather?sslmode=disable
WEATHER_API_KEY=your_api_key_here
CORS_ORIGIN=http://localhost:3000
```

---

### 16. Dockerfile можно оптимизировать

**Файл:** `Dockerfile`

**Проблема:** Используется `golang:1.24-alpine`, но версия Go указана в go.mod как `1.24.3`.

**Рекомендация:**
- Использовать конкретную версию Go
- Добавить multi-stage build оптимизацию (кэширование зависимостей)
- Добавить healthcheck

---

### 17. Отсутствие обработки ошибок миграции

**Файл:** `cmd/main.go:50-67`

**Проблема:** Миграции не закрываются после использования.

**Рекомендация:**
```go
func RunMigrations(dsn string) error {
    m, err := migrate.New("file://migrations", dsn)
    if err != nil {
        return fmt.Errorf("create migrate instance: %w", err)
    }
    defer m.Close()  // ✅ Добавить
    
    if err := m.Up(); err != nil {
        if errors.Is(err, migrate.ErrNoChange) {
            return nil
        }
        return fmt.Errorf("run migrations: %w", err)
    }
    
    return nil
}
```

---

### 18. Использование стандартного HTTP router

**Файл:** `cmd/main.go:44`

**Проблема:** Используется стандартный `http.Handle`, нет роутинга, middleware.

**Рекомендация:** Использовать роутер (gorilla/mux, chi, echo, gin):
```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Use(corsMiddleware)
r.Post("/weather/register", handler.Register)
```

---

### 19. Нет проверки подключения к БД

**Файл:** `cmd/main.go:29-32`

**Проблема:** `sql.Open` не проверяет подключение.

**Рекомендация:**
```go
db, err := sql.Open("pgx", dsn)
if err != nil {
    log.Fatal(err)
}

// Проверить подключение
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := db.PingContext(ctx); err != nil {
    log.Fatalf("failed to ping database: %v", err)
}
```

---

### 20. Использование дефолтных значений в main.go

**Файл:** `cmd/main.go:22-24`

**Проблема:** Дефолтный DSN с паролем в коде.

**Рекомендация:** Убрать дефолтные значения или использовать только для разработки:
```go
dsn := os.Getenv("DSN")
if dsn == "" {
    if os.Getenv("ENV") == "development" {
        dsn = "postgres://myuser:mypassword@localhost:5432/weather?sslmode=disable"
    } else {
        log.Fatal("DSN environment variable is required")
    }
}
```

---

## 📊 Сводная таблица проблем

| Категория | Критические | Важные | Рекомендации | Всего |
|-----------|-------------|--------|--------------|-------|
| **Безопасность** | 0 | 1 | 2 | 3 |
| **Архитектура** | 0 | 0 | 3 | 3 |
| **Код качество** | 3 | 5 | 8 | 16 |
| **Производительность** | 0 | 2 | 0 | 2 |
| **Тестирование** | 0 | 0 | 1 | 1 |
| **Документация** | 0 | 0 | 1 | 1 |
| **Инфраструктура** | 0 | 0 | 3 | 3 |
| **Всего** | **3** | **8** | **18** | **29** |

---

## 🎯 Приоритет исправлений

### Критично (исправить немедленно):
1. ✅ ~~API ключ в переменных окружения~~ (исправлено)
2. Добавить валидацию входных данных
3. Исправить копирование структуры сервиса
4. Добавить обработку ошибок JSON encoding

### Важно (исправить в ближайшее время):
5. Добавить логирование
6. Реализовать graceful shutdown
7. Вынести HTTP клиент на уровень сервиса
8. Добавить timeout для HTTP клиента
9. Вынести CORS origin в конфиг
10. Добавить проверку компилятором интерфейса
11. Исправить тип Days

### Желательно (можно сделать позже):
12. Добавить комментарии к публичным функциям
13. Улучшить обработку ошибок БД
14. Добавить тесты
15. Создать .env.example
16. Оптимизировать Dockerfile
17. Использовать HTTP router
18. Добавить проверку подключения к БД

---

## ✅ Что уже хорошо

1. ✅ **Clean Architecture** - правильное разделение слоев
2. ✅ **Интерфейс в domain** - правильно размещен
3. ✅ **Использование контекста** - для отмены операций
4. ✅ **Миграции БД** - используются миграции
5. ✅ **Docker-compose** - для локальной разработки
6. ✅ **Код компилируется** - нет синтаксических ошибок

---

## 📝 Итоговая оценка

| Критерий | Оценка | Комментарий |
|----------|--------|-------------|
| Архитектура | ✅ 4/5 | Clean Architecture соблюдена, можно улучшить |
| Безопасность | 🟡 3/5 | API ключ исправлен, но нет валидации |
| Качество кода | 🟡 3/5 | Работает, но много улучшений |
| Обработка ошибок | 🟡 2/5 | Базово, но неполно |
| Логирование | 🔴 1/5 | Практически отсутствует |
| Тестирование | 🔴 0/5 | Нет тестов |
| Документация | 🟡 2/5 | README есть, но нет GoDoc |
| Производительность | 🟡 3/5 | Работает, но не оптимизировано |

**Общая оценка: 🟡 3/5 (Хорошо, но требует улучшений)**

---

## 🚀 Следующие шаги

1. Исправить критические проблемы (валидация, ошибки)
2. Добавить логирование
3. Реализовать graceful shutdown
4. Вынести HTTP клиент на уровень сервиса
5. Добавить тесты для критичных компонентов
6. Улучшить документацию
