CREATE TABLE weather (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    location_name  TEXT NOT NULL,
    last_updated   TIMESTAMPTZ NOT NULL,
    temp_c         DOUBLE PRECISION NOT NULL,
    humidity       DOUBLE PRECISION NOT NULL,
    pressure_mb    DOUBLE PRECISION NOT NULL,
    wind_kph       DOUBLE PRECISION NOT NULL,
    condition_text TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Индексы для удобного поиска
CREATE INDEX idx_weather_api_response_last_updated ON weather (last_updated);
CREATE INDEX idx_weather_api_response_location ON weather (location_name);