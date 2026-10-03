package dto

import (
	"github.com/shopspring/decimal"
)

type ServingInfo struct {
	Id       int16           `json:"id"`
	Name     string          `json:"name"`
	Quantity decimal.Decimal `json:"quantity"`
}
