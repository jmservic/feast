package main

import (
	"encoding/json"
	"github.com/google/uuid"
	//	"github.com/jmservic/feast/internal/auth"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"
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

	household, err := cfg.db.GetHouseholdByUserId(r.Context(), userId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdRetrievalByUserErrStr, err)
		return
	}

	rtnVals := dto.HouseholdResources{
		Id:        household.ID,
		CreatedAt: household.CreatedAt,
		UpdatedAt: household.UpdatedAt,
		Name:      household.Name,
	}
	respondWithJSON(w, http.StatusCreated, rtnVals)

}

func (cfg apiConfig) handlerUpdateHousehold(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
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
	householdId, err := uuid.Parse(r.PathValue("household_id"))
	if err != nil {
		household, err := cfg.db.GetHouseholdByUserId(r.Context(), userId)
		if err != nil {
			respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdRetrievalByUserErrStr, err)
			return
		}
		householdId = household.ID
	}

	err = cfg.db.UpdateHousehold(r.Context(), database.UpdateHouseholdParams{
		UserID:      userId,
		HouseholdID: householdId,
		NewName:     params.Name,
	})
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdUpdateErrStr, err)
		return
	}

	updatedHousehold, err := cfg.db.GetHouseholdById(r.Context(), householdId)

	rtnVals := dto.HouseholdResources{
		Id:        updatedHousehold.ID,
		CreatedAt: updatedHousehold.CreatedAt,
		UpdatedAt: updatedHousehold.UpdatedAt,
		Name:      updatedHousehold.Name,
	}
	respondWithJSON(w, http.StatusOK, rtnVals)
}

func (cfg apiConfig) handlerGetHousehold(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	var household database.Household

	householdId, err := uuid.Parse(r.PathValue("household_id"))
	if err != nil {
		household, err = cfg.db.GetHouseholdByUserId(r.Context(), userId)
		if err != nil {
			respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdRetrievalByUserErrStr, err)
			return
		}
	} else {
		household, err = cfg.db.GetHouseholdByUserId(r.Context(), householdId)
		if err != nil {
			respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdRetrievalByIdErrStr, err)
			return
		}
	}

	rtnVals := dto.HouseholdResources{
		Id:        household.ID,
		CreatedAt: household.CreatedAt,
		UpdatedAt: household.UpdatedAt,
		Name:      household.Name,
	}
	respondWithJSON(w, http.StatusOK, rtnVals)

}

func (cfg apiConfig) handlerDeleteHousehold(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}
