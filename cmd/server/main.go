package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jmservic/feast/internal/database"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL must be set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		platform = "production"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		log.Fatalf("Error opening the database: %s", err)
	}

	dbQueries := database.New(conn)

	cfg := apiConfig{
		db:       dbQueries,
		platform: platform,
		secret:   jwtSecret,
	}

	handler := http.NewServeMux()

	handler.HandleFunc("/", func(w http.ResponseWriter, res *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World! because of course..."))
	})

	handler.HandleFunc("POST /register", cfg.handlerCreateUser)
	handler.HandleFunc("POST /login", cfg.handlerLogin)
	handler.HandleFunc("POST /admin/reset", cfg.handlerReset)

	server := http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	log.Printf("Serving on port: %v\n", port)
	log.Fatalln(server.ListenAndServe())
}
