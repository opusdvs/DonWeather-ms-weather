package domain

import "context"

type Weather struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`

	Current struct {
		LastUpdated string  `json:"last_updated"`
		TempC       float64 `json:"temp_c"`
		Humidity    float64 `json:"humidity"`
		PressureMb  float64 `json:"pressure_mb"`
		WindKph     float64 `json:"wind_kph"`
		Condition   struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
	Forecast struct {
		Forecastday []struct {
			Date string `json:"date"`
			Day  struct {
				Maxtemp_c   float64 `json:"maxtemp_c"`
				AvgHumidity float64 `json:"avghumidity"`
				MaxwindKph  float64 `json:"maxwind_kph"`
				Condition   struct {
					Text string `json:"text"`
				} `json:"condition"`
			} `json:"day"`
			Hour []struct {
				PressureMb float64 `json:"pressure_mb"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

type WeatherRepository interface {
	Save(context.Context, *Weather) error
	Delete(context.Context, string) error
}
