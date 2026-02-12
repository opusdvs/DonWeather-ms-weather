-- База georesolve (weather создаётся через POSTGRES_DB)
-- Создаём базу с поддержкой русского языка (UTF-8)
-- Используем C.UTF-8 как универсальную локаль, поддерживающую UTF-8 и русский язык
-- Если нужна именно ru_RU.UTF-8, убедитесь что она установлена в контейнере
CREATE DATABASE georesolve
  WITH ENCODING = 'UTF8'
       LC_COLLATE = 'C.UTF-8'
       LC_CTYPE = 'C.UTF-8'
       TEMPLATE = template0;

CREATE DATABASE subscribe
  WITH ENCODING = 'UTF8'
       LC_COLLATE = 'C.UTF-8'
       LC_CTYPE = 'C.UTF-8'
       TEMPLATE = template0;
-- Роль для базы georesolve
CREATE ROLE georesolv_user WITH LOGIN PASSWORD 'georesolv_password';
CREATE ROLE subscribe_user WITH LOGIN PASSWORD 'subscribe_password';

-- PostGIS только в georesolve
\c georesolve
CREATE EXTENSION IF NOT EXISTS postgis;

-- Таблица cities по структуре koord_russia.csv (CSV, разделитель ';')
-- Колонки: Город;Регион;Федеральный округ;lat;lng
CREATE TABLE cities (
  id              SERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  region          TEXT,
  federal_district TEXT,
  latitude        DOUBLE PRECISION NOT NULL,
  longitude       DOUBLE PRECISION NOT NULL,
  geom            GEOMETRY(Point, 4326)
);

CREATE INDEX cities_geom_idx ON cities USING GIST (geom);
CREATE INDEX cities_name_idx ON cities (name);
CREATE INDEX cities_region_idx ON cities (region);
CREATE INDEX cities_federal_district_idx ON cities (federal_district);


-- Выдаём права пользователю georesolv_user на таблицу cities
GRANT SELECT ON cities TO georesolv_user;
GRANT USAGE, SELECT ON SEQUENCE cities_id_seq TO georesolv_user;

\c subscribe
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
GRANT USAGE, CREATE ON SCHEMA public TO subscribe_user;

-- Владельцы баз
\c postgres
ALTER DATABASE georesolve OWNER TO georesolv_user;
ALTER DATABASE subscribe OWNER TO subscribe_user;

