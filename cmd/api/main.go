package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/matteobnvl/Platform-meteo/db"
	"github.com/matteobnvl/Platform-meteo/db/migrations"
)

func main() {
	db := initDB()
	defer db.Close()
	fmt.Println("Hello, World! API")
}

func initDB() *sql.DB {
	db, err := db.Connect()
	if err != nil {
		log.Fatalf("connexion BDD : %v", err)
	}

	if err := migrations.Run(db); err != nil {
		log.Fatalf("migrations : %v", err)
	}

	return db
}
