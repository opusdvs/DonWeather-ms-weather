# Варианты передачи API токена в FetchAndSaveWeather

## 🔴 Текущая проблема

API токен захардкожен в коде (строка 32):
```go
query.Set("key", "cd3200bf0f914528862150404260801")  // ❌ Небезопасно!
```

---

## ✅ Вариант 1: Через поле структуры (РЕКОМЕНДУЕТСЯ)

### Преимущества:
- ✅ Безопасно (нет хардкода в коде)
- ✅ Легко тестировать
- ✅ Соответствует принципам Dependency Injection
- ✅ Можно использовать переменные окружения
- ✅ Не меняет сигнатуру метода

### Реализация:

**internal/usecase/weather_service.go:**
```go
type WeatherService struct {
	weatherRepo domain.WeatherRepository
	apiKey      string  // Новое поле
}

func NewWeatherService(weatherRepo domain.WeatherRepository, apiKey string) *WeatherService {
	if apiKey == "" {
		panic("WEATHER_API_KEY is required")  // или возвращать ошибку
	}
	return &WeatherService{
		weatherRepo: weatherRepo,
		apiKey:      apiKey,
	}
}

func (ws *WeatherService) FetchAndSaveWeather(ctx context.Context, q, lang, days string) (*domain.Weather, error) {
	// ...
	query.Set("key", ws.apiKey)  // ✅ Используем поле
	// ...
}
```

**cmd/main.go:**
```go
apiKey := os.Getenv("WEATHER_API_KEY")
if apiKey == "" {
	log.Fatal("WEATHER_API_KEY environment variable is required")
}

repo := repository.NewPostgresWeatherRepository(db)
act := usecase.NewWeatherService(repo, apiKey)  // Передаем токен
handler := delivery.NewWeatherHandler(act)
```

---

## ✅ Вариант 2: Через параметр метода

### Преимущества:
- ✅ Гибко (можно передавать разные токены)
- ✅ Не нужно хранить в структуре

### Недостатки:
- ❌ Усложняет сигнатуру метода
- ❌ Нужно передавать токен каждый раз

### Реализация:

```go
func (ws *WeatherService) FetchAndSaveWeather(
	ctx context.Context, 
	q, lang, days string,
	apiKey string,  // Новый параметр
) (*domain.Weather, error) {
	// ...
	query.Set("key", apiKey)
	// ...
}
```

**Использование в handler:**
```go
apiKey := os.Getenv("WEATHER_API_KEY")
weather, err := wh.svc.FetchAndSaveWeather(ctx, reqBody.Q, reqBody.Lang, reqBody.Days, apiKey)
```

---

## ✅ Вариант 3: Через конфигурационный интерфейс

### Преимущества:
- ✅ Максимальная гибкость
- ✅ Можно добавить другие настройки
- ✅ Соответствует Clean Architecture

### Недостатки:
- ❌ Избыточно для простого случая

### Реализация:

**internal/domain/config.go:**
```go
type WeatherAPIConfig interface {
	APIKey() string
	BaseURL() string
}
```

**internal/usecase/weather_service.go:**
```go
type WeatherService struct {
	weatherRepo domain.WeatherRepository
	config      domain.WeatherAPIConfig
}

func NewWeatherService(weatherRepo domain.WeatherRepository, config domain.WeatherAPIConfig) *WeatherService {
	return &WeatherService{
		weatherRepo: weatherRepo,
		config:      config,
	}
}

func (ws *WeatherService) FetchAndSaveWeather(...) (*domain.Weather, error) {
	query.Set("key", ws.config.APIKey())
	u, err := url.Parse(ws.config.BaseURL())
	// ...
}
```

---

## ✅ Вариант 4: Через переменную окружения в usecase

### Преимущества:
- ✅ Просто
- ✅ Нет необходимости передавать

### Недостатки:
- ❌ Сложнее тестировать
- ❌ Нарушает принципы DI

### Реализация:

```go
func (ws *WeatherService) FetchAndSaveWeather(...) (*domain.Weather, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("WEATHER_API_KEY environment variable is required")
	}
	query.Set("key", apiKey)
	// ...
}
```

---

## 📊 Сравнение вариантов

| Критерий | Вариант 1 (поле) | Вариант 2 (параметр) | Вариант 3 (интерфейс) | Вариант 4 (env в usecase) |
|----------|------------------|---------------------|----------------------|---------------------------|
| Безопасность | ✅✅ | ✅✅ | ✅✅ | ✅ |
| Тестируемость | ✅✅ | ✅✅ | ✅✅ | ❌ |
| Гибкость | ✅✅ | ✅✅ | ✅✅✅ | ❌ |
| Простота | ✅✅ | ✅ | ❌ | ✅✅ |
| DI принципы | ✅✅ | ✅ | ✅✅✅ | ❌ |
| Рекомендация | ⭐⭐⭐ | ⭐⭐ | ⭐ | ⭐ |

---

## 🎯 Рекомендация

**Использовать Вариант 1** - через поле структуры:

1. Безопасно - нет хардкода
2. Легко тестировать - можно передать мок токен
3. Просто реализовать
4. Соответствует принципам Dependency Injection
5. Можно использовать переменные окружения на уровне main
