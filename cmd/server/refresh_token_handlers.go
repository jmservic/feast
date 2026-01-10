package main

import (
	"github.com/jackc/pgx/v5"
	"github.com/jmservic/feast/internal/auth"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"
	"net/http"
	"time"
)

func (cfg apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.RefreshTokenRetrievalErrStr, err)
		return
	}

	//When checking the expiration, check for the error pgx.ErrNoRows, if we get that return 401 else it is a server error.
	isExpired, err := cfg.db.RefreshTokenExpired(r.Context(), refreshToken)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, constants.RefreshTokenInvalidErrStr, err)
			return
		} else {

			respondWithError(w, http.StatusInternalServerError, constants.RefreshTokenCheckingErrStr, err)
			return
		}
	}
	if isExpired.Bool {
		respondWithError(w, http.StatusUnauthorized, constants.RefreshTokenInvalidErrStr, nil)
		return
	}

	userId, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.UserIdRetrievalErrStr, err)
		return
	}

	token, err := auth.MakeJWT(userId, cfg.secret, constants.AccessTokenLength)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JwtCreationErrStr, err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.RefreshTokenRevokeErrStr, err)
		return
	}

	newRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.RefreshTokenCreationErrStr, err)
		return
	}

	err = cfg.db.StoreRefreshToken(r.Context(), database.StoreRefreshTokenParams{
		Token:     newRefreshToken,
		UserID:    userId,
		ExpiresAt: time.Now().Add(constants.RefreshTokenLength),
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.RefreshTokenStorageErrStr, err)
		return
	}

	respondWithJSON(w, http.StatusOK, dto.TokenResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
	})
}

func (cfg apiConfig) handlerRevoke(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.RefreshTokenRetrievalErrStr, err)
		return
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.RefreshTokenRevokeErrStr, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
