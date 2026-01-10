# Ревью проекта DonWeather-ms-weather (Текущее состояние)

**Дата:** $(date)  
**Статус:** После частичных исправлений

---

## ✅ Что исправлено

1. ✅ **Копирование структуры сервиса** - теперь используется указатель (`weather_handler.go:16`)
2. ✅ **Валидация входных данных** - добавлена функция `ValidateWeatherRequest`
3. ✅ **Обработка ошибок JSON encoding** - добавлена проверка ошибок (`weather_handler.go:43-46`)
4. ✅ **CORS origin** - вынесен в переменные окружения (`weather_handler.go:51-54`)
5. ✅ **HTTP клиент** - вынесен на уровень сервиса (`weather_service.go:18,25`)
6. ✅ **Тип Days** - изменен с `string` на `int` (`weather_http_handler.go:12`)

---

## 🔴 Критические проблемы

### 1. Валидация вызывается до декодирования JSON

**Файл:** `internal/delivery/http/weather_handler.go:28-33`

**Проблема:** Валидация вызывается до того, как JSON декодирован в структуру.

**Текущий код:**
```go
func (wh *weatherHandler) Register(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    var reqBody WeatherRequest

    if err := ValidateWeatherRequest(reqBody); err != nil {  // ❌ reqBody еще пустой!
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Декодирование JSON происходит ПОСЛЕ валидации, но валидация уже выполнена!
```

**Решение:** Переместить валидацию после декодирования:
```go
func (wh *weatherHandler) Register(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    var reqBody WeatherRequest

    if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
        http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
        return
    }

    if err := ValidateWeatherRequest(reqBody); err != nil {  // ✅ После декодирования
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    weather, err := wh.svc.FetchAndSaveWeather(ctx, reqBody.Q, reqBody.Lang, reqBody.Days)
    // ...
}
```

**Строка 30:** `if err := ValidateWeatherRequest(reqBody);` - валидация пустой структуры

---

### 2. Неправильный статус код при ошибке encoding

**Файл:** `internal/delivery/http/weather_handler.go:41-46`

**Проблема:** После `WriteHeader(http.StatusCreated)` нельзя изменить статус код.

**Текущий код:**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)  // ✅ Записали 201
if err := json.NewEncoder(w).Encode(weather); err != nil {
    http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)  // ❌ Не сработает
    return
}
```

**Решение:** Проверить ошибку перед записью статуса:
```go
w.Header().Set("Content-Type", "application/json")
if err := json.NewEncoder(w).Encode(weather); err != nil {
    http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
    return
}
w.WriteHeader(http.StatusCreated)  // ✅ После успешного encoding
```

Или не использовать `WriteHeader` вообще (будет автоматически 200):
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
if err := json.NewEncoder(w).Encode(weather); err != nil {
    // Уже записали статус, можем только залогировать
    log.Printf("failed to encode response after status written: %v", err)
    return
}
```

---

### 3. Валидация Days проверяет неинициализированное значение

**Файл:** `internal/delivery/http/weather_handler.go:77-79`

**Проблема:** Если `Days` не передан в JSON, будет 0, и валидация вернет ошибку, хотя это может быть валидным дефолтным значением.

**Текущий код:**
```go
func ValidateWeatherRequest(reqBody WeatherRequest) error {
    if reqBody.Q == "" {
        return errors.New("field 'q' is required")
    }
    if reqBody.Lang == "" {
        return errors.New("field 'lang' is required")
    }
    if reqBody.Days < 1 || reqBody.Days > 14 {  // ❌ 0 не пройдет валидацию
        return errors.New("field 'days' must be between 1 and 14")
    }
    return nil
}
```

