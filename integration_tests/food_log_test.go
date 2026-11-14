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

func TestFoodItemGet(t *testing.T) {
	helpers.LoadDotEnv()

	member := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}

	otherMember := UserInfo{
		name:     "cassidy",
		email:    "Casstadon@example.com",
		password: "very-secreto",
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

	res = client.CreateUser(otherMember.name, otherMember.email, otherMember.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(otherMember.email, otherMember.password)
	otherMemberLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

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
	memberIngredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)

	res = client.CreateFoodItem(token, map[dto.FoodItemKeys]any{
		dto.FoodItemId:             memberIngredient.Id,
		dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
		dto.FoodItemCategory:       dto.Lunch,
		dto.FoodItemAmount:         1,
		dto.FoodItemServingsName:   memberIngredient.Servings[0].Name,
		dto.FoodItemServingsAmount: memberIngredient.Servings[0].Quantity,
	})
	memberFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusCreated)

	res = client.CreateIngredient(otherMemberLoginResponse.Token, map[dto.IngredientKeys]any{
		dto.IngredientName:     "Organic Whole Milk",
		dto.IngredientProtein:  8,
		dto.IngredientFats:     0.5,
		dto.IngredientCarbs:    13,
		dto.IngredientCalories: 160,
		dto.IngredientBrand:    "Horizon Organic",
		dto.IngredientAmount:   240,
		dto.IngredientUnitType: dto.Volume,
		dto.IngredientServings: []dto.ServingInfo{
			{
				Name:     "Cups",
				Quantity: decimal.NewFromInt(240),
			},
			{
				Name:     "1/4 Cups",
				Quantity: decimal.NewFromInt(60),
			},
			{
				Name:     "1/2 Cups",
				Quantity: decimal.NewFromInt(120),
			},
			{
				Name:     "2 Cups",
				Quantity: decimal.NewFromInt(240),
			},
		},
	})
	otherMemberIngredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)

	res = client.CreateFoodItem(otherMemberLoginResponse.Token, map[dto.FoodItemKeys]any{
		dto.FoodItemId:             otherMemberIngredient.Id,
		dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
		dto.FoodItemCategory:       dto.Lunch,
		dto.FoodItemAmount:         1,
		dto.FoodItemServingsName:   memberIngredient.Servings[0].Name,
		dto.FoodItemServingsAmount: memberIngredient.Servings[0].Quantity,
	})
	otherMemberFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusCreated)

	testCases := []struct {
		testName, token string
		foodItemId      uuid.UUID
		expectedItem    dto.FoodItem
		statusCode      int
	}{
		{
			testName:     "Successfully get user's food item",
			foodItemId:   memberFoodItem.Id,
			expectedItem: memberFoodItem.Normalize(),
			token:        token,
			statusCode:   http.StatusOK,
		},
		{
			testName:   "Receive not found for a nonexistent food item",
			foodItemId: uuid.New(),
			token:      token,
			statusCode: http.StatusNotFound,
		},
		{
			testName:   "Receive forbidden for another user's food item",
			foodItemId: otherMemberFoodItem.Id,
			statusCode: http.StatusForbidden,
			token:      token,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.GetFoodItem(testCase.token, testCase.foodItemId)

			if res.StatusCode != testCase.statusCode {
				t.Fatalf("Expected get food item to return %d, received %d", testCase.statusCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				foodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusOK).Normalize()
				assert.Equal(t, testCase.expectedItem, foodItem)
			} else {
				res.Body.Close()
			}

		})

	}
}

