package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jmservic/feast/integration_tests/dto"
	"github.com/jmservic/feast/integration_tests/helpers"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestFoodItemCreate(t *testing.T) {
	helpers.LoadDotEnv()

	member := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	res := client.CreateUser(member.name, member.email, member.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(member.email, member.password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)
	token := loginResponse.Token

	res = client.CreateIngredient(token, map[dto.IngredientKeys]any{
		dto.IngredientName:     "Simple Italian Chicken",
		dto.IngredientProtein:  45.1,
		dto.IngredientFats:     13,
		dto.IngredientCarbs:    8.9,
		dto.IngredientCalories: 345,
		dto.IngredientBrand:    "Service Fam",
		dto.IngredientAmount:   1,
		dto.IngredientUnitType: dto.Weight,
		dto.IngredientServings: []dto.ServingInfo{
			{
				Name:     "Servings",
				Quantity: decimal.NewFromInt(128),
			},
		}})
	ingredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)

	testCases := []struct {
		testName   string
		params     map[dto.FoodItemKeys]any
		statusCode int
	}{
		{
			testName: "Successfully create Food Item",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemId:             ingredient.Id,
				dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemAmount:         2,
				dto.FoodItemServingsName:   ingredient.Servings[0].Name,
				dto.FoodItemServingsAmount: ingredient.Servings[0].Quantity,
			},
			statusCode: http.StatusCreated,
		},
		{
			testName: "Failed to create Food Item with zero amount",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemId:             ingredient.Id,
				dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemAmount:         0,
				dto.FoodItemServingsName:   ingredient.Servings[0].Name,
				dto.FoodItemServingsAmount: ingredient.Servings[0].Quantity,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Failed to create Food Item with missing amount",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemId:             ingredient.Id,
				dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemServingsName:   ingredient.Servings[0].Name,
				dto.FoodItemServingsAmount: ingredient.Servings[0].Quantity,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Failed to create Food Item with negative amount",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemId:             ingredient.Id,
				dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemAmount:         -1,
				dto.FoodItemServingsName:   ingredient.Servings[0].Name,
				dto.FoodItemServingsAmount: ingredient.Servings[0].Quantity,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Failed to create Food Item with negative servings amount",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemId:             ingredient.Id,
				dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemAmount:         1,
				dto.FoodItemServingsName:   ingredient.Servings[0].Name,
				dto.FoodItemServingsAmount: -1,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Failed to create Food Item with empty servings name",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemId:             ingredient.Id,
				dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemAmount:         1,
				dto.FoodItemServingsName:   "",
				dto.FoodItemServingsAmount: ingredient.Servings[0].Quantity,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Failed to create Food Item with missing servings name",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemId:             ingredient.Id,
				dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemAmount:         1,
				dto.FoodItemServingsAmount: ingredient.Servings[0].Quantity,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Failed to create Food Item with bad date",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemId:             ingredient.Id,
				dto.FoodItemDate:           "can't parse this, can you?",
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemAmount:         1,
				dto.FoodItemServingsName:   ingredient.Servings[0].Name,
				dto.FoodItemServingsAmount: ingredient.Servings[0].Quantity,
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			testName: "Failed to create Food Item with unknown ingredient id",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemId:             uuid.New(),
				dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemAmount:         1,
				dto.FoodItemServingsName:   ingredient.Servings[0].Name,
				dto.FoodItemServingsAmount: ingredient.Servings[0].Quantity,
			},
			statusCode: http.StatusNotFound,
		},
		{
			testName: "Failed to create Food Item with missin ingredient id",
			params: map[dto.FoodItemKeys]any{
				dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
				dto.FoodItemCategory:       dto.Lunch,
				dto.FoodItemAmount:         1,
				dto.FoodItemServingsName:   ingredient.Servings[0].Name,
				dto.FoodItemServingsAmount: ingredient.Servings[0].Quantity,
			},
			statusCode: http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.CreateFoodItem(token, testCase.params)
			if res.StatusCode != testCase.statusCode {
				t.Fatalf("Expected create food item to return %d, received %d", testCase.statusCode, res.StatusCode)
			}

			if testCase.statusCode == http.StatusCreated {
				sut := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusCreated)
				sut.Id = uuid.Nil
				amount, err := helpers.CreateDecimal(testCase.params[dto.FoodItemAmount])
				if err != nil {
					t.Fatalf("Error getting the food item amount: %v", err)
				}
				servingsAmount, err := helpers.CreateDecimal(testCase.params[dto.FoodItemServingsAmount])
				if err != nil {
					t.Fatalf("Error getting the food item servings amount: %v", err)
				}
				expected := dto.FoodItem{
					Id:             uuid.Nil,
					Date:           testCase.params[dto.FoodItemDate].(time.Time),
					Category:       testCase.params[dto.FoodItemCategory].(dto.Category),
					Name:           ingredient.Name,
					Fats:           ingredient.Fats.Mul(amount),
					Carbohydrates:  ingredient.Carbohydrates.Mul(amount),
					Protein:        ingredient.Protein.Mul(amount),
					Calories:       ingredient.Calories.Mul(amount),
					Brand:          ingredient.Brand,
					UnitType:       ingredient.UnitType,
					Amount:         amount,
					ServingsName:   testCase.params[dto.FoodItemServingsName].(string),
					ServingsAmount: servingsAmount,
					UserId:         loginResponse.Id,
				}
				assert.Equal(t, expected, sut)
			} else {
				res.Body.Close()
			}
		})
	}
}
