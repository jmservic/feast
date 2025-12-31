package main

import (
	"github.com/jmservic/feast/internal/database"
	"net/http"
)

type apiConfig struct {
	db *database.Queries
}

func (cfg apiConfig) register(w http.ResponseWriter, res *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("You've registered!!!"))
}
