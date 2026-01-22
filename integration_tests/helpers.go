package integration

import (
	"bytes"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func loadDotEnv() {
	platform := os.Getenv("PLATFORM")

	if platform != "test" {
		err := godotenv.Load("../.env")
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
		}
	}
}

func resetDatabase(feast_url string) {
	res, err := http.Post(feast_url+"/admin/reset", "application/json", bytes.NewReader([]byte("")))
	if res.StatusCode != http.StatusOK {
		log.Printf("Error resetting the database, got the following status code: %d\n", res.StatusCode)
	}
	if err != nil {
		log.Printf("Error with post request: %v\n", err)
	}
	res.Body.Close()
}
