package main

import (
	"encoding/json"
	//	"fmt"
	"github.com/google/uuid"
	//"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
	/*"slices"
	"strconv"
	"strings"
	"sync "*/)

//FoodLog, DailyLog, FoodItem

func (cfg apiConfig) handlerGetFoodItem(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	foodItemId, err := uuid.Parse(r.PathValue("log_item_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	storedFoodItem, err := cfg.db.GetFoodItemById(r.Context(), foodItemId)

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetFoodItemErrStr, err)
		return
	}

	if storedFoodItem.UserID != userId {
		respondWithError(w, http.StatusForbidden, constants.NotAuthorizedErrStr, nil)
		return
	}

	var brand string
	if storedFoodItem.Brand != nil {
		brand = *storedFoodItem.Brand
	}

	respondWithJSON(w, http.StatusOK, dto.FoodItem{
		Id:             storedFoodItem.ID,
		Date:           storedFoodItem.Date.Time,
		Category:       dto.Category(storedFoodItem.Category),
		Name:           storedFoodItem.Name,
		Fats:           storedFoodItem.Fats,
		Carbohydrates:  storedFoodItem.Carbohydrates,
		Protein:        storedFoodItem.Protein,
		Calories:       storedFoodItem.Calories,
		Brand:          brand,
		UnitType:       dto.Measurement(storedFoodItem.UnitType),
		Amount:         storedFoodItem.Quantity,
		ServingsName:   storedFoodItem.ServingsName,
		ServingsAmount: storedFoodItem.ServingsQuantity,
		UserId:         userId,
	})
}

func (cfg apiConfig) handlerGetLogs(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	// Need name, date, category, fats, protein, carbs, calories, unit type, brand, ingredient id, quantity,serving name, serving quantity
}

func (cfg apiConfig) handlerCreateFoodItem(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	itemInfo := struct {
		Id               uuid.UUID       `json:"id"`
		Date             time.Time       `json:"date"`
		Category         dto.Category    `json:"category"`
		Amount           decimal.Decimal `json:"amount"`
		ServingsName     string          `json:"servings_name"`
		AmountInServings decimal.Decimal `json:"amount_in_servings"`
	}{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&itemInfo); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if !itemInfo.Amount.IsPositive() || itemInfo.AmountInServings.IsNegative() {
		respondWithError(w, http.StatusBadRequest, constants.InvalidAmountErrStr, nil)
		return
	}

	if itemInfo.ServingsName == "" {
		respondWithError(w, http.StatusBadRequest, constants.InvalidServingsErrStr, nil)
		return
	}

	ingredient, err := cfg.db.GetIngredientById(r.Context(), itemInfo.Id)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetIngredientErrStr, err)
		return
	}
	date := pgtype.Date{}
	err = date.Scan(itemInfo.Date)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.ConvertTimeToDbDateErrStr, err)
		return
	}

	storedFoodItem, err := cfg.db.CreateFoodItem(r.Context(), database.CreateFoodItemParams{
		Name:             ingredient.Name,
		Date:             date,
		Category:         int(itemInfo.Category),
		Fats:             ingredient.Fats.Mul(itemInfo.Amount),
		Carbohydrates:    ingredient.Carbohydrates.Mul(itemInfo.Amount),
		Protein:          ingredient.Protein.Mul(itemInfo.Amount),
		Calories:         ingredient.Calories.Mul(itemInfo.Amount),
		UnitType:         ingredient.UnitType,
		Brand:            ingredient.Brand,
		IngredientsID:    &ingredient.ID,
		Quantity:         itemInfo.Amount,
		ServingsName:     itemInfo.ServingsName,
		ServingsQuantity: itemInfo.AmountInServings,
		UserID:           userId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.FoodItemCreationErrStr, err)
		return
	}

	var brand string
	if storedFoodItem.Brand != nil {
		brand = *storedFoodItem.Brand
	}

	respondWithJSON(w, http.StatusCreated, dto.FoodItem{
		Id:             storedFoodItem.ID,
		Date:           storedFoodItem.Date.Time,
		Category:       dto.Category(storedFoodItem.Category),
		Name:           storedFoodItem.Name,
		Fats:           storedFoodItem.Fats,
		Carbohydrates:  storedFoodItem.Carbohydrates,
		Protein:        storedFoodItem.Protein,
		Calories:       storedFoodItem.Calories,
		Brand:          brand,
		UnitType:       dto.Measurement(storedFoodItem.UnitType),
		Amount:         storedFoodItem.Quantity,
		ServingsName:   storedFoodItem.ServingsName,
		ServingsAmount: storedFoodItem.ServingsQuantity,
		UserId:         userId,
	})
}

