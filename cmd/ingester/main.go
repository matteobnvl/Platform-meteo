package main

import (
	"github.com/matteobnvl/Platform-meteo/db"
	"github.com/matteobnvl/Platform-meteo/internal/fetcher"
)

func main() {
	db := db.InitDB(false)
	defer db.Close()

	fetcher.Run(db)
}
