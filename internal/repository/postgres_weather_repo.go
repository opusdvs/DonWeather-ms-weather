package repository

import (
	"database/sql"

	"github.com/opusdvs/DonWeather-ms-weather/internal/domain"
)

type PostgresWeatherRepository struct {
	db *sql.DB
}

func NewPostgresWeatherRepository(db *sql.DB) *PostgresWeatherRepository {
	return &PostgresWeatherRepository{
		db: db,
	}
}

func (p *PostgresWeatherRepository) Save(weather *domain.Weather) error {
	query := `
		INSERT INTO weather (location_name,last_updated,temp_c,humidity,pressure_mb,wind_kph,condition_text)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := p.db.Exec(query,
		weather.Location.Name,
		weather.Current.LastUpdated,
		weather.Current.TempC,
		weather.Current.Humidity,
		weather.Current.PressureMb,
		weather.Current.WindKph,
		weather.Current.Condition.Text,
	)
	return err
}

func (p *PostgresWeatherRepository) Delete(id string) error {
	query := `DELETE FROM weather WHERE id=$1`
	_, err := p.db.Exec(query, id)

	return err
}
