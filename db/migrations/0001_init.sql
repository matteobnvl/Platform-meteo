CREATE TABLE stations (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    latitude    FLOAT,
    longitude   FLOAT
);

CREATE TABLE observations (
    id               SERIAL PRIMARY KEY,
    station_id       TEXT REFERENCES stations(id),
    observed_at      TIMESTAMP NOT NULL,
    temperature      FLOAT,
    wind_speed       FLOAT,
    wind_direction   FLOAT,
    precipitation1   FLOAT,
    precipitation3   FLOAT,
    precipitation6   FLOAT,
    precipitation12  FLOAT,
    precipitation24  FLOAT
);

CREATE TABLE events (
    id          SERIAL PRIMARY KEY,
    station_id  TEXT REFERENCES stations(id),
    type        TEXT NOT NULL,
    started_at  TIMESTAMP NOT NULL,
    ended_at    TIMESTAMP,
    metadata    JSONB
);