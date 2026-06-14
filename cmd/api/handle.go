package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/matteobnvl/Platform-meteo/db"
	"github.com/matteobnvl/Platform-meteo/internal/model"
)

type App struct{ db *sql.DB }

func (a *App) listStations(w http.ResponseWriter, r *http.Request) {
	stations, err := db.GetStations(a.db)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stations)
}

func (a *App) getStation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	st, err := db.GetStationByID(a.db, id)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("station %q introuvable", id), 404)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(st)
}

func (a *App) createStation(w http.ResponseWriter, r *http.Request) {
	var st model.Station
	if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
		http.Error(w, "JSON invalide", 400)
		return
	}
	if err := db.InsertStation(a.db, st); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(st)
}

func (a *App) updateStation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var st model.Station
	if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
		http.Error(w, "JSON invalide", 400)
		return
	}
	st.Id = id
	if err := db.UpdateStation(a.db, st); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(st)
}

func (a *App) deleteStation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := db.DeleteStation(a.db, id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}

func (a *App) listObservations(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	q := r.URL.Query()

	var from, to *time.Time
	if v := q.Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			http.Error(w, "format from invalide, attendu YYYY-MM-DD", 400)
			return
		}
		from = &t
	}
	if v := q.Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			http.Error(w, "format to invalide, attendu YYYY-MM-DD", 400)
			return
		}
		to = &t
	}

	limit, offset := 0, 0
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			http.Error(w, "limit doit être un entier positif", 400)
			return
		}
		limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			http.Error(w, "offset doit être un entier positif", 400)
			return
		}
		offset = n
	}

	obs, err := db.GetObservationsByStation(a.db, id, from, to, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(obs)
}

func (a *App) aggregateObservations(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station_id")
	if stationID == "" {
		http.Error(w, "station_id requis", 400)
		return
	}
	period := r.URL.Query().Get("period")
	if period != "daily" {
		http.Error(w, "seul period=daily est supporté", 400)
		return
	}
	results, err := db.GetDailyAggregate(a.db, stationID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (a *App) listEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	eventType := q.Get("type")

	var from, to *time.Time
	if v := q.Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			http.Error(w, "format from invalide, attendu YYYY-MM-DD", 400)
			return
		}
		from = &t
	}
	if v := q.Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			http.Error(w, "format to invalide, attendu YYYY-MM-DD", 400)
			return
		}
		to = &t
	}

	limit, offset := 0, 0
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			http.Error(w, "limit doit être un entier positif", 400)
			return
		}
		limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			http.Error(w, "offset doit être un entier positif", 400)
			return
		}
		offset = n
	}

	events, err := db.GetEvents(a.db, eventType, from, to, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (a *App) getEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "id invalide", 400)
		return
	}
	event, err := db.GetEventByID(a.db, id)
	if err == sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("événement %d introuvable", id), 404)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (a *App) listStationEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	events, err := db.GetEventsByStation(a.db, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (a *App) eventStats(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("type")
	country := r.URL.Query().Get("country")

	stats, err := db.GetEventStats(a.db, eventType, country)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
