package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/opusdvs/DonWeather-ms-weather/internal/domain"
)

type WeatherService struct {
	weatherRepo domain.WeatherRepository
	apiKey      string
	httpClient  *http.Client
}

func NewWeatherService(weatherRepo domain.WeatherRepository, apiKey string) *WeatherService {
	return &WeatherService{
		weatherRepo: weatherRepo,
		apiKey:      apiKey,
		httpClient:  &http.Client{},
	}
}

func (ws *WeatherService) FetchAndSaveWeather(ctx context.Context, q, lang string, days int) (*domain.Weather, error) {
	baseUrl := "https://api.weatherapi.com/v1/forecast.json"
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	query := u.Query()
	query.Set("q", q)
	query.Set("key", ws.apiKey)
	query.Set("lang", lang)
	query.Set("days", strconv.Itoa(days))
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := ws.httpClient.Do(req)
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
	if err := ws.weatherRepo.Save(ctx, &weather); err != nil {
		return nil, err
	}

	return &weather, nil
}
