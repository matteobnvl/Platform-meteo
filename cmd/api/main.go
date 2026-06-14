package main

import (
	_ "embed"
	"log/slog"
	"net/http"
	"os"

	"github.com/matteobnvl/Platform-meteo/db"
	"github.com/matteobnvl/Platform-meteo/internal/handler"
)

//go:embed static/swagger.yaml
var swaggerYAML []byte

//go:embed static/swagger.html
var swaggerHTML []byte

func main() {
	database := db.InitDB(false)
	defer database.Close()

	app := &handler.App{DB: database}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok\n"))
	})

	// docs
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(swaggerHTML)
	})
	mux.HandleFunc("GET /docs/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(swaggerYAML)
	})

	// stations
	mux.HandleFunc("GET /stations", app.ListStations)
	mux.HandleFunc("POST /stations", app.CreateStation)
	mux.HandleFunc("GET /stations/{id}", app.GetStation)
	mux.HandleFunc("PUT /stations/{id}", app.UpdateStation)
	mux.HandleFunc("DELETE /stations/{id}", app.DeleteStation)

	// observations
	mux.HandleFunc("GET /stations/{id}/observations", app.ListObservations)
	mux.HandleFunc("GET /observations/aggregate", app.AggregateObservations)

	// events
	mux.HandleFunc("GET /events", app.ListEvents)
	mux.HandleFunc("GET /events/stats", app.EventStats)
	mux.HandleFunc("GET /events/{id}", app.GetEvent)
	mux.HandleFunc("GET /stations/{id}/events", app.ListStationEvents)

	slog.Info("serveur démarré", "port", 8080)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		slog.Error("serveur arrêté", "err", err)
		os.Exit(1)
	}
}
