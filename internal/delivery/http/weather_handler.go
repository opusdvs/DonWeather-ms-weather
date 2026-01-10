package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/opusdvs/DonWeather-ms-weather/internal/usecase"
)

type weatherHandler struct {
	svc *usecase.WeatherService
}

func NewWeatherHandler(svc *usecase.WeatherService) WeatherHTTPHandler {
	return &weatherHandler{
		svc: svc,
	}
}

func (wh *weatherHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var reqBody WeatherRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if err := ValidateWeatherRequest(reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	weather, err := wh.svc.FetchAndSaveWeather(ctx, reqBody.Q, reqBody.Lang, reqBody.Days)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch/save weather: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(weather); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corsOrigin := os.Getenv("CORS_ORIGIN")
		if corsOrigin == "" {
			corsOrigin = "*"
		}
		// Разрешаем конкретный origin frontend
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
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

func ValidateWeatherRequest(reqBody WeatherRequest) error {
	if reqBody.Q == "" {
		return errors.New("field 'q' is required")
	}
	if reqBody.Lang == "" {
		return errors.New("field 'lang' is required")
	}
	if reqBody.Days < 0 || reqBody.Days > 14 { // 0 не пройдет валидацию, так как Days int
		return errors.New("field 'days' must be between 0 and 14")
	}
	return nil
}
