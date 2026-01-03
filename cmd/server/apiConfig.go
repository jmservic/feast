package main

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmservic/feast/internal/database"
	"net/http"
)

type apiConfig struct {
	db       *database.Queries
	platform string
}

func mapDbErrorToHttpStatusCode(err error) int {
	pgErr := &pgconn.PgError{}
	code := http.StatusInternalServerError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" || pgErr.Code == "23503" {
			code = http.StatusBadRequest
		}
	}

	return code
}
