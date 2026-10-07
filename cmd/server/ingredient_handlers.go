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
	"strconv"
	"sync"
)

func (cfg apiConfig) handlerGetIngredients(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	query := r.FormValue("q")

	storedIngredients, err := cfg.db.GetIngredientsByQuery(r.Context(), database.GetIngredientsByQueryParams{
		Query:  query,
		UserID: &userId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetIngredientsErrStr, err)
		return
	}

	ingredients := make([]dto.Ingredient, 0, len(storedIngredients))
	ingredientIndexes := make(map[int64]int, len(storedIngredients))
	ingredientIds := make([]int, 0, len(storedIngredients))

	for index, storedIngredient := range storedIngredients {
		ingredient := dto.Ingredient{
			Amount:        decimal.NewFromInt(1),
			Name:          storedIngredient.Name,
			Fats:          storedIngredient.Fats,
			Carbohydrates: storedIngredient.Carbohydrates,
			Protein:       storedIngredient.Protein,
			Unittype:      dto.Measurement(storedIngredient.UnitType),
			Id:            storedIngredient.ID,
		}

		ingredients = append(ingredients, ingredient)
		ingredientIds = append(ingredientIds, int(ingredient.Id))
		ingredientIndexes[ingredient.Id] = index
	}

	servings, err := cfg.db.GetIngredientsServingsByIds(r.Context(), ingredientIds)

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetIngredientsServingsErrStr, err)
		return
	}

	for _, dbServing := range servings {
		index := ingredientIndexes[dbServing.IngredientID]
		serving := dto.ServingInfo{
			Id:       dbServing.ID,
			Name:     dbServing.Name,
			Quantity: dbServing.Quantity,
		}
		ingredients[index].Servings = append(ingredients[index].Servings, serving)
	}

	respondWithJSON(w, http.StatusOK, ingredients)
}

func (cfg apiConfig) handlerGetIngredient(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	ingredientId, err := strconv.ParseInt(r.PathValue("ingredient_id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidIntStringErrStr, err)
		return
	}

	wg := sync.WaitGroup{}
	errs := []error{}
	ingredient := dto.Ingredient{}

	wg.Go(func() {
		storedIngredient, err := cfg.db.GetIngredientById(r.Context(), ingredientId)
		if err != nil {
			errs = append(errs, err)
			return
		}

		ingredient.Amount = decimal.NewFromInt(1)
		ingredient.Name = storedIngredient.Name
		ingredient.Fats = storedIngredient.Fats
		ingredient.Carbohydrates = storedIngredient.Carbohydrates
		ingredient.Protein = storedIngredient.Protein
		ingredient.Unittype = dto.Measurement(storedIngredient.UnitType)
		ingredient.UserId = storedIngredient.UserID
		ingredient.Id = storedIngredient.ID
	})

	wg.Go(func() {
		dbServings, err := cfg.db.GetIngredientsServingsById(r.Context(), ingredientId)
		if err != nil {
			errs = append(errs, err)
		}

		servings := make([]dto.ServingInfo, 0, len(dbServings))

		for _, serving := range dbServings {
			servings = append(servings, dto.ServingInfo{
				Name:     serving.Name,
				Quantity: serving.Quantity,
				Id:       serving.ID,
			})
		}

		ingredient.Servings = servings
	})

	wg.Wait()

	//TODO: handle this in a way that utilizes both errors
	if len(errs) > 0 {
		respondWithError(w, mapDbErrorToHttpStatusCode(errs[0]), constants.GetIngredientErrStr, err)
		return
	}

	respondWithJSON(w, http.StatusOK, ingredient)
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

	if newIngredient.Protein.IsNegative() {
		respondWithError(w, http.StatusBadRequest, constants.InvalidProteinErrStr, nil)
		return
	}

	if newIngredient.Fats.IsNegative() {
		respondWithError(w, http.StatusBadRequest, constants.InvalidFatsErrStr, nil)
		return
	}

	if newIngredient.Carbohydrates.IsNegative() {
		respondWithError(w, http.StatusBadRequest, constants.InvalidCarbohydratesErrStr, nil)
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
