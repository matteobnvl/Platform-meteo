package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matteobnvl/Platform-meteo/internal/model"
)

func InsertEventIfNew(database *sql.DB, e model.Event) (bool, error) {
	metaJSON, err := json.Marshal(e.Metadata)
	if err != nil {
		return false, err
	}

	res, err := database.Exec(`
		INSERT INTO events (station_id, type, started_at, ended_at, severity, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (station_id, type, started_at) DO NOTHING
	`, e.StationId, e.Type, e.StartedAt, e.EndedAt, e.Severity, metaJSON)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func GetEvents(database *sql.DB, eventType string, from, to *time.Time, limit, offset int) ([]model.Event, error) {
	query := `SELECT id, station_id, type, started_at, ended_at, severity, metadata FROM events WHERE 1=1`
	args := []any{}
	i := 1

	if eventType != "" {
		query += fmt.Sprintf(" AND type = $%d", i)
		args = append(args, eventType)
		i++
	}
	if from != nil {
		query += fmt.Sprintf(" AND started_at >= $%d", i)
		args = append(args, *from)
		i++
	}
	if to != nil {
		query += fmt.Sprintf(" AND started_at <= $%d", i)
		args = append(args, *to)
		i++
	}
	query += " ORDER BY started_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", i, i+1)
		args = append(args, limit, offset)
	}

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func GetEventByID(database *sql.DB, id int) (model.Event, error) {
	row := database.QueryRow(`
		SELECT id, station_id, type, started_at, ended_at, severity, metadata
		FROM events WHERE id = $1
	`, id)
	return scanEvent(row)
}

func GetEventsByStation(database *sql.DB, stationID string) ([]model.Event, error) {
	rows, err := database.Query(`
		SELECT id, station_id, type, started_at, ended_at, severity, metadata
		FROM events WHERE station_id = $1 ORDER BY started_at DESC
	`, stationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func GetEventStats(database *sql.DB, eventType, country string) ([]model.EventStat, error) {
	query := `
		SELECT e.type, COALESCE(s.country, ''), COUNT(*)
		FROM events e
		JOIN stations s ON e.station_id = s.id
		WHERE 1=1
	`
	args := []any{}
	i := 1

	if eventType != "" {
		query += fmt.Sprintf(" AND e.type = $%d", i)
		args = append(args, eventType)
		i++
	}
	if country != "" {
		query += fmt.Sprintf(" AND s.country = $%d", i)
		args = append(args, country)
		i++
	}
	query += " GROUP BY e.type, s.country ORDER BY e.type, s.country"

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []model.EventStat
	for rows.Next() {
		var st model.EventStat
		if err := rows.Scan(&st.Type, &st.Country, &st.Count); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}

func scanEvent(row *sql.Row) (model.Event, error) {
	var e model.Event
	var endedAt sql.NullTime
	var metaRaw []byte

	err := row.Scan(&e.Id, &e.StationId, &e.Type, &e.StartedAt, &endedAt, &e.Severity, &metaRaw)
	if err != nil {
		return e, err
	}
	if endedAt.Valid {
		e.EndedAt = &endedAt.Time
	}
	if metaRaw != nil {
		json.Unmarshal(metaRaw, &e.Metadata)
	}
	return e, nil
}

func scanEvents(rows *sql.Rows) ([]model.Event, error) {
	var events []model.Event
	for rows.Next() {
		var e model.Event
		var endedAt sql.NullTime
		var metaRaw []byte

		if err := rows.Scan(&e.Id, &e.StationId, &e.Type, &e.StartedAt, &endedAt, &e.Severity, &metaRaw); err != nil {
			return nil, err
		}
		if endedAt.Valid {
			e.EndedAt = &endedAt.Time
		}
		if metaRaw != nil {
			json.Unmarshal(metaRaw, &e.Metadata)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