func TestFoodItemUpdate(t *testing.T) {
	helpers.LoadDotEnv()

	member := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}

	otherMember := UserInfo{
		name:     "cassidy",
		email:    "Casstadon@example.com",
		password: "very-secreto",
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

	res = client.CreateUser(otherMember.name, otherMember.email, otherMember.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(otherMember.email, otherMember.password)
	otherMemberLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

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
				Quantity: decimal.NewFromInt(1),
			},
		}})
	memberIngredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)

	res = client.CreateIngredient(otherMemberLoginResponse.Token, map[dto.IngredientKeys]any{
		dto.IngredientName:     "Organic Whole Milk",
		dto.IngredientProtein:  8,
		dto.IngredientFats:     0.5,
		dto.IngredientCarbs:    13,
		dto.IngredientCalories: 160,
		dto.IngredientBrand:    "Horizon Organic",
		dto.IngredientAmount:   240,
		dto.IngredientUnitType: dto.Volume,
		dto.IngredientServings: []dto.ServingInfo{
			{
				Name:     "Cups",
				Quantity: decimal.NewFromInt(240),
			},
			{
				Name:     "1/4 Cups",
				Quantity: decimal.NewFromInt(60),
			},
			{
				Name:     "1/2 Cups",
				Quantity: decimal.NewFromInt(120),
			},
			{
				Name:     "2 Cups",
				Quantity: decimal.NewFromInt(240),
			},
		},
	})
	otherMemberIngredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)

	res = client.CreateFoodItem(token, map[dto.FoodItemKeys]any{
		dto.FoodItemId:             memberIngredient.Id,
		dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
		dto.FoodItemCategory:       dto.Lunch,
		dto.FoodItemAmount:         1,
		dto.FoodItemServingsName:   memberIngredient.Servings[0].Name,
		dto.FoodItemServingsAmount: memberIngredient.Servings[0].Quantity,
	})
	memberFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusCreated)

	res = client.CreateFoodItem(otherMemberLoginResponse.Token, map[dto.FoodItemKeys]any{
		dto.FoodItemId:             otherMemberIngredient.Id,
		dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
		dto.FoodItemCategory:       dto.Lunch,
		dto.FoodItemAmount:         1,
		dto.FoodItemServingsName:   memberIngredient.Servings[0].Name,
		dto.FoodItemServingsAmount: memberIngredient.Servings[0].Quantity,
	})
	otherMemberFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusCreated)

	testCases := []struct {
		testName, token, foodItemUserToken string
		id                                 uuid.UUID
		params                             map[string]any
		statusCode                         int
	}{
		{
			testName: "Successfully update all fields",
			id:       memberFoodItem.Id,
			params: map[string]any{
				string(dto.FoodItemDate):           time.Now(),
				string(dto.FoodItemCategory):       dto.Dinner,
				string(dto.FoodItemAmount):         2,
				string(dto.FoodItemServingsName):   "double helping",
				string(dto.FoodItemServingsAmount): 2,
			},
			token:             token,
			foodItemUserToken: token,
			statusCode:        http.StatusOK,
		},
		{
			testName: "Successfully update date only",
			id:       memberFoodItem.Id,
			params: map[string]any{
				string(dto.FoodItemDate): time.Date(2016, time.May, 5, 0, 0, 0, 0, time.UTC),
			},
			token:             token,
			foodItemUserToken: token,
			statusCode:        http.StatusOK,
		},
		{
			testName: "Successfully update category only",
			id:       memberFoodItem.Id,
			params: map[string]any{
				string(dto.FoodItemCategory): dto.Snacks,
			},
			token:             token,
			foodItemUserToken: token,
			statusCode:        http.StatusOK,
		},
		{
			testName: "Successfully update Amount only",
			id:       memberFoodItem.Id,
			params: map[string]any{
				string(dto.FoodItemAmount): 8,
			},
			token:             token,
			foodItemUserToken: token,
			statusCode:        http.StatusOK,
		},
		{
			testName: "Successfully update servings name only",
			id:       memberFoodItem.Id,
			params: map[string]any{
				string(dto.FoodItemServingsName): "HachiBan Servings",
			},
			token:             token,
			foodItemUserToken: token,
			statusCode:        http.StatusOK,
		},
		{
			testName: "Successfully update servings amount only",
			id:       memberFoodItem.Id,
			params: map[string]any{
				string(dto.FoodItemServingsAmount): 8,
			},
			token:             token,
			foodItemUserToken: token,
			statusCode:        http.StatusOK,
		},

		{
			testName:   "Receives not found when trying to update a nonexistent food item",
			id:         uuid.New(),
			token:      token,
			statusCode: http.StatusNotFound,
		},
		{
			testName: "Receives forbidden when trying to update another user's food item",
			id:       otherMemberFoodItem.Id,
			params: map[string]any{
				string(dto.FoodItemDate):           time.Now(),
				string(dto.FoodItemCategory):       dto.Dinner,
				string(dto.FoodItemAmount):         2,
				string(dto.FoodItemServingsName):   "double helping",
				string(dto.FoodItemServingsAmount): 2,
			},
			token:             token,
			foodItemUserToken: otherMemberLoginResponse.Token,
			statusCode:        http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			var prevFoodItem dto.FoodItem

			res := client.GetFoodItem(testCase.foodItemUserToken, testCase.id)

			if res.StatusCode == http.StatusOK {
				prevFoodItem = helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusOK)
			} else {
				res.Body.Close()
			}

			res = client.UpdateFoodItem(testCase.token, testCase.id, testCase.params)
			if res.StatusCode != testCase.statusCode {
				t.Fatalf("expected update food item to return %d, received %d", testCase.statusCode, res.StatusCode)

			}

			if res.StatusCode == http.StatusOK {
				currFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusOK).Normalize()
				expectedFoodItem := prevFoodItem.UpdateClone(testCase.params).Normalize()
				assert.Equal(t, expectedFoodItem, currFoodItem)

				t.Log(currFoodItem.Normalize().Amount)
			} else {
				res.Body.Close()

				if res.StatusCode != http.StatusNotFound {
					res := client.GetFoodItem(testCase.foodItemUserToken, testCase.id)
					currFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusOK)
					assert.Equal(t, prevFoodItem.Normalize(), currFoodItem.Normalize())
				}
			}
		})
	}
}
