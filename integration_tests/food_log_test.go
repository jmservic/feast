package integration

import (
	"errors"
	"github.com/google/uuid"
	"net/http"
	"testing"

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
				Quantity: decimal.Zero,
			},
		}})
	ingredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)

	testCases := []struct {
		testName   string
		params     map[dto.FoodItemKeys]any
		statusCode int
	}{}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {

		})
	}
}
