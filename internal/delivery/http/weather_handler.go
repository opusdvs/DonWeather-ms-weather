package delivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
	ctx := r.Context()
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

func ValidateWeatherRequest(reqBody WeatherRequest) error {
	if reqBody.Q == "" {
		return errors.New("field 'q' is required")
	}
	if reqBody.Lang == "" {
		return errors.New("field 'lang' is required")
	}
	days, err := strconv.Atoi(reqBody.Days)
	if err != nil {
		return errors.New("field 'days' must be a number")
	}
	if days < 0 || days > 14 { // 0 не пройдет валидацию, так как Days int
		return errors.New("field 'days' must be between 0 and 14")
	}
	return nil
}
