# Объяснение проблемы с интерфейсом репозитория

## 🔴 Текущая проблема

### Структура зависимостей сейчас:
```
┌─────────────────┐
│     usecase     │  ← Определяет интерфейс WeatherAction
│                 │     (строка 14-17 в weather_service.go)
└────────┬────────┘
         │ использует
         ▼
┌─────────────────┐
│   repository    │  ← Конкретная реализация PostgresWeatherRepository
│   (postgres)    │     НЕ реализует явно интерфейс
└─────────────────┘
```

### Почему это плохо:

1. **Нарушение Clean Architecture**
   - Domain должен быть независимым от других слоев
   - Интерфейс должен быть в domain, а не в usecase
   - Зависимости должны идти ВНУТРЬ (к domain), а не наружу

2. **Тестирование затруднено**
   ```go
   // Сейчас можно только мокать интерфейс из usecase
   // Но это не соответствует архитектуре
   ```

3. **Сильная связанность**
   - usecase знает о деталях реализации
   - Невозможно легко заменить PostgreSQL на MongoDB/MySQL

4. **Плохое именование**
   - `WeatherAction` - неясное название
   - Должно быть `WeatherRepository` в domain

---

## ✅ Правильное решение

### Правильная структура зависимостей:
```
┌─────────────────┐
│     domain      │  ← Определяет интерфейс WeatherRepository
│                 │     (контракт, что может делать репозиторий)
└────────┬────────┘
         ▲
         │ реализует
┌────────┴────────┐
│   repository    │  ← PostgresWeatherRepository реализует интерфейс
│   (postgres)    │
└────────┬────────┘
         ▲
         │ использует
┌────────┴────────┐
│     usecase     │  ← Работает только с интерфейсом, не знает о БД
└─────────────────┘
```

### Преимущества:

1. **✅ Clean Architecture соблюдена**
   - Domain не зависит ни от чего
   - Все слои зависят от domain
   - Легко добавить новые реализации (MongoDB, In-Memory для тестов)

2. **✅ Легкое тестирование**
   ```go
   // Можно создать мок-репозиторий в domain
   type MockWeatherRepository struct { ... }
   ```

3. **✅ Слабая связанность**
   - usecase не знает, что используется PostgreSQL
   - Можно менять БД без изменения usecase

4. **✅ Правильное именование**
   - `WeatherRepository` - понятное название
   - Находится в правильном месте (domain)

---

## 📝 Пример правильного кода

### 1. domain/weather_repository.go (НОВЫЙ ФАЙЛ)
```go
package domain

import "context"

// WeatherRepository определяет контракт для работы с погодными данными
// Этот интерфейс находится в domain - самом внутреннем слое
type WeatherRepository interface {
    Save(ctx context.Context, weather *Weather) error
    Delete(ctx context.Context, id string) error
    // В будущем можно добавить:
    // GetByLocation(ctx context.Context, location string) (*Weather, error)
    // GetAll(ctx context.Context) ([]*Weather, error)
}
```

### 2. usecase/weather_service.go (ИЗМЕНИТЬ)
```go
package usecase

import (
    "context"
    // ... другие импорты
    "github.com/opusdvs/DonWeather-ms-weather/internal/domain"
)

// УДАЛИТЬ интерфейс WeatherAction отсюда!

type WeatherService struct {
    // Использовать интерфейс из domain
    weatherRepo domain.WeatherRepository
}

func NewWeatherService(weatherRepo domain.WeatherRepository) *WeatherService {
    return &WeatherService{
        weatherRepo: weatherRepo, // Используем интерфейс из domain
    }
}

func (ws *WeatherService) FetchAndSaveWeather(...) (*domain.Weather, error) {
    // ...
    // Используем интерфейс из domain
    if err := ws.weatherRepo.Save(ctx, &weather); err != nil {
        return nil, err
    }
    // ...
}
```

### 3. repository/postgres_weather_repo.go (ИЗМЕНИТЬ)
```go
package repository

import (
    "context"
    "database/sql"
    "github.com/opusdvs/DonWeather-ms-weather/internal/domain"
)

// PostgresWeatherRepository реализует интерфейс из domain
// Это гарантируется компилятором - если метод не реализован, будет ошибка
var _ domain.WeatherRepository = (*PostgresWeatherRepository)(nil)

type PostgresWeatherRepository struct {
    db *sql.DB
}

func NewPostgresWeatherRepository(db *sql.DB) domain.WeatherRepository {
    return &PostgresWeatherRepository{
        db: db,
    }
}

// Save реализует domain.WeatherRepository
func (p *PostgresWeatherRepository) Save(ctx context.Context, weather *domain.Weather) error {
    // ... существующий код
}

// Delete реализует domain.WeatherRepository
func (p *PostgresWeatherRepository) Delete(ctx context.Context, id string) error {
    // ... существующий код
}
```

### 4. cmd/main.go (БЕЗ ИЗМЕНЕНИЙ)
```go
// Код остается таким же - работает через интерфейс
repo := repository.NewPostgresWeatherRepository(db)
act := usecase.NewWeatherService(repo) // Принимает интерфейс domain.WeatherRepository
```

---

## 🎯 Ключевые моменты

1. **Интерфейс в domain** - определяет ЧТО нужно, а не КАК реализовано
2. **Реализация в repository** - конкретная реализация (PostgreSQL, MongoDB и т.д.)
3. **Использование в usecase** - работает только с интерфейсом
4. **Проверка компилятором** - `var _ domain.WeatherRepository = (*PostgresWeatherRepository)(nil)` гарантирует, что все методы реализованы

---

## 🧪 Преимущества для тестирования

Теперь можно легко создать mock для тестов:

```go
// internal/usecase/weather_service_test.go
type mockWeatherRepository struct {
    saveFunc func(ctx context.Context, weather *domain.Weather) error
}

func (m *mockWeatherRepository) Save(ctx context.Context, weather *domain.Weather) error {
    if m.saveFunc != nil {
        return m.saveFunc(ctx, weather)
    }
    return nil
}

func (m *mockWeatherRepository) Delete(ctx context.Context, id string) error {
    return nil
}

func TestWeatherService_FetchAndSaveWeather(t *testing.T) {
    mockRepo := &mockWeatherRepository{
        saveFunc: func(ctx context.Context, weather *domain.Weather) error {
            // Проверяем, что сохраняется правильная погода
            assert.Equal(t, "Amsterdam", weather.Location.Name)
            return nil
        },
    }
    
    service := NewWeatherService(mockRepo)
    // ... тест
}
```

---

## 📊 Сравнение

| Аспект | Сейчас (плохо) | После исправления (хорошо) |
|--------|----------------|----------------------------|
| Где интерфейс | usecase | domain |
| Тестирование | Сложно | Легко |
| Замена БД | Нужно менять usecase | Меняем только repository |
| Соответствие Clean Architecture | ❌ | ✅ |
| Понятность кода | Средняя | Высокая |
