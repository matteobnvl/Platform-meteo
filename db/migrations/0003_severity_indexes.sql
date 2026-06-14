ALTER TABLE events ADD COLUMN IF NOT EXISTS severity TEXT DEFAULT 'medium';
ALTER TABLE stations ADD COLUMN IF NOT EXISTS country TEXT DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS uq_events_station_type_start
    ON events(station_id, type, started_at);

CREATE INDEX IF NOT EXISTS idx_observations_station_time
    ON observations(station_id, observed_at);
CREATE INDEX IF NOT EXISTS idx_events_station_id ON events(station_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
CREATE INDEX IF NOT EXISTS idx_events_started_at ON events(started_at);
