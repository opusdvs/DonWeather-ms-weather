package delivery

import "net/http"

type WeatherHTTPHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
}

type WeatherRequest struct {
	Q    string `json:"q"`
	Lang string `json:"lang"`
	Days string `json:"days"`
}
