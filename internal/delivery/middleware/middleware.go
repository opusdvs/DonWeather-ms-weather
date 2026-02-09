package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/opusdvs/DonWeather-ms-weather/internal/domain"
)

// Типизированный ключ для traceId в контексте
type traceIDKey struct{}

var traceIDKeyValue = traceIDKey{}

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), traceIDKeyValue, traceID)
		w.Header().Set("X-Trace-Id", traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTraceID извлекает traceId из контекста
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKeyValue).(string); ok {
		return traceID
	}
	return ""
}

func MiddlewareChain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range middleware {
		h = mw(h)
	}
	return h
}

func LoggerMiddleware(logger domain.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			traceID := GetTraceID(r.Context())

			logger.Info(r.Context(), "request received", domain.Fields{
				Key:   "method",
				Value: r.Method,
			}, domain.Fields{
				Key:   "url",
				Value: r.URL.String(),
			}, domain.Fields{
				Key:   "traceId",
				Value: traceID,
			})
			next.ServeHTTP(w, r)
		})
	}
}
