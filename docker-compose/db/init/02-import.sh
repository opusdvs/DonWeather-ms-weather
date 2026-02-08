#!/bin/bash
set -e

CITIES_FILE="/data/koord_russia.csv"
TEMP_FILE="/tmp/cities_import.csv"

if [ ! -f "$CITIES_FILE" ]; then
  echo "Skipping cities import: $CITIES_FILE not found."
  exit 0
fi

echo "Importing cities..."

export PGPASSWORD="${POSTGRES_PASSWORD:-mypassword}"
export LC_ALL=C.UTF-8
export LANG=C.UTF-8

# Конвертируем CSV: пропускаем заголовок, заменяем запятую на точку в координатах (lat, lng)
# Формат: Город;Регион;Федеральный округ;lat;lng
# Координаты в файле с запятой (52,651657), нужно заменить на точку для PostgreSQL
# Пробуем конвертировать из Windows-1251 в UTF-8, если не получается - используем как есть
if command -v iconv >/dev/null 2>&1; then
  tail -n +2 "$CITIES_FILE" | iconv -f WINDOWS-1251 -t UTF-8 2>/dev/null | sed 's/,/./g' > "$TEMP_FILE" || \
  tail -n +2 "$CITIES_FILE" | sed 's/,/./g' > "$TEMP_FILE"
else
  tail -n +2 "$CITIES_FILE" | sed 's/,/./g' > "$TEMP_FILE"
fi

psql -U "${POSTGRES_USER:-postgres}" -d georesolve -v ON_ERROR_STOP=1 <<EOF
-- Устанавливаем кодировку клиента UTF-8 для корректной работы с русским языком
SET client_encoding = 'UTF8';

-- Импорт данных (пропускаем заголовок, уже удалён через tail)
\copy cities(name, region, federal_district, latitude, longitude) FROM '$TEMP_FILE' WITH (FORMAT csv, DELIMITER ';', HEADER false, NULL '');

-- Заполняем геометрию из координат
UPDATE cities
SET geom = ST_SetSRID(ST_Point(longitude, latitude), 4326);

EOF

rm -f "$TEMP_FILE"
echo "Done."
