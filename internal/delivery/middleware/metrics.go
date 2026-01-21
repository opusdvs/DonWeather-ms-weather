package middleware

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal - общее количество HTTP запросов
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTPRequestsSuccessTotal - количество успешных HTTP запросов (2xx, 3xx)
	HTTPRequestsSuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_success_total",
			Help: "Total number of successful HTTP requests",
		},
		[]string{"method", "path"},
	)

	// HTTPRequestsErrorTotal - количество HTTP запросов с ошибками (4xx, 5xx)
	HTTPRequestsErrorTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_error_total",
			Help: "Total number of HTTP requests with errors",
		},
		[]string{"method", "path", "status_code"},
	)
)

// MetricsMiddleware собирает метрики для всех HTTP запросов
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Создаем ResponseWriter для перехвата статус кода
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Выполняем следующий обработчик
		next.ServeHTTP(rw, r)

		// Извлекаем метки для метрик
		method := r.Method
		path := r.URL.Path
		statusCode := strconv.Itoa(rw.statusCode)

		// Увеличиваем счетчик общего количества запросов
		HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()

		// Увеличиваем счетчик успешных запросов (2xx, 3xx)
		if rw.statusCode >= 200 && rw.statusCode < 400 {
			HTTPRequestsSuccessTotal.WithLabelValues(method, path).Inc()
		}

		// Увеличиваем счетчик запросов с ошибками (4xx, 5xx)
		if rw.statusCode >= 400 {
			HTTPRequestsErrorTotal.WithLabelValues(method, path, statusCode).Inc()
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
