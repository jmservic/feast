package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"slices"
	"strings"
)

type IngredientResponse struct {
	Id            int64           `json:"id"`
	Name          string          `json:"name"`
	Fats          decimal.Decimal `json:"fats"`
	Carbohydrates decimal.Decimal `json:"carbohydrates"`
	Protein       decimal.Decimal `json:"protein"`
	Calories      decimal.Decimal `json:"calories"`
	Brand         string          `json:"brand"`
	UserId        *uuid.UUID      `json:"user_id"`
	Amount        decimal.Decimal `json:"amount"`
	Unittype      Measurement     `json:"unittype"`
	Servings      []ServingInfo   `json:"servings"`
}

type IngredientKeys string

const (
	IngredientId              IngredientKeys = "id"
	IngredientName            IngredientKeys = "name"
	IngredientFats            IngredientKeys = "fats"
	IngredientCarbs           IngredientKeys = "carbohydrates"
	IngredientProtein         IngredientKeys = "protein"
	IngredientCalories        IngredientKeys = "calories"
	IngredientBrand           IngredientKeys = "brand"
	IngredientUserId          IngredientKeys = "user_id"
	IngredientAmount          IngredientKeys = "amount"
	IngredientUnitType        IngredientKeys = "unittype"
	IngredientServings        IngredientKeys = "servings"
	IngredientRemovedServings IngredientKeys = "removed_servings"
)

type Measurement bool

const (
	Weight Measurement = false
	Volume Measurement = true
)

func (ing IngredientResponse) Normalize() IngredientResponse {
	ing.Fats = ing.Fats.Normalize()
	ing.Carbohydrates = ing.Carbohydrates.Normalize()
	ing.Protein = ing.Protein.Normalize()
	ing.Calories = ing.Calories.Normalize()
	ing.Amount = ing.Amount.Normalize()
	for i, servings := range ing.Servings {
		servings.Id = -1
		ing.Servings[i] = servings
	}
	slices.SortFunc(ing.Servings, func(a, b ServingInfo) int {
		return strings.Compare(a.Name, b.Name)
	})
	return ing
}