**Решение:** Использовать указатель для optional поля или установить дефолт:
```go
type WeatherRequest struct {
    Q    string `json:"q"`
    Lang string `json:"lang"`
    Days *int   `json:"days,omitempty"`  // Указатель для optional
}

func ValidateWeatherRequest(reqBody WeatherRequest) error {
    if reqBody.Q == "" {
        return errors.New("field 'q' is required")
    }
    if reqBody.Lang == "" {
        return errors.New("field 'lang' is required")
    }
    if reqBody.Days != nil && (*reqBody.Days < 1 || *reqBody.Days > 14) {
        return errors.New("field 'days' must be between 1 and 14")
    }
    return nil
}

// В handler после валидации:
days := 1  // дефолт
if reqBody.Days != nil {
    days = *reqBody.Days
}
weather, err := wh.svc.FetchAndSaveWeather(ctx, reqBody.Q, reqBody.Lang, days)
```

Или установить дефолт перед валидацией:
```go
// После декодирования, перед валидацией:
if reqBody.Days == 0 {
    reqBody.Days = 1  // дефолтное значение
}
if err := ValidateWeatherRequest(reqBody); err != nil {
    // ...
}
```

---

## 🟡 Важные проблемы

### 4. Нет timeout для HTTP клиента

**Файл:** `internal/usecase/weather_service.go:25`

**Проблема:** HTTP клиент не имеет timeout, может висеть бесконечно.

**Текущий код:**
```go
httpClient: &http.Client{},  // ❌ Нет timeout
```

**Решение:**
```go
import "time"

func NewWeatherService(weatherRepo domain.WeatherRepository, apiKey string) *WeatherService {
    return &WeatherService{
        weatherRepo: weatherRepo,
        apiKey:      apiKey,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,  // ✅ Добавить timeout
        },
    }
}
```

---

### 5. Отсутствие логирования

**Проблема:** Нет логирования запросов, ошибок, важных событий.

**Файлы:**
- `internal/delivery/http/weather_handler.go` - нет логов запросов
- `internal/usecase/weather_service.go` - нет логов ошибок API
- `cmd/main.go` - только стандартный `log`

**Рекомендация:** Добавить структурированное логирование (zap, logrus, slog):
```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

// В handler:
logger.Info("weather request received",
    "location", reqBody.Q,
    "lang", reqBody.Lang,
    "days", reqBody.Days,
)

// В usecase при ошибках:
logger.Error("failed to fetch weather",
    "error", err,
    "location", q,
)
```

---

### 6. Нет graceful shutdown

**Файл:** `cmd/main.go:44-47`

**Проблема:** При остановке приложения соединения не закрываются корректно.

**Текущий код:**
```go
log.Fatal(http.ListenAndServe(":8080", nil))  // Нет обработки сигналов
```

**Решение:** См. FULL_CODE_REVIEW.md раздел 5.

---

### 7. Нет проверки компилятором реализации интерфейса

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

---

### 8. Неоптимальная валидация Days

**Файл:** `internal/delivery/http/weather_handler.go:77-79`

**Проблема:** Валидация не учитывает, что Days может быть optional.

**Решение:** См. пункт 3 выше.

---

### 9. Ошибка может быть записана в response дважды

**Файл:** `internal/delivery/http/weather_handler.go:43-46`

**Проблема:** После `WriteHeader` нельзя вызвать `http.Error` корректно.

**Решение:** Исправить порядок операций (см. пункт 2).

---

### 10. Нет закрытия Body при ошибках

**Файл:** `internal/usecase/weather_service.go:54-56`

**Проблема:** При ошибке статус кода тело ответа не читается полностью.

**Текущий код:**
```go
if resp.StatusCode != http.StatusOK {
    body, _ := io.ReadAll(resp.Body)  // ✅ Читаем, но defer закроет
    return nil, fmt.Errorf("weather API error: %s", string(body))
}
```

**Решение:** Уже есть `defer resp.Body.Close()` на строке 52, это правильно. Но можно улучшить обработку ошибок.

---

## 🟢 Рекомендации по улучшению

### 11. Улучшить сообщения об ошибках

**Файл:** `internal/delivery/http/weather_handler.go:37`

