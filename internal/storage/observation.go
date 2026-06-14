package storage

import (
	"database/sql"

	"github.com/matteobnvl/Platform-meteo/internal/model"
)

func InsertBatch(db *sql.DB, observations []model.Observation) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO observations 
			(station_id, observed_at, temperature, wind_speed, wind_direction, precipitation)
		VALUES 
			($1, $2, $3, $4, $5, $6)
		ON CONFLICT (station_id, observed_at) DO NOTHING`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, o := range observations {
		_, err = stmt.Exec(o.StationId, o.ObservedAt, o.Temperature, o.WindSpeed, o.WindDirection, o.Precipitation)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
