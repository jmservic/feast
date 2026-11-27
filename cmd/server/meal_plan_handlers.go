package main

import (
	"encoding/json"
	//	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	//	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	//	"github.com/jmservic/feast/internal/dto"
	"github.com/shopspring/decimal"
	"net/http"
	//"time"
	/*"slices"
	"strconv"
	"strings"
	"sync "*/)

func (cfg apiConfig) handlerGetMealPlan(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}

func (cfg apiConfig) handlerCreateMealPlan(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	params := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Items       []struct {
			Id             uuid.UUID       `json:"id"`
			Day            int             `json:"day"`
			Category       int             `json:"category"`
			Amount         decimal.Decimal `json:"amount"`
			ServingsName   string          `json:"servings_name"`
			ServingsAmount string          `json:"servings_amount"`
		} `json:"items"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr)
		return
	}

	// create a repeatable read transaction
	tx, err := cfg.pool.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.BeginTransactionErrStr, err)
		return
	}
	defer tx.Rollback(r.Context())

	queries := cfg.db.WithTx(tx)

	storedMealPlan, err := queries.CreateMealPlan(r.Context(), database.CreateMealPlanParams{
		Name:        params.Name,
		Description: params.Description,
		UserID:      userId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.MealPlanCreationErrStr)
		return
	}

}

func (cfg apiConfig) handlerUpdateMealPlan(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}

func (cfg apiConfig) handlerDeleteMealPlan(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}
