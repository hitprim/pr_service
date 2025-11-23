package main

import (
	"log"
	"pr_service/internal/db"
	"pr_service/internal/router"
)

func main() {
	database := db.Connect()
	r := router.SetupRouter(database)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
