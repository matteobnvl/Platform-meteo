package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/matteobnvl/Platform-meteo/db/migrations"
)

func connect() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

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
		log.Fatalf("connexion BDD : %v", err)
	} else {
		fmt.Println("Connexion database successful")
	}

	if runMigrations {
		if err := migrations.Run(db); err != nil {
			log.Fatalf("migrations : %v", err)
		}
	}

	return db
}
