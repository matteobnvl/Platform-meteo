package main

import (
	"log/slog"
	"time"

	"github.com/matteobnvl/Platform-meteo/db"
	"github.com/matteobnvl/Platform-meteo/internal/detector"
	"github.com/matteobnvl/Platform-meteo/internal/fetcher"
)

func main() {
	database := db.InitDB(true)
	defer database.Close()

	cfg := detector.LoadConfig()

	// lancement immédiat au démarrage
	fetcher.FetchSynop(database)
	detector.RunOnce(database, cfg)

	// ingestion quotidienne
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			slog.Info("ingestion périodique démarrée")
			fetcher.FetchSynop(database)
		}
	}()

	// détection toutes les 10 minutes
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.IntervalMinutes) * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			slog.Info("détection périodique démarrée")
			detector.RunOnce(database, cfg)
		}
	}()

	slog.Info("ingester en marche", "fetcher_interval", "24h", "detector_interval", "10min")
	select {}
}
