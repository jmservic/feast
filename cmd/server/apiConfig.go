package main

import (
	"github.com/jmservic/feast/internal/auth"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	//"github.com/jmservic/feast/internal/dto"
	"encoding/json"
	"net/http"
)

type apiConfig struct {
	db *database.Queries
}

func (cfg apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("You've registered!!!"))
}
