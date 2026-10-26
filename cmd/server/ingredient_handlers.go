package main

import (
	"encoding/json"
	//	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"
	"github.com/shopspring/decimal"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
)

func (cfg apiConfig) handlerGetIngredients(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	query := r.FormValue("q")
	queryParts := strings.Fields(strings.TrimSpace(query))
	lastIndex := len(queryParts) - 1
	tsQuery := ""
	if lastIndex >= 0 {
		queryParts[lastIndex] = queryParts[lastIndex] + ":*"
		tsQuery = strings.Join(queryParts, " & ")
		//		fmt.Println(tsQuery)
	}

	storedIngredients, err := cfg.db.GetIngredientsByQuery(r.Context(), database.GetIngredientsByQueryParams{
		Query:  tsQuery,
		UserID: &userId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetIngredientsErrStr, err)
		return
	}

	if len(storedIngredients) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ingredients := make([]dto.Ingredient, 0, len(storedIngredients))
	ingredientIndexes := make(map[int64]int, len(storedIngredients))
	ingredientIds := make([]int, 0, len(storedIngredients))

	for index, storedIngredient := range storedIngredients {
		var brand string
		if storedIngredient.Brand != nil {
			brand = *storedIngredient.Brand
		}
		ingredient := dto.Ingredient{
			Amount:        decimal.NewFromInt(1),
			Name:          storedIngredient.Name,
			Fats:          storedIngredient.Fats,
			Carbohydrates: storedIngredient.Carbohydrates,
			Protein:       storedIngredient.Protein,
			Calories:      storedIngredient.Calories,
			Unittype:      dto.Measurement(storedIngredient.UnitType),
			Id:            storedIngredient.ID,
			Brand:         brand,
			UserId:        storedIngredient.UserID,
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
		brand := ""
		if storedIngredient.Brand != nil {
			brand = *storedIngredient.Brand
		}

		ingredient.Amount = decimal.NewFromInt(1)
		ingredient.Name = storedIngredient.Name
		ingredient.Fats = storedIngredient.Fats
		ingredient.Carbohydrates = storedIngredient.Carbohydrates
		ingredient.Protein = storedIngredient.Protein
		ingredient.Calories = storedIngredient.Calories
		ingredient.Unittype = dto.Measurement(storedIngredient.UnitType)
		ingredient.Brand = brand
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
	tx, err := cfg.pool.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
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
	ingredientId, err := strconv.ParseInt(r.PathValue("ingredient_id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidIntStringErrStr, err)
		return
	}

	storedIngredient, err := cfg.db.GetIngredientById(r.Context(), ingredientId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetIngredientErrStr, err)
		return
	}

	if storedIngredient.UserID == nil || *storedIngredient.UserID != userId {
		respondWithError(w, http.StatusForbidden, constants.NotAuthorizedErrStr, nil)
		return

	}

	storedServings, err := cfg.db.GetIngredientsServingsById(r.Context(), ingredientId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetIngredientServingsErrStr, err)
		return
	}
	if len(storedServings) == 0 {
		respondWithError(w, http.StatusNotFound, constants.GetIngredientServingsErrStr, nil)
		return
	}

	params := struct {
		Name            *string           `json:"name"`
		Fats            *decimal.Decimal  `json:"fats"`
		Carbohydrates   *decimal.Decimal  `json:"carbohydrates"`
		Protein         *decimal.Decimal  `json:"protein"`
		Calories        *decimal.Decimal  `json:"calories"`
		Brand           *string           `json:"brand"`
		Amount          *decimal.Decimal  `json:"amount"`
		Unittype        *dto.Measurement  `json:"unittype"`
		Servings        []dto.ServingInfo `json:"servings"`
		RemovedServings []int16           `json:"removed_servings"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	//Check parameters
	if (params.Fats != nil || params.Carbohydrates != nil || params.Protein != nil || params.Calories != nil) && params.Amount == nil {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, nil)
		return
	}

	if params.Amount != nil && !params.Amount.IsPositive() {
		respondWithError(w, http.StatusBadRequest, constants.InvalidAmountErrStr, nil)
		return
	}

	updateParams := database.UpdateIngredientParams{
		ID: ingredientId}

	if params.Name != nil {
		if *params.Name == "" {
			respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, nil)
			return
		}
		updateParams.Name = *params.Name
	} else {
		updateParams.Name = storedIngredient.Name
	}

	if params.Fats != nil {
		if params.Fats.IsNegative() {
			respondWithError(w, http.StatusBadRequest, constants.InvalidFatsErrStr, nil)
			return
		}
		updateParams.Fats = params.Fats.Div(*params.Amount)
	} else {
		updateParams.Fats = storedIngredient.Fats
	}

	if params.Protein != nil {
		if params.Protein.IsNegative() {
			respondWithError(w, http.StatusBadRequest, constants.InvalidProteinErrStr, nil)
			return
		}
		updateParams.Protein = params.Protein.Div(*params.Amount)
	} else {
		updateParams.Protein = storedIngredient.Protein
	}

	if params.Carbohydrates != nil {
		if params.Carbohydrates.IsNegative() {
			respondWithError(w, http.StatusBadRequest, constants.InvalidCarbohydratesErrStr, nil)
			return

		}
		updateParams.Carbohydrates = params.Carbohydrates.Div(*params.Amount)
	} else {
		updateParams.Carbohydrates = storedIngredient.Carbohydrates
	}

	if params.Calories != nil {
		if params.Calories.IsNegative() {
			respondWithError(w, http.StatusBadRequest, constants.InvalidCalorieErrStr, nil)
			return
		}
		updateParams.Calories = params.Calories.Div(*params.Amount)
	} else {
		updateParams.Calories = storedIngredient.Calories
	}

	if params.Brand != nil {
		updateParams.Brand = params.Brand
	} else {
		updateParams.Brand = storedIngredient.Brand
	}

	if params.Unittype != nil {
		updateParams.UnitType = bool(*params.Unittype)
	} else {
		updateParams.UnitType = storedIngredient.UnitType
	}

	finalServingsState := map[int16]dto.ServingInfo{}

	//Add stored Servings
	for _, serving := range storedServings {
		serv := dto.ServingInfo{
			Id:       serving.ID,
			Name:     serving.Name,
			Quantity: serving.Quantity,
		}

		finalServingsState[serv.Id] = serv
	}

	//Add updated and new Servings
	newServingsCounter := int16(0)
	for _, serving := range params.Servings {
		storageId := serving.Id
		if serving.Id <= 0 {
			storageId = newServingsCounter
			newServingsCounter--
		}
		finalServingsState[storageId] = serving
	}

	//Remove deleted Servings
	for _, id := range params.RemovedServings {
		delete(finalServingsState, id)
	}

	numOfFinalServings := len(finalServingsState)

	//Validate the New Servings
	if numOfFinalServings == 0 {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, nil)
		return
	}

	for _, serving := range finalServingsState {

		if serving.Quantity.IsNegative() {
			respondWithError(w, http.StatusBadRequest, constants.ServingsCreationErrStr, nil)
			return
		}

		if serving.Quantity.IsZero() && numOfFinalServings > 1 {
			respondWithError(w, http.StatusBadRequest, constants.InvalidMultipleServingsErrStr, nil)
			return
		}
	}

	rtnServings := []dto.ServingInfo{}

	//update Ingredient
	updatedIngredient, err := cfg.db.UpdateIngredient(r.Context(), updateParams)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.IngredientUpdateErrStr, err)
		return
	}

	//update Servings
	updatedServings := make([]dto.ServingInfo, 0, len(finalServingsState))
	for _, serving := range finalServingsState {
		updatedServings = append(updatedServings, serving)
	}

	err = cfg.db.CreateIngredientServings(r.Context(), database.CreateIngredientServingsParams{
		IngredientID: ingredientId,
		Servings:     updatedServings,
	})
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.IngredientUpdateErrStr, err)
		return
	}

	//delete Servings
	err = cfg.db.DeleteIngredientServings(r.Context(), database.DeleteIngredientServingsParams{
		IngredientID: ingredientId,
		Ids:          params.RemovedServings,
	})
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.IngredientUpdateErrStr, err)
		return

	}

	servings, err := cfg.db.GetIngredientsServingsById(r.Context(), ingredientId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.IngredientUpdateErrStr, err)
		return
	}

	for _, serving := range servings {
		rtnServings = append(rtnServings, dto.ServingInfo{
			Id:       serving.ID,
			Name:     serving.Name,
			Quantity: serving.Quantity,
		})
	}

	brandRtnVal := ""
	if updatedIngredient.Brand != nil {
		brandRtnVal = *updatedIngredient.Brand
	}

	respondWithJSON(w, http.StatusOK, dto.Ingredient{
		Id:            updatedIngredient.ID,
		Name:          updatedIngredient.Name,
		Fats:          updatedIngredient.Fats,
		Carbohydrates: updatedIngredient.Carbohydrates,
		Protein:       updatedIngredient.Protein,
		Calories:      updatedIngredient.Calories,
		Amount:        decimal.NewFromInt(1),
		Unittype:      dto.Measurement(updatedIngredient.UnitType),
		Brand:         brandRtnVal,
		UserId:        updatedIngredient.UserID,
		Servings:      rtnServings,
	})
}

func (cfg apiConfig) handlerDeleteIngredient(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	ingredientId, err := strconv.ParseInt(r.PathValue("ingredient_id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidIntStringErrStr, err)
		return
	}

	storedIngredient, err := cfg.db.GetIngredientById(r.Context(), ingredientId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetIngredientErrStr, err)
		return
	}

	if storedIngredient.UserID == nil || *storedIngredient.UserID != userId {
		respondWithError(w, http.StatusForbidden, constants.NotAuthorizedErrStr, nil)
		return
	}

	err = cfg.db.DeleteIngredient(r.Context(), ingredientId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetIngredientErrStr, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
