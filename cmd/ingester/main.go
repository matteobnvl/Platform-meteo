package main

import (
	"github.com/matteobnvl/Platform-meteo/db"
	"github.com/matteobnvl/Platform-meteo/internal/fetcher"
)

func main() {
	db := db.InitDB(true)
	defer db.Close()

	fetcher.FetchSynop(db)
}
