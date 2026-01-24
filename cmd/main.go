package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	delivery "github.com/opusdvs/DonWeather-ms-weather/internal/delivery/http"
	"github.com/opusdvs/DonWeather-ms-weather/internal/delivery/middleware"
	"github.com/opusdvs/DonWeather-ms-weather/internal/domain"
	"github.com/opusdvs/DonWeather-ms-weather/internal/infrastructure"
	"github.com/opusdvs/DonWeather-ms-weather/internal/repository"
	"github.com/opusdvs/DonWeather-ms-weather/internal/usecase"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger := infrastructure.NewLoggerService()
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		log.Fatal("DB_PASSWORD environment variable is required")
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		log.Fatal("DB_USER environment variable is required")
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		log.Fatal("DB_HOST environment variable is required")
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		log.Fatal("DB_PORT environment variable is required")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("DB_NAME environment variable is required")
	}
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("WEATHER_API_KEY environment variable is required")
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := db.PingContext(pingCtx); err != nil {
		logger.Error(context.Background(), "failed to ping database", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return
	}

	repo := repository.NewPostgresWeatherRepository(db, logger)
	act := usecase.NewWeatherService(repo, apiKey, logger)
	handler := delivery.NewWeatherHandler(act)

	if err := RunMigrations(logger, dsn); err != nil {
		logger.Error(context.Background(), "migration error", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return
	}

	healthHandler := delivery.NewHealthHandler(logger, db)
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health/live", healthHandler.LivenessProbe)
	healthMux.HandleFunc("/health/ready", healthHandler.ReadinessProbe)

	mux := http.NewServeMux()
	mux.HandleFunc("/weather/register", handler.Register)
	handlerMiddleware := middleware.MiddlewareChain(mux,
		middleware.MetricsMiddleware,
		middleware.CorsMiddleware,
		middleware.LoggerMiddleware(logger),
		middleware.TraceMiddleware,
	)

	mainMux := http.NewServeMux()
	mainMux.Handle("/", handlerMiddleware)
	mainMux.Handle("/health/", healthMux)
	mainMux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":8080",
		Handler: mainMux,
	}

	go func(s *http.Server) {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "Server listen and serve failed", domain.Fields{
				Key:   "error",
				Value: err.Error(),
			})
			return
		}
		logger.Info(context.Background(), "Server listen and serve success", domain.Fields{
			Key:   "server",
			Value: "Server started at :8080",
		})
	}(server)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = server.Shutdown(ctx)
	if err != nil {
		logger.Error(context.Background(), "Server Shutdown failed", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return
	}
	logger.Info(context.Background(), "Server Shutdown success", domain.Fields{
		Key:   "message",
		Value: "Server gracefully shut down",
	})
}

func RunMigrations(logger domain.Logger, dsn string) error {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		logger.Error(context.Background(), "failed to create migrate instance", domain.Fields{
			Key:   "error",
			Value: err.Error(),
		})
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
