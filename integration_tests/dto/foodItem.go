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
	Unittype       Measurement     `json:"unittype"`
	Amount         decimal.Decimal `json:"amount"`
	ServingsName   string          `json:"servings_name"`
	ServingsAmount decimal.Decimal `json:"servings_amount"`
}

type FoodItemKeys string
