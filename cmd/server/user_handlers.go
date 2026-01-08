package main

import (
	"encoding/json"
	"github.com/jmservic/feast/internal/auth"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"
	"net/http"
	"time"
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
	params := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, constants.InvalidCredentialsErrStr, err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.HashCheckErrStr, err)
		return
	}
	if !match {
		respondWithError(w, http.StatusUnauthorized, constants.InvalidCredentialsErrStr, err)
		return
	}

	//Time to create the access and refresh tokens!!
	token, err := auth.MakeJWT(user.ID, cfg.secret, constants.AccessTokenLength)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JwtCreationErrStr, err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.RefreshTokenCreationErrStr, err)
		return
	}

	err = cfg.db.StoreRefreshToken(r.Context(), database.StoreRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(constants.RefreshTokenLength),
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.RefreshTokenStorageErrStr, err)
		return
	}

	respondWithJSON(w, http.StatusOK, dto.UserAuthentication{
		UserResources: dto.UserResources{
			Id:        user.ID,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

// What should we do for web browsers... cookies for access token / refresh token?
