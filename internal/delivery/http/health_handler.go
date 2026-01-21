package delivery

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/opusdvs/DonWeather-ms-weather/internal/domain"
)

type HealthHandler struct {
	logger domain.Logger
	db     *sql.DB
}

func NewHealthHandler(logger domain.Logger, db *sql.DB) *HealthHandler {
	return &HealthHandler{
		logger: logger,
		db:     db,
	}
}

func (h *HealthHandler) LivenessProbe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	h.logger.Info(r.Context(), "Liveness probe successful", domain.Fields{
		Key:   "status",
		Value: "OK",
	}, domain.Fields{
		Key:   "message",
		Value: "Liveness probe successful",
	}, domain.Fields{
		Key:   "url",
		Value: r.URL.String(),
	})
}

func (h *HealthHandler) ReadinessProbe(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		http.Error(w, "Database not ready", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	h.logger.Info(r.Context(), "Readiness probe successful", domain.Fields{
		Key:   "status",
		Value: "OK",
	}, domain.Fields{
		Key:   "message",
		Value: "Readiness probe successful",
	}, domain.Fields{
		Key:   "url",
		Value: r.URL.String(),
	})
}