func (cfg apiConfig) handlerUpdateFoodItem(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	foodItemId, err := uuid.Parse(r.PathValue("log_item_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	storedFoodItem, err := cfg.db.GetFoodItemById(r.Context(), foodItemId)

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.GetFoodItemErrStr, err)
		return
	}

	if storedFoodItem.UserID != userId {
		respondWithError(w, http.StatusForbidden, constants.NotAuthorizedErrStr, nil)
		return
	}

	itemUpdateInfo := struct {
		Id               uuid.UUID        `json:"id"`
		Date             *time.Time       `json:"date"`
		Category         *dto.Category    `json:"category"`
		Amount           *decimal.Decimal `json:"amount"`
		ServingsName     *string          `json:"servings_name"`
		AmountInServings *decimal.Decimal `json:"amount_in_servings"`
	}{}

	foodItemUpdateParams := database.UpdateFoodItemParams{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&itemUpdateInfo); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if itemUpdateInfo.Date != nil {
		date := pgtype.Date{}
		err = date.Scan(itemUpdateInfo.Date)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, constants.ConvertTimeToDbDateErrStr, err)
			return
		}

		foodItemUpdateParams.Date = date
	} else {
		foodItemUpdateParams.Date = storedFoodItem.Date
	}

	if itemUpdateInfo.Category != nil {
		foodItemUpdateParams.Category = int(*itemUpdateInfo.Category)
	} else {
		foodItemUpdateParams.Category = storedFoodItem.Category
	}

	if itemUpdateInfo.Amount != nil && itemUpdateInfo.Amount.IsPositive() && *itemUpdateInfo.Amount != storedFoodItem.Quantity {
		oldAmount := foodItemUpdateParams.Quantity
		newAmount := *itemUpdateInfo.Amount
		foodItemUpdateParams.Fats = storedFoodItem.Fats.Div(oldAmount).Mul(newAmount)
		foodItemUpdateParams.Carbohydrates = storedFoodItem.Carbohydrates.Div(oldAmount).Mul(newAmount)
		foodItemUpdateParams.Protein = storedFoodItem.Protein.Div(oldAmount).Mul(newAmount)
		foodItemUpdateParams.Calories = storedFoodItem.Calories.Div(oldAmount).Mul(newAmount)
	} else {
		foodItemUpdateParams.Fats = storedFoodItem.Fats
		foodItemUpdateParams.Carbohydrates = storedFoodItem.Carbohydrates
		foodItemUpdateParams.Protein = storedFoodItem.Protein
		foodItemUpdateParams.Calories = storedFoodItem.Calories
	}

	if itemUpdateInfo.ServingsName != nil && *itemUpdateInfo.ServingsName != "" {
		foodItemUpdateParams.ServingsName = *itemUpdateInfo.ServingsName
	} else {
		foodItemUpdateParams.ServingsName = storedFoodItem.ServingsName
	}

	if itemUpdateInfo.AmountInServings != nil && !itemUpdateInfo.AmountInServings.IsNegative() {
		foodItemUpdateParams.ServingsQuantity = *itemUpdateInfo.AmountInServings
	} else {
		foodItemUpdateParams.ServingsQuantity = storedFoodItem.ServingsQuantity
	}

	updatedFoodItem, err := cfg.db.UpdateFoodItem(r.Context(), foodItemUpdateParams)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.FoodItemUpdateErrStr, err)
		return
	}

	brand := ""
	if updatedFoodItem.Brand != nil {
		brand = *updatedFoodItem.Brand
	}

	respondWithJSON(w, http.StatusOK, dto.FoodItem{
		Id:             updatedFoodItem.ID,
		Date:           updatedFoodItem.Date.Time,
		Category:       dto.Category(updatedFoodItem.Category),
		Name:           updatedFoodItem.Name,
		Fats:           updatedFoodItem.Fats,
		Carbohydrates:  updatedFoodItem.Carbohydrates,
		Protein:        updatedFoodItem.Protein,
		Calories:       updatedFoodItem.Calories,
		Brand:          brand,
		UnitType:       dto.Measurement(updatedFoodItem.UnitType),
		Amount:         updatedFoodItem.Quantity,
		ServingsName:   updatedFoodItem.ServingsName,
		ServingsAmount: updatedFoodItem.ServingsQuantity,
		UserId:         userId,
	})
}

func (cfg apiConfig) handlerDeleteFoodItem(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}
