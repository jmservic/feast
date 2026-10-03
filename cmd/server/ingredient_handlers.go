package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"
	"github.com/shopspring/decimal"
	"net/http"
	"slices"
)

func (cfg apiConfig) handlerGetIngredients(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}

func (cfg apiConfig) handlerGetIngredient(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}

func (cfg apiConfig) handlerCreateIngredient(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

	newIngredient := dto.Ingredient{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newIngredient); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if newIngredient.Name == "" {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, nil)
		return
	}

	if !newIngredient.Amount.IsPositive() {
		respondWithError(w, http.StatusBadRequest, constants.InvalidAmountErrStr, nil)
		return

	}

	if newIngredient.Calories.IsNegative() {
		respondWithError(w, http.StatusBadRequest, constants.InvalidCalorieErrStr, nil)
		return
	}

	if len(newIngredient.Servings) == 0 {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, nil)
		return
	}

	for _, servings := range newIngredient.Servings {
		if servings.Quantity.IsNegative() {
			respondWithError(w, http.StatusBadRequest, constants.InvalidServingsErrStr, nil)
			return
		}
		if len(newIngredient.Servings) > 1 && servings.Quantity.IsZero() {
			respondWithError(w, http.StatusBadRequest, constants.InvalidMultipleServingsErrStr, nil)
			return
		}
	}

	// create a repeatable read transaction
	tx, err := cfg.conn.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.BeginTransactionErrStr, err)
		return
	}
	defer tx.Rollback(r.Context())

	queries := cfg.db.WithTx(tx)

	var (
		brandStoreVal *string
	)
	if newIngredient.Brand != "" {
		brandStoreVal = &newIngredient.Brand
	}

	storedIngredient, err := queries.CreateIngredient(r.Context(), database.CreateIngredientParams{
		Name:          newIngredient.Name,
		Fats:          newIngredient.Fats.Div(newIngredient.Amount),
		Carbohydrates: newIngredient.Carbohydrates.Div(newIngredient.Amount),
		Protein:       newIngredient.Protein.Div(newIngredient.Amount),
		Calories:      newIngredient.Calories.Div(newIngredient.Amount),
		UnitType:      bool(newIngredient.Unittype),
		Brand:         brandStoreVal,
		UserID:        &userId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.IngredientCreationErrStr, err)
		return
	}

	slices.Reverse(newIngredient.Servings) //currently the servings are being stored backwards.. weird

	err = queries.CreateIngredientServings(r.Context(), database.CreateIngredientServingsParams{
		IngredientID: storedIngredient.ID,
		Servings:     newIngredient.Servings,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.ServingsCreationErrStr, err)
		return
	}

	dbServings, err := queries.GetIngredientsServingsById(r.Context(), storedIngredient.ID)

	if err != nil || len(dbServings) == 0 {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetIngredientServingsErrStr, err)
		return
	}

	servings := make([]dto.ServingInfo, 0, len(dbServings))

	for _, serving := range dbServings {
		servings = append(servings, dto.ServingInfo{
			Name:     serving.Name,
			Quantity: serving.Quantity,
			Id:       serving.ID,
		})
	}

	err = tx.Commit(r.Context())
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.CommitTransactionErrStr, err)
		return
	}

	brandRtnVal := ""

	if storedIngredient.Brand != nil {
		brandRtnVal = *storedIngredient.Brand
	}

	respondWithJSON(w, http.StatusCreated, dto.Ingredient{
		Id:            storedIngredient.ID,
		Name:          storedIngredient.Name,
		Fats:          storedIngredient.Fats,
		Carbohydrates: storedIngredient.Carbohydrates,
		Protein:       storedIngredient.Protein,
		Calories:      storedIngredient.Calories,
		Amount:        decimal.NewFromInt(1),
		Unittype:      dto.Measurement(storedIngredient.UnitType),
		Brand:         brandRtnVal,
		UserId:        storedIngredient.UserID,
		Servings:      servings,
	})
}

func (cfg apiConfig) handlerUpdateIngredient(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}

func (cfg apiConfig) handlerDeleteIngredient(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}
