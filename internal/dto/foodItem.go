package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type FoodItem struct {
	Id             uuid.UUID       `json:"id"`
	Date           time.Time       `json:"date"`
	Category       Category        `json:"category"`
	Name           string          `json:"name"`
	Fats           decimal.Decimal `json:"fats"`
	Carbohydrates  decimal.Decimal `json:"carbohydrates"`
	Protein        decimal.Decimal `json:"protein"`
	Calories       decimal.Decimal `json:"calories"`
	Brand          string          `json:"brand"`
	UnitType       Measurement     `json:"unit_type"`
	Amount         decimal.Decimal `json:"amount"`
	ServingsName   string          `json:"servings_name"`
	ServingsAmount decimal.Decimal `json:"amount_in_servings"`
	UserId         uuid.UUID       `json:"user_id"`
}