**Текущий код:**
```go
http.Error(w, fmt.Sprintf("failed to fetch/save weather: %v", err), http.StatusInternalServerError)
```

**Рекомендация:** Различать типы ошибок:
```go
if err != nil {
    // Различать типы ошибок для разных статус кодов
    if strings.Contains(err.Error(), "weather API error") {
        http.Error(w, fmt.Sprintf("external API error: %v", err), http.StatusBadGateway)
    } else if strings.Contains(err.Error(), "context deadline exceeded") {
        http.Error(w, "request timeout", http.StatusRequestTimeout)
    } else {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        log.Printf("internal error: %v", err)  // Логировать детали
    }
    return
}
```

---

### 12. Добавить комментарии к публичным функциям

**Проблема:** Нет GoDoc комментариев.

**Рекомендация:**
```go
// ValidateWeatherRequest проверяет корректность запроса на получение погоды.
// Возвращает ошибку, если обязательные поля отсутствуют или имеют недопустимые значения.
func ValidateWeatherRequest(reqBody WeatherRequest) error {
    // ...
}
```

---

### 13. Улучшить обработку ошибок БД

**Файл:** `internal/repository/postgres_weather_repo.go:35`

**Рекомендация:**
```go
_, err := p.db.ExecContext(ctx, query, ...)
if err != nil {
    return fmt.Errorf("failed to save weather for location %s: %w", 
        weather.Location.Name, err)
}
return nil
```

---

### 14. SQL запросы без пробелов

**Файл:** `internal/repository/postgres_weather_repo.go:21-23`

**Рекомендация:** Добавить пробелы для читаемости:
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

### 15. Проверить подключение к БД

**Файл:** `cmd/main.go:29-32`

**Рекомендация:** Добавить ping:
```go
if err := db.Ping(); err != nil {
    log.Fatalf("failed to ping database: %v", err)
}
```

---

## 📊 Сравнение с предыдущим ревью

| Проблема | Было | Стало | Статус |
|----------|------|-------|--------|
| Копирование структуры | ❌ | ✅ | Исправлено |
| Валидация входных | ❌ | 🟡 | Частично (баг) |
| JSON encoding ошибки | ❌ | ✅ | Исправлено |
| CORS origin | ❌ | ✅ | Исправлено |
| HTTP клиент создается заново | ❌ | ✅ | Исправлено |
| Тип Days | ❌ | ✅ | Исправлено |
| Timeout HTTP клиента | ❌ | ❌ | Не исправлено |
| Graceful shutdown | ❌ | ❌ | Не исправлено |
| Логирование | ❌ | ❌ | Не исправлено |
| Проверка интерфейса | ❌ | ❌ | Не исправлено |

---

## 🎯 Приоритет исправлений

### Критично (исправить немедленно):
1. 🔴 **Валидация вызывается до декодирования JSON** - строка 30
2. 🔴 **Неправильный порядок WriteHeader/Encode** - строки 41-46
3. 🔴 **Валидация Days для 0 значения** - строка 77

### Важно (исправить в ближайшее время):
4. 🟡 Добавить timeout для HTTP клиента
5. 🟡 Добавить логирование
6. 🟡 Реализовать graceful shutdown
7. 🟡 Добавить проверку компилятором интерфейса

### Желательно:
8. Улучшить обработку ошибок
9. Добавить комментарии
10. Улучшить SQL форматирование

---

## 📝 Итоговая оценка

**Прогресс:** 6 из 29 проблем исправлено (20%)  
**Критические баги:** 3 (новые после изменений)  
**Важные проблемы:** 7  
**Рекомендации:** 10  

**Общая оценка:** 🟡 3.5/5 (Улучшилось, но появились новые баги)

---

## ⚠️ Важно

**После исправлений появились новые баги!** Нужно срочно исправить:
1. Порядок вызова валидации (после декодирования JSON)
2. Порядок WriteHeader/Encode
3. Обработку optional поля Days
