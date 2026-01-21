package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/opusdvs/DonWeather-ms-weather/internal/domain"
)

type WeatherService struct {
	weatherRepo domain.WeatherRepository
	apiKey      string
	httpClient  *http.Client
	logger      domain.Logger
}

func NewWeatherService(weatherRepo domain.WeatherRepository, apiKey string, logger domain.Logger) *WeatherService {
	return &WeatherService{
		weatherRepo: weatherRepo,
		apiKey:      apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (ws *WeatherService) FetchAndSaveWeather(ctx context.Context, q, lang string, days string) (*domain.Weather, error) {
	baseUrl := "https://api.weatherapi.com/v1/forecast.json"
	u, err := url.Parse(baseUrl)
	if err != nil {
		ws.logger.Error(ctx, "failed to parse URL", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return nil, err
	}
	query := u.Query()
	query.Set("q", q)
	query.Set("key", ws.apiKey)
	query.Set("lang", lang)
	query.Set("days", days)
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		ws.logger.Error(ctx, "failed to create request", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return nil, err
	}
	resp, err := ws.httpClient.Do(req)
	if err != nil {
		ws.logger.Error(ctx, "failed to do request", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			ws.logger.Error(ctx, "failed to read response body", domain.Fields{
				Key:   "error",
				Value: err.Error(),
			})
			return nil, err
		}
		ws.logger.Error(ctx, "failed to get weather", domain.Fields{
			Key:   "status",
			Value: resp.StatusCode,
		}, domain.Fields{
			Key:   "body",
			Value: string(body),
		})
		return nil, fmt.Errorf("failed to get weather: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		ws.logger.Error(ctx, "failed to read response body", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return nil, err
	}
	var weather domain.Weather
	if err := json.Unmarshal(data, &weather); err != nil {
		ws.logger.Error(ctx, "failed to unmarshal response body", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return nil, err
	}
	ws.logger.Info(ctx, "response from weather API", domain.Fields{
		Key:   "status",
		Value: resp.StatusCode,
	}, domain.Fields{
		Key:   "url",
		Value: u.String(),
	}, domain.Fields{
		Key:   "weather",
		Value: weather,
	})

	if err := ws.weatherRepo.Save(ctx, &weather); err != nil {
		ws.logger.Error(ctx, "failed to save weather", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return nil, err
	}
	return &weather, nil
}
