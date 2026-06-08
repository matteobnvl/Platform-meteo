package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/matteobnvl/Platform-meteo/db"
)

func main() {
	database := db.InitDB(false)
	defer database.Close()

	app := &App{db: database}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /stations", app.listStations)
	mux.HandleFunc("GET /stations/{id}", app.getStation)
	mux.HandleFunc("GET /stations/{id}/observations", app.listObservations)
	mux.HandleFunc("POST /stations", app.createStation)
	mux.HandleFunc("PUT /stations/{id}", app.updateStation)
	mux.HandleFunc("DELETE /stations/{id}", app.deleteStation)

	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
