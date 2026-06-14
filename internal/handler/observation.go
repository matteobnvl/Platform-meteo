package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/matteobnvl/Platform-meteo/db"
)

func (a *App) ListObservations(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	q := r.URL.Query()

	var from, to *time.Time
	if v := q.Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			jsonError(w, "format from invalide, attendu YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		from = &t
	}
	if v := q.Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			jsonError(w, "format to invalide, attendu YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		to = &t
	}

	limit, offset := 0, 0
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			jsonError(w, "limit doit être un entier positif", http.StatusBadRequest)
			return
		}
		limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			jsonError(w, "offset doit être un entier positif", http.StatusBadRequest)
			return
		}
		offset = n
	}

	obs, err := db.GetObservationsByStation(a.DB, id, from, to, limit, offset)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(obs)
}

func (a *App) AggregateObservations(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station_id")
	if stationID == "" {
		jsonError(w, "station_id requis", http.StatusBadRequest)
		return
	}
	period := r.URL.Query().Get("period")
	if period != "daily" {
		jsonError(w, "seul period=daily est supporté", http.StatusBadRequest)
		return
	}
	results, err := db.GetDailyAggregate(a.DB, stationID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
