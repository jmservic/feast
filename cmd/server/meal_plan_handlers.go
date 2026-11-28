package main

import (
	"encoding/json"
	//	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	//	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"

	//	"github.com/jmservic/feast/internal/dto"
	"net/http"

	"github.com/shopspring/decimal"
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
			ServingsAmount decimal.Decimal `json:"servings_amount"`
		} `json:"items"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, nil)
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
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.MealPlanCreationErrStr, err)
		return
	}

	ingIdMap := map[uuid.UUID]dto.Ingredient{}

	for _, item := range params.Items {
		ingIdMap[item.Id] = dto.Ingredient{}
	}

	ingIds := make([]uuid.UUID, 0, len(ingIdMap))

	for id := range ingIdMap {
		ingIds = append(ingIds, id)
	}

	storedIngredients, err := queries.GetIngredientsByIds(r.Context(), ingIds)

	for _, storedIngredient := range storedIngredients {
		brand := ""
		if storedIngredient.Brand != nil {
			brand = *storedIngredient.Brand
		}

		ingIdMap[storedIngredient.ID] = dto.Ingredient{
			Id:            storedIngredient.ID,
			Name:          storedIngredient.Name,
			Fats:          storedIngredient.Fats,
			Carbohydrates: storedIngredient.Carbohydrates,
			Protein:       storedIngredient.Protein,
			Calories:      storedIngredient.Calories,
			Brand:         brand,
			Amount:        decimal.NewFromInt(1),
			UnitType:      dto.Measurement(storedIngredient.UnitType),
			UserId:        storedIngredient.UserID,
		}
	}

	mealPlanItems := make([]dto.MealPlanItem, 0, len(params.Items))
	for _, item := range params.Items {
		ingredient := ingIdMap[item.Id]
		mealPlanItems = append(mealPlanItems, dto.MealPlanItem{
			Name:           ingredient.Name,
			Day:            item.Day,
			Category:       item.Category,
			Fats:           ingredient.Fats,
			Carbohydrates:  ingredient.Carbohydrates,
			Protein:        ingredient.Protein,
			UnitType:       bool(ingredient.UnitType),
			Brand:          ingredient.Brand,
			IngredientId:   &item.Id,
			Amount:         item.Amount,
			ServingsName:   item.ServingsName,
			ServingsAmount: item.ServingsAmount,
		})
	}

	err = queries.CreateUpdateMealPlanItems(r.Context(), database.CreateUpdateMealPlanItemsParams{
		MealPlanID:    storedMealPlan.ID,
		MealPlanItems: mealPlanItems,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.MealPlanItemCreationErrStr, err)
		return
	}

}

func (cfg apiConfig) handlerUpdateMealPlan(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}

func (cfg apiConfig) handlerDeleteMealPlan(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {

}
