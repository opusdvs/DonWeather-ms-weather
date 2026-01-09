package usecase

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/opusdvs/DonWeather-ms-weather/internal/domain"
)

type WeatherAction interface {
	Save(*domain.Weather) error
	Delete(id string) error
}

type WeatherServie struct {
	weather WeatherAction
}

func NewWeatherServie(weather WeatherAction) *WeatherServie {
	return &WeatherServie{
		weather: weather,
	}
}

func (ws *WeatherServie) FeatchAndSaveWeather(q, lang, days string) (*domain.Weather, error) {
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
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(u.String())
	if err != nil {
		fmt.Println("тут")
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("weather API error: %s", string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("тут 3")
		return nil, err
	}
	var weather domain.Weather
	if err := json.Unmarshal(data, &weather); err != nil {
		return nil, err
	}
	if err := ws.weather.Save(&weather); err != nil {
		return nil, err
	}

	return &weather, nil
}
