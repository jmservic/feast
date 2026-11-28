package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MealPlanItem struct {
	Id             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Day            int             `json:"day"`
	Category       int             `json:"category"`
	Fats           decimal.Decimal `json:"fats"`
	Carbohydrates  decimal.Decimal `json:"carbohydrates"`
	Protein        decimal.Decimal `json:"protein"`
	Calories       decimal.Decimal `json:"calories"`
	UnitType       bool            `json:"unit_type"`
	Brand          string          `json:"brand"`
	IngredientId   *uuid.UUID      `json:"ingredient_id"`
	Amount         decimal.Decimal `json:"amount"`
	ServingsName   string          `json:"servings_name"`
	ServingsAmount decimal.Decimal `json:"servings_amount"`
	MealPlanId     uuid.UUID       `json:"meal_plan_id"`
}
