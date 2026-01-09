package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/opusdvs/DonWeather-ms-weather/internal/usecase"
)

type weatherHandler struct {
	svc usecase.WeatherServie
}

func NewWeatherHandler(svc *usecase.WeatherServie) WeatherHTTPHandler {
	return &weatherHandler{
		svc: *svc,
	}
}

func (wh *weatherHandler) Register(w http.ResponseWriter, r *http.Request) {

	var reqBody WeatherRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	weather, err := wh.svc.FeatchAndSaveWeather(reqBody.Q, reqBody.Lang, reqBody.Days)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch/save weather: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(weather)
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем конкретный origin frontend
		w.Header().Set("Access-Control-Allow-Origin", "http://185.196.117.162:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Обрабатываем preflight OPTIONS
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
