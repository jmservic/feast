package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jmservic/feast/internal/auth"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"
	"net/http"
	"strings"
	"time"
)

func (cfg apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"` // #nosec G117
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
		Email:          strings.ToLower(params.Email),
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

func (cfg apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	params := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"` // #nosec G117
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}
	params.Email = strings.ToLower(params.Email)

	// Need to get the user by id
	user, err := cfg.db.GetUserById(r.Context(), userId)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, constants.InvalidCredentialsErrStr, err)
		return
	}

	// Check if the password or email is different
	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.HashCheckErrStr, err)
		return
	}

	// Hash the password
	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.PasswordHashErrStr, err)
		return
	}

	// update the user
	updated_user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Name:           params.Name,
		Email:          params.Email,
		HashedPassword: hash,
		ID:             userId,
	})
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.UserUpdateErrStr, err)
		return
	}

	if !match || (user.Email != params.Email) {
		// Revoke refresh tokens
		err = cfg.db.RevokeUserRefreshTokens(r.Context(), userId)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, constants.RefreshTokenRevokeErrStr, err)
			return
		}
	}

	//Time to create new access and refresh tokens!!
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
			Id:        updated_user.ID,
			Name:      updated_user.Name,
			CreatedAt: updated_user.CreatedAt,
			UpdatedAt: updated_user.UpdatedAt,
			Email:     updated_user.Email,
		},
		TokenResponse: dto.TokenResponse{
			Token:        token,
			RefreshToken: refreshToken,
		},
	})
}

func (cfg apiConfig) handlerDeleteUser(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	err := cfg.db.DeleteUser(r.Context(), userId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.UserDeleteErrStr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (cfg apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	params := struct {
		Email    string `json:"email"`
		Password string `json:"password"` // #nosec G117
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), strings.ToLower(params.Email))
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
		TokenResponse: dto.TokenResponse{
			Token:        token,
			RefreshToken: refreshToken,
		},
	})
}

// What should we do for web browsers... cookies for access token / refresh token?
