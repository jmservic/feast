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

func (cfg apiConfig) handlerCreateHouseholdMember(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	householdId, err := uuid.Parse(r.PathValue("household_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	params := struct {
		Name   string    `json:"name"`
		UserId uuid.UUID `json:"user_id"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, err)
		return
	}

	err = cfg.db.CreateHouseholdMember(r.Context(), database.CreateHouseholdMemberParams{
		CreatorID:   userId,
		MemberName:  params.Name,
		UserID:      params.UserId,
		HouseholdID: householdId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberCreateErrStr, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (cfg apiConfig) handlerUpdateHouseholdMember(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	householdId, err := uuid.Parse(r.PathValue("household_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	memberId, err := uuid.Parse(r.PathValue("member_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}
	//can change name, role, and userid

	params := struct {
		Name   string    `json:"name"`
		UserId uuid.UUID `json:"user_id"`
		Role   int       `json:"role"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, err)
		return
	}

	err = cfg.db.UpdateHouseholdMember(r.Context(), database.UpdateHouseholdMemberParams{
		Name:   params.Name,
		UserId: params.UserId,
		Role:   params.Role,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberUpdateErrStr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (cfg apiConfig) handlerGetHouseholdMember(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	householdId, err := uuid.Parse(r.PathValue("household_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	memberId, err := uuid.Parse(r.PathValue("member_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	err, memberInfo := cfg.db.GetHouseholdMember(r.Context())
}

func (cfg apiConfig) handlerGetHouseholdMembers(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}

func (cfg apiConfig) handlerDeleteHouseholdMember(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}
