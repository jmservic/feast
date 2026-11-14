package dto

import (
	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/helpers"
	"github.com/shopspring/decimal"
	"slices"
	"strings"
)

type IngredientResponse struct {
	Id            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Fats          decimal.Decimal `json:"fats"`
	Carbohydrates decimal.Decimal `json:"carbohydrates"`
	Protein       decimal.Decimal `json:"protein"`
	Calories      decimal.Decimal `json:"calories"`
	Brand         string          `json:"brand"`
	UserId        *uuid.UUID      `json:"user_id"`
	Amount        decimal.Decimal `json:"amount"`
	UnitType      Measurement     `json:"unit_type"`
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
	IngredientUnitType        IngredientKeys = "unit_type"
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

func (ing IngredientResponse) UpdateClone(updates map[string]any) IngredientResponse {
	amount := decimal.NewFromInt(1)
	if amountVal, ok := updates[string(IngredientAmount)]; ok {
		amount, _ = helpers.CreateDecimal(amountVal)
	}

	for key, value := range updates {
		ingKey := IngredientKeys(key)
		switch ingKey {
		case IngredientName:
			ing.Name = value.(string)
		case IngredientFats:
			decimalVal, _ := helpers.CreateDecimal(value)
			ing.Fats = decimalVal.Div(amount)
		case IngredientCarbs:
			decimalVal, _ := helpers.CreateDecimal(value)
			ing.Carbohydrates = decimalVal.Div(amount)
		case IngredientProtein:
			decimalVal, _ := helpers.CreateDecimal(value)
			ing.Protein = decimalVal.Div(amount)
		case IngredientCalories:
			decimalVal, _ := helpers.CreateDecimal(value)
			ing.Calories = decimalVal.Div(amount)
		case IngredientUnitType:
			switch typeVal := value.(type) {
			case bool:
				ing.UnitType = Measurement(typeVal)
			case Measurement:
				ing.UnitType = typeVal
			}
		case IngredientServings:
			ing.Servings = value.([]ServingInfo)
		case IngredientBrand:
			ing.Brand = value.(string)
		}
	}

	return ing
}
