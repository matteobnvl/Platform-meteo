package main

import (
	"fmt"

	"github.com/matteobnvl/Platform-meteo/db"
)

func main() {
	db := db.InitDB(false)
	defer db.Close()
	fmt.Println("Hello, World! API")
}
