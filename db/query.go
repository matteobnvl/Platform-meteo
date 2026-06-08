package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/matteobnvl/Platform-meteo/internal/model"
)

func GetStations(db *sql.DB) ([]model.Station, error) {
	rows, err := db.Query(`SELECT id, name, latitude, longitude FROM stations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stations []model.Station
	for rows.Next() {
		var s model.Station
		if err := rows.Scan(&s.Id, &s.Name, &s.Latitude, &s.Longitude); err != nil {
			return nil, err
		}
		stations = append(stations, s)
	}
	return stations, rows.Err()
}

func GetStationByID(db *sql.DB, id string) (model.Station, error) {
	var s model.Station
	err := db.QueryRow(`SELECT id, name, latitude, longitude FROM stations WHERE id=$1`, id).
		Scan(&s.Id, &s.Name, &s.Latitude, &s.Longitude)
	return s, err
}

func GetObservationsByStation(db *sql.DB, stationID string, from, to *time.Time, limit, offset int) ([]model.Observation, error) {
	query := `
        SELECT id, station_id, observed_at, temperature, wind_speed, wind_direction, precipitation
        FROM observations WHERE station_id = $1
    `
	args := []any{stationID}
	i := 2
	if from != nil {
		query += fmt.Sprintf(" AND observed_at >= $%d", i)
		args = append(args, *from)
		i++
	}
	if to != nil {
		query += fmt.Sprintf(" AND observed_at <= $%d", i)
		args = append(args, *to)
		i++
	}
	query += " ORDER BY observed_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
		args = append(args, limit, offset)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var obs []model.Observation
	for rows.Next() {
		var o model.Observation
		if err := rows.Scan(&o.Id, &o.StationId, &o.ObservedAt, &o.Temperature, &o.WindSpeed, &o.WindDirection, &o.Precipitation); err != nil {
			return nil, err
		}
		obs = append(obs, o)
	}
	return obs, rows.Err()
}
func InsertStation(db *sql.DB, s model.Station) error {
	_, err := db.Exec(`
        INSERT INTO stations (id, name, latitude, longitude)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO NOTHING
    `, s.Id, s.Name, s.Latitude, s.Longitude)
	return err
}

func UpdateStation(db *sql.DB, s model.Station) error {
	_, err := db.Exec(`
        UPDATE stations SET name=$2, latitude=$3, longitude=$4 WHERE id=$1
    `, s.Id, s.Name, s.Latitude, s.Longitude)
	return err
}

func DeleteStation(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM stations WHERE id=$1`, id)
	return err
}
