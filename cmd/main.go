package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	delivery "github.com/opusdvs/DonWeather-ms-weather/internal/delivery/http"
	"github.com/opusdvs/DonWeather-ms-weather/internal/repository"
	"github.com/opusdvs/DonWeather-ms-weather/internal/usecase"
)

func main() {
	//dsn := "host=localhost user=myuser password=mypassword dbname=weather port=5432 sslmode=disable"
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "postgres://myuser:mypassword@localhost:5432/weather?sslmode=disable"
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if err := RunMigrations(dsn); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	repo := repository.NewPostgresWeatherRepository(db)
	act := usecase.NewWeatherService(repo)
	handler := delivery.NewWeatherHandler(act)

	http.Handle("/weather/register", delivery.CorsMiddleware(http.HandlerFunc(handler.Register)))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func RunMigrations(dsn string) error {
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
