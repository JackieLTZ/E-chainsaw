package main

import (
	"a/cmd/api"
	"log"
	"os"
)

func main() {

	db, err := DB_Connect()
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
		log.Println("Database connection closed.")
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := api.Mount(db)

	err = api.RunServer(port, mux)
	if err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}

}
