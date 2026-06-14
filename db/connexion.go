package db

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/matteobnvl/Platform-meteo/db/migrations"
)

func connect() (*sql.DB, error) {
	dsn := "host=" + os.Getenv("DB_HOST") +
		" port=" + os.Getenv("DB_PORT") +
		" user=" + os.Getenv("DB_USER") +
		" password=" + os.Getenv("DB_PASSWORD") +
		" dbname=" + os.Getenv("DB_NAME") +
		" sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, err
}

func InitDB(runMigrations bool) *sql.DB {
	godotenv.Load()

	db, err := connect()
	if err != nil {
		slog.Error("connexion BDD", "err", err)
		os.Exit(1)
	}
	slog.Info("connexion BDD établie")

	if runMigrations {
		if err := migrations.Run(db); err != nil {
			slog.Error("migrations", "err", err)
			os.Exit(1)
		}
	}

	return db
}
