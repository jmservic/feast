package main

import (
	"encoding/json"
	"github.com/google/uuid"
	//	"github.com/jmservic/feast/internal/auth"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	//	"github.com/jmservic/feast/internal/dto"
	"net/http"
	// "time"
)

func (cfg apiConfig) handlerCreateHousehold(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	params := struct {
		Name string `json:"name"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, nil)
	}

	err := cfg.db.CreateHousehold(r.Context(), database.CreateHouseholdParams{
		Name:   params.Name,
		UserID: userId,
	})
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdCreationErrStr, err)
		return
	}

}

func (cfg apiConfig) handlerUpdateHousehold(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}

func (cfg apiConfig) handlerGetHousehold(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}

func (cfg apiConfig) handlerDeleteHousehold(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}
