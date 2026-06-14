package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/matteobnvl/Platform-meteo/internal/model"
)

func GetStations(db *sql.DB, country string) ([]model.Station, error) {
	query := `SELECT id, name, country, latitude, longitude FROM stations`
	args := []any{}
	if country != "" {
		query += ` WHERE country = $1`
		args = append(args, country)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stations []model.Station
	for rows.Next() {
		var s model.Station
		if err := rows.Scan(&s.Id, &s.Name, &s.Country, &s.Latitude, &s.Longitude); err != nil {
			return nil, err
		}
		stations = append(stations, s)
	}
	return stations, rows.Err()
}

func GetStationByID(db *sql.DB, id string) (model.Station, error) {
	var s model.Station
	err := db.QueryRow(`SELECT id, name, country, latitude, longitude FROM stations WHERE id=$1`, id).
		Scan(&s.Id, &s.Name, &s.Country, &s.Latitude, &s.Longitude)
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
        INSERT INTO stations (id, name, country, latitude, longitude)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id) DO NOTHING
    `, s.Id, s.Name, s.Country, s.Latitude, s.Longitude)
	return err
}

func UpdateStation(db *sql.DB, s model.Station) error {
	_, err := db.Exec(`
        UPDATE stations SET name=$2, country=$3, latitude=$4, longitude=$5 WHERE id=$1
    `, s.Id, s.Name, s.Country, s.Latitude, s.Longitude)
	return err
}

func DeleteStation(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM stations WHERE id=$1`, id)
	return err
}

func GetDailyAggregate(db *sql.DB, stationID string) ([]model.AggregateResult, error) {
	rows, err := db.Query(`
        SELECT
            DATE(observed_at)::text,
            ROUND(AVG(temperature)::numeric, 2),
            ROUND(MAX(temperature)::numeric, 2),
            ROUND(MIN(temperature)::numeric, 2)
        FROM observations
        WHERE station_id = $1
        GROUP BY DATE(observed_at)
        ORDER BY DATE(observed_at) DESC
    `, stationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []model.AggregateResult
	for rows.Next() {
		var r model.AggregateResult
		if err := rows.Scan(&r.Period, &r.AvgTemp, &r.MaxTemp, &r.MinTemp); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
