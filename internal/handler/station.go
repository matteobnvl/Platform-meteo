package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/matteobnvl/Platform-meteo/db"
	"github.com/matteobnvl/Platform-meteo/internal/model"
)

type App struct{ DB *sql.DB }

func (a *App) ListStations(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	stations, err := db.GetStations(a.DB, country)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stations)
}

func (a *App) GetStation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	st, err := db.GetStationByID(a.DB, id)
	if err == sql.ErrNoRows {
		jsonError(w, fmt.Sprintf("station %q introuvable", id), http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(st)
}

func (a *App) CreateStation(w http.ResponseWriter, r *http.Request) {
	var st model.Station
	if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
		jsonError(w, "JSON invalide", http.StatusBadRequest)
		return
	}
	if err := db.InsertStation(a.DB, st); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(st)
}

func (a *App) UpdateStation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var st model.Station
	if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
		jsonError(w, "JSON invalide", http.StatusBadRequest)
		return
	}
	st.Id = id
	if err := db.UpdateStation(a.DB, st); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(st)
}

func (a *App) DeleteStation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := db.DeleteStation(a.DB, id); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
