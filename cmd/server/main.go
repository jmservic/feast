package main

import (
	"context"
	//"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jmservic/feast/internal/database"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	platform := os.Getenv("PLATFORM")

	if platform != "test" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
		}
		platform = os.Getenv("PLATFORM")
		if platform == "" {
			platform = "production"
		}
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL must be set")
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

	/*handler.HandleFunc("/", func(w http.ResponseWriter, res *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World! because of course..."))
	})*/
	// Authentication and Authorization
	handler.HandleFunc("POST /api/login", cfg.handlerLogin)
	handler.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	handler.HandleFunc("POST /api/revoke", cfg.handlerRevoke)

	// Users
	handler.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	handler.Handle("PUT /api/users/{user_id}", cfg.middlewareAuthentication(cfg.handlerUpdateUser))

	// Households
	handler.Handle("POST /api/households", cfg.middlewareAuthentication(cfg.handlerCreateHousehold))
	handler.Handle("GET /api/households/{householdId}", cfg.middlewareAuthentication(cfg.handlerGetHousehold))
	handler.Handle("PUT /api/households/{householdId}", cfg.middlewareAuthentication(cfg.handlerUpdateHousehold))
	handler.Handle("DELETE /api/households/{householdId}", cfg.middlewareAuthentication(cfg.handlerDeleteHousehold))

	// Household Members
	//Might not need this one
	handler.Handle("GET /api/households/{household_id}/members", cfg.middlewareAuthentication(cfg.handlerGetHouseholdMembers))
	handler.Handle("POST /api/households/{household_id}/members/{member_id}", cfg.middlewareAuthentication(cfg.handlerCreateHouseholdMember))
	handler.Handle("GET /api/households/{household_id}/members/{member_id}", cfg.middlewareAuthentication(cfg.handlerGetHouseholdMember))
	handler.Handle("PUT /api/households/{household_id}/members/{member_id}", cfg.middlewareAuthentication(cfg.handlerUpdateHouseholdMember))
	handler.Handle("DELETE /api/households/{household_id}/members/{member_id}", cfg.middlewareAuthentication(cfg.handlerDeleteHouseholdMember))

	// Admin
	handler.HandleFunc("POST /admin/reset", cfg.handlerReset)
	/*handler.Handle("GET /authorized-endpoint", cfg.middlewareAuthentication(func(w http.ResponseWriter, res *http.Request, userId uuid.UUID) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World! because of course..."))
	}))*/

	server := http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 16,
	}

	log.Printf("Serving on port: %v\n", port)
	log.Fatalln(server.ListenAndServe())
}
