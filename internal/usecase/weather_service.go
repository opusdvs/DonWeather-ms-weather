package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/opusdvs/DonWeather-ms-weather/internal/domain"
)

type WeatherAction interface {
	Save(context.Context, *domain.Weather) error
	Delete(context.Context, string) error
}

type WeatherServie struct {
	weather WeatherAction
}

func NewWeatherServie(weather WeatherAction) *WeatherServie {
	return &WeatherServie{
		weather: weather,
	}
}

func (ws *WeatherServie) FeatchAndSaveWeather(ctx context.Context, q, lang, days string) (*domain.Weather, error) {
	baseUrl := "https://api.weatherapi.com/v1/forecast.json"
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	query.Set("q", q)
	query.Set("key", "cd3200bf0f914528862150404260801")
	query.Set("lang", lang)
	query.Set("days", days)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("weather API error: %s", string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var weather domain.Weather
	if err := json.Unmarshal(data, &weather); err != nil {
		return nil, err
	}
	if err := ws.weather.Save(ctx, &weather); err != nil {
		return nil, err
	}

	return &weather, nil
}
