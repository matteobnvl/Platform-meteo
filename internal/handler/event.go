package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/matteobnvl/Platform-meteo/db"
)

func (a *App) ListEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	eventType := q.Get("type")

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

	events, err := db.GetEvents(a.DB, eventType, from, to, limit, offset)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (a *App) GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		jsonError(w, "id invalide", http.StatusBadRequest)
		return
	}
	event, err := db.GetEventByID(a.DB, id)
	if err == sql.ErrNoRows {
		jsonError(w, fmt.Sprintf("événement %d introuvable", id), http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func (a *App) ListStationEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	events, err := db.GetEventsByStation(a.DB, id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (a *App) EventStats(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("type")
	country := r.URL.Query().Get("country")

	stats, err := db.GetEventStats(a.DB, eventType, country)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
