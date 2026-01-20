package integration

import (
	"github.com/joho/godotenv"
	"log"
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
