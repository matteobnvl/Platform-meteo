package db

import (
	"database/sql"
	"log"
)

func Connect() (*sql.DB, error) {
	dsn := "host=postgres port=5432 user=meteo_user password=meteo_password dbname=meteo_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db, err
}
