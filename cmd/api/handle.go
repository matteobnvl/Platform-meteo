package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

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
	obs, err := db.GetObservationsByStation(a.db, id, nil, nil, 0, 0)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(obs)
}
