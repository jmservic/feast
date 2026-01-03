package main

import (
	"encoding/json"
	"github.com/jmservic/feast/internal/auth"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"
	"net/http"
)

func (cfg apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if params.Password == "" {
		respondWithError(w, http.StatusBadRequest, constants.EmptyPasswordErrStr, nil)
		return
	}
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.PasswordHashErrStr, err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Name:           params.Name,
		Email:          params.Email,
		HashedPassword: hash,
	})
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.UserCreationErrStr, err)
		return
	}

	rtnVals := dto.UserResources{
		Id:        user.ID,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	respondWithJSON(w, http.StatusCreated, rtnVals)
}

func (cfg apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {

}
