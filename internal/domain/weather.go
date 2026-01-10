package domain

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
}
