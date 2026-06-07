package storage

import (
	"database/sql"

	"github.com/matteobnvl/Platform-meteo/internal/model"
)

func InsertStation(db *sql.DB, s model.Station) error {
	_, err := db.Exec(`
        INSERT INTO stations (id, name, latitude, longitude)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO NOTHING
    `, s.Id, s.Name, s.Latitude, s.Longitude)
	return err
}
