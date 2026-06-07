package main

import (
	"github.com/matteobnvl/Platform-meteo/db"
)

func main() {
	db := db.InitDB(true)
	defer db.Close()

	// fetcher.FetchSynop(db)
}
