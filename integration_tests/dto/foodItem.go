package dto

import (
	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/helpers"
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

type FoodItemKeys string

const (
	FoodItemId             FoodItemKeys = "id"
	FoodItemDate           FoodItemKeys = "date"
	FoodItemCategory       FoodItemKeys = "category"
	FoodItemName           FoodItemKeys = "name"
	FoodItemFats           FoodItemKeys = "fats"
	FoodItemCarbs          FoodItemKeys = "carboyhydrates"
	FoodItemProtein        FoodItemKeys = "protein"
	FoodItemCalories       FoodItemKeys = "calories"
	FoodItemBrand          FoodItemKeys = "brand"
	FoodItemUnitType       FoodItemKeys = "unit_type"
	FoodItemAmount         FoodItemKeys = "amount"
	FoodItemServingsName   FoodItemKeys = "servings_name"
	FoodItemServingsAmount FoodItemKeys = "amount_in_servings"
	FoodItemUserId         FoodItemKeys = "user_id"
)

func (fI FoodItem) Normalize() FoodItem {
	fI.Fats = fI.Fats.Normalize()
	fI.Carbohydrates = fI.Carbohydrates.Normalize()
	fI.Protein = fI.Protein.Normalize()
	fI.Calories = fI.Calories.Normalize()
	fI.Amount = fI.Amount.Normalize()
	fI.ServingsAmount = fI.ServingsAmount.Normalize()
	return fI
}

func (fI FoodItem) UpdateClone(updates map[string]any) FoodItem {
	oldAmount := fI.Amount
	newAmount := fI.Amount
	if amountVal, ok := updates[string(FoodItemAmount)]; ok {
		newAmount, _ = helpers.CreateDecimal(amountVal)
		if !newAmount.IsPositive() {
			return fI
		}

		fI.Fats = fI.Fats.Div(oldAmount).Mul(newAmount)
		fI.Carbohydrates = fI.Carbohydrates.Div(oldAmount).Mul(newAmount)
		fI.Protein = fI.Protein.Div(oldAmount).Mul(newAmount)
		fI.Calories = fI.Calories.Div(oldAmount).Mul(newAmount)
		fI.Amount = newAmount
	}

	for key, value := range updates {
		foodItemKey := FoodItemKeys(key)
		switch foodItemKey {
		case FoodItemDate:
			year, month, day := value.(time.Time).Date()
			fI.Date = time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
		case FoodItemCategory:
			fI.Category = value.(Category)
		case FoodItemBrand:
			fI.Brand = value.(string)
		case FoodItemServingsName:
			fI.ServingsName = value.(string)
		case FoodItemServingsAmount:
			decimalVal, _ := helpers.CreateDecimal(value)
			fI.ServingsAmount = decimalVal
		}
	}

	return fI
}
