package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Ingredient struct {
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

type Measurement bool

const (
	Weight Measurement = false
	Volume Measurement = true
)
