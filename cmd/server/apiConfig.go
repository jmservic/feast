package main

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmservic/feast/internal/auth"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"log"
	"net/http"
)

type apiConfig struct {
	db       *database.Queries
	conn     *pgx.Conn
	platform string
	secret   string
}

func mapDbErrorToHttpStatusCode(err error) int {
	pgErr := &pgconn.PgError{}
	code := http.StatusInternalServerError

	if errors.As(err, &pgErr) {
		log.Printf("Database error (%s) message: %s\n", pgErr.Code, pgErr.Detail)
		if pgErr.Code == "23505" || pgErr.Code == "23503" {
			code = http.StatusBadRequest
		}
		if pgErr.Code == "42501" || pgErr.Code == "P0001" {
			code = http.StatusForbidden
		}
		if pgErr.Code == "P0002" {
			code = http.StatusNotFound
		}
	} else {
		if err == pgx.ErrNoRows {
			code = http.StatusNotFound
		}
	}

	return code
}

func (cfg apiConfig) middlewareAuthentication(next func(http.ResponseWriter, *http.Request, uuid.UUID)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, constants.JwtRetrievalErrStr, err)
			return
		}

		userId, err := auth.ValidateJWT(token, cfg.secret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, constants.JwtValidationErrStr, err)
			return
		}

		next(w, r, userId)
	})
}
