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
				Name:   "Servings",
				Amount: decimal.NewFromInt(128),
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
				dto.FoodItemServingsAmount: ingredient.Servings[0].Amount,
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
				dto.FoodItemServingsAmount: ingredient.Servings[0].Amount,
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
				dto.FoodItemServingsAmount: ingredient.Servings[0].Amount,
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
				dto.FoodItemServingsAmount: ingredient.Servings[0].Amount,
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
				dto.FoodItemServingsAmount: ingredient.Servings[0].Amount,
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
				dto.FoodItemServingsAmount: ingredient.Servings[0].Amount,
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
				dto.FoodItemServingsAmount: ingredient.Servings[0].Amount,
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
				dto.FoodItemServingsAmount: ingredient.Servings[0].Amount,
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
				dto.FoodItemServingsAmount: ingredient.Servings[0].Amount,
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
				Name:   "Servings",
				Amount: decimal.NewFromInt(128),
			},
		}})
	memberIngredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)

	res = client.CreateFoodItem(token, map[dto.FoodItemKeys]any{
		dto.FoodItemId:             memberIngredient.Id,
		dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
		dto.FoodItemCategory:       dto.Lunch,
		dto.FoodItemAmount:         1,
		dto.FoodItemServingsName:   memberIngredient.Servings[0].Name,
		dto.FoodItemServingsAmount: memberIngredient.Servings[0].Amount,
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
				Name:   "Cups",
				Amount: decimal.NewFromInt(240),
			},
			{
				Name:   "1/4 Cups",
				Amount: decimal.NewFromInt(60),
			},
			{
				Name:   "1/2 Cups",
				Amount: decimal.NewFromInt(120),
			},
			{
				Name:   "2 Cups",
				Amount: decimal.NewFromInt(240),
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
		dto.FoodItemServingsAmount: memberIngredient.Servings[0].Amount,
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
				Name:   "Servings",
				Amount: decimal.NewFromInt(1),
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
				Name:   "Cups",
				Amount: decimal.NewFromInt(240),
			},
			{
				Name:   "1/4 Cups",
				Amount: decimal.NewFromInt(60),
			},
			{
				Name:   "1/2 Cups",
				Amount: decimal.NewFromInt(120),
			},
			{
				Name:   "2 Cups",
				Amount: decimal.NewFromInt(240),
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
		dto.FoodItemServingsAmount: memberIngredient.Servings[0].Amount,
	})
	memberFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusCreated)

	res = client.CreateFoodItem(otherMemberLoginResponse.Token, map[dto.FoodItemKeys]any{
		dto.FoodItemId:             otherMemberIngredient.Id,
		dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
		dto.FoodItemCategory:       dto.Lunch,
		dto.FoodItemAmount:         1,
		dto.FoodItemServingsName:   memberIngredient.Servings[0].Name,
		dto.FoodItemServingsAmount: memberIngredient.Servings[0].Amount,
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
			testName:          "Receives not found when trying to update a nonexistent food item",
			id:                uuid.New(),
			token:             token,
			foodItemUserToken: token,
			statusCode:        http.StatusNotFound,
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

func TestFoodItemDelete(t *testing.T) {
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
				Name:   "Servings",
				Amount: decimal.NewFromInt(1),
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
				Name:   "Cups",
				Amount: decimal.NewFromInt(240),
			},
			{
				Name:   "1/4 Cups",
				Amount: decimal.NewFromInt(60),
			},
			{
				Name:   "1/2 Cups",
				Amount: decimal.NewFromInt(120),
			},
			{
				Name:   "2 Cups",
				Amount: decimal.NewFromInt(240),
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
		dto.FoodItemServingsAmount: memberIngredient.Servings[0].Amount,
	})
	memberFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusCreated)

	res = client.CreateFoodItem(otherMemberLoginResponse.Token, map[dto.FoodItemKeys]any{
		dto.FoodItemId:             otherMemberIngredient.Id,
		dto.FoodItemDate:           time.Date(1994, time.November, 21, 0, 0, 0, 0, time.UTC),
		dto.FoodItemCategory:       dto.Lunch,
		dto.FoodItemAmount:         1,
		dto.FoodItemServingsName:   otherMemberIngredient.Servings[0].Name,
		dto.FoodItemServingsAmount: otherMemberIngredient.Servings[0].Amount,
	})
	otherMemberFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusCreated)

	testCases := []struct {
		testName, token, foodItemUserToken string
		id                                 uuid.UUID
		statusCode                         int
	}{
		{
			testName:          "Successfully delete an user's food item",
			id:                memberFoodItem.Id,
			token:             token,
			foodItemUserToken: token,
			statusCode:        http.StatusNoContent,
		},
		{
			testName:          "Receive not found when deleting a nonexistent food item",
			id:                uuid.New(),
			token:             token,
			foodItemUserToken: token,
			statusCode:        http.StatusNotFound,
		},
		{
			testName:          "Receive forbidden when deleting another user's food item",
			id:                otherMemberFoodItem.Id,
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
				prevFoodItem = helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusOK).Normalize()
			} else {
				res.Body.Close()
			}

			res = client.DeleteFoodItem(testCase.token, testCase.id)
			res.Body.Close()

			if res.StatusCode != testCase.statusCode {
				t.Fatalf("expected delete food item to return %d, received %d", testCase.statusCode, res.StatusCode)
			}

			res = client.GetFoodItem(testCase.foodItemUserToken, testCase.id)

			switch testCase.statusCode {
			case http.StatusNoContent:
				res.Body.Close()
				if res.StatusCode != http.StatusNotFound {
					t.Fatalf("expected not to find the ingredient, received %d", res.StatusCode)
				}
			case http.StatusForbidden:
				currFoodItem := helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusOK).Normalize()
				assert.Equal(t, prevFoodItem, currFoodItem)
			default:
				res.Body.Close()
			}
		})
	}
}

func TestGetFoodItems(t *testing.T) {
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

	memberFoodItems := make(dto.FoodItems, 0)
	foodInfos := map[string]map[dto.IngredientKeys]any{
		"Enriched Long Grain White Rice": {
			dto.IngredientName:     "Enriched Long Grain White Rice",
			dto.IngredientCarbs:    decimal.NewFromInt(36),
			dto.IngredientProtein:  decimal.NewFromInt(3),
			dto.IngredientCalories: decimal.NewFromInt(160),
			dto.IngredientBrand:    "Food Lion",
			dto.IngredientAmount:   45,
			dto.IngredientUnitType: dto.Weight,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "grams",
					Amount: decimal.NewFromInt(1),
				},
			},
		},
		"Black Beans": {
			dto.IngredientName:     "Black Beans",
			dto.IngredientFats:     0.5,
			dto.IngredientCarbs:    36,
			dto.IngredientProtein:  8,
			dto.IngredientCalories: 130,
			dto.IngredientBrand:    "Goya",
			dto.IngredientAmount:   122,
			dto.IngredientUnitType: dto.Weight,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "grams",
					Amount: decimal.NewFromInt(1),
				},
			},
		},
		"Green Beans": {
			dto.IngredientName:     "Green Beans",
			dto.IngredientCarbs:    6,
			dto.IngredientProtein:  2,
			dto.IngredientCalories: 25,
			dto.IngredientBrand:    "Pero Family Farms",
			dto.IngredientAmount:   85,
			dto.IngredientUnitType: dto.Weight,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "grams",
					Amount: decimal.NewFromInt(1),
				},
			},
		},
		"Organic Whole Milk": {
			dto.IngredientName:     "Organic Whole Milk",
			dto.IngredientFats:     8,
			dto.IngredientCarbs:    13,
			dto.IngredientProtein:  8,
			dto.IngredientCalories: 160,
			dto.IngredientBrand:    "Horizon Organic",
			dto.IngredientAmount:   240,
			dto.IngredientUnitType: dto.Volume,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "mL",
					Amount: decimal.NewFromInt(1),
				},
			},
		},
		"Simple Italian Chicken": {
			dto.IngredientName:     "Simple Italian Chicken",
			dto.IngredientFats:     14.4,
			dto.IngredientCarbs:    8.9,
			dto.IngredientProtein:  45.1,
			dto.IngredientCalories: 345,
			dto.IngredientAmount:   1,
			dto.IngredientUnitType: dto.Weight,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "dto.IngredientServings",
					Amount: decimal.Zero,
				},
			},
		},
		"Simple Spinach Scramble": {
			dto.IngredientName:     "Simple Spinach Scramble",
			dto.IngredientFats:     10.3,
			dto.IngredientCarbs:    5.2,
			dto.IngredientProtein:  35.9,
			dto.IngredientCalories: 277,
			dto.IngredientAmount:   1,
			dto.IngredientUnitType: dto.Weight,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "dto.IngredientServings",
					Amount: decimal.Zero,
				},
			},
		},
		"Whole Milk Greek Yogurt": {
			dto.IngredientName:     "Whole Milk Greek Yogurt",
			dto.IngredientFats:     8,
			dto.IngredientCarbs:    8,
			dto.IngredientProtein:  16,
			dto.IngredientCalories: 160,
			dto.IngredientBrand:    "Food Lion",
			dto.IngredientAmount:   170,
			dto.IngredientUnitType: dto.Weight,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "grams",
					Amount: decimal.NewFromInt(1),
				},
			},
		},
		"Egg Whites": {
			dto.IngredientName:     "Egg Whites",
			dto.IngredientProtein:  5,
			dto.IngredientCalories: 25,
			dto.IngredientBrand:    "Food lion",
			dto.IngredientAmount:   46,
			dto.IngredientUnitType: dto.Weight,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "grams",
					Amount: decimal.NewFromInt(1),
				},
			},
		},
		"Sour Cream Grade A": {
			dto.IngredientName:     "Sour Cream Grade A",
			dto.IngredientFats:     5,
			dto.IngredientCarbs:    1,
			dto.IngredientProtein:  0.5,
			dto.IngredientCalories: 60,
			dto.IngredientBrand:    "food Lion",
			dto.IngredientAmount:   30,
			dto.IngredientUnitType: dto.Weight,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "grams",
					Amount: decimal.NewFromInt(1),
				},
			},
		},
		"Chicken Breast": {
			dto.IngredientName:     "Chicken Breast",
			dto.IngredientFats:     2,
			dto.IngredientCarbs:    0,
			dto.IngredientProtein:  23,
			dto.IngredientCalories: 110,
			dto.IngredientAmount:   113.398093,
			dto.IngredientUnitType: dto.Weight,
			dto.IngredientServings: []dto.ServingInfo{
				{
					Name:   "grams",
					Amount: decimal.NewFromInt(1),
				},
			},
		},
	}

	memberIngredients := make(map[string]dto.IngredientResponse, len(foodInfos))

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

	for name, foodInfo := range foodInfos {
		res := client.CreateIngredient(token, foodInfo)
		memberIngredients[name] = helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)
	}

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
				Name:   "Cups",
				Amount: decimal.NewFromInt(240),
			},
			{
				Name:   "1/4 Cups",
				Amount: decimal.NewFromInt(60),
			},
			{
				Name:   "1/2 Cups",
				Amount: decimal.NewFromInt(120),
			},
			{
				Name:   "2 Cups",
				Amount: decimal.NewFromInt(240),
			},
		},
	})
	otherMemberIngredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)
	res = client.CreateFoodItem(otherMemberLoginResponse.Token, map[dto.FoodItemKeys]any{
		dto.FoodItemId:             otherMemberIngredient.Id,
		dto.FoodItemDate:           time.Date(2026, time.November, 16, 0, 0, 0, 0, time.UTC),
		dto.FoodItemCategory:       dto.Lunch,
		dto.FoodItemAmount:         1,
		dto.FoodItemServingsName:   otherMemberIngredient.Servings[0].Name,
		dto.FoodItemServingsAmount: otherMemberIngredient.Servings[0].Amount,
	})
	res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected the create food item to return status created, received %d", res.StatusCode)
	}

	foodItemParams := []map[dto.FoodItemKeys]any{
		{
			dto.FoodItemId:             memberIngredients["Organic Whole Milk"].Id,
			dto.FoodItemDate:           time.Date(2026, time.November, 16, 0, 0, 0, 0, time.UTC),
			dto.FoodItemCategory:       dto.Breakfast,
			dto.FoodItemAmount:         240,
			dto.FoodItemServingsName:   memberIngredients["Organic Whole Milk"].Servings[0].Name,
			dto.FoodItemServingsAmount: memberIngredients["Organic Whole Milk"].Servings[0].Amount,
		},
		{
			dto.FoodItemId:             memberIngredients["Simple Italian Chicken"].Id,
			dto.FoodItemDate:           time.Date(2026, time.November, 16, 0, 0, 0, 0, time.UTC),
			dto.FoodItemCategory:       dto.Lunch,
			dto.FoodItemAmount:         1,
			dto.FoodItemServingsName:   memberIngredients["Simple Italian Chicken"].Servings[0].Name,
			dto.FoodItemServingsAmount: memberIngredients["Simple Italian Chicken"].Servings[0].Amount,
		},
		{
			dto.FoodItemId:             memberIngredients["Whole Milk Greek Yogurt"].Id,
			dto.FoodItemDate:           time.Date(2026, time.November, 16, 0, 0, 0, 0, time.UTC),
			dto.FoodItemCategory:       dto.Dinner,
			dto.FoodItemAmount:         170,
			dto.FoodItemServingsName:   memberIngredients["Whole Milk Greek Yogurt"].Servings[0].Name,
			dto.FoodItemServingsAmount: memberIngredients["Whole Milk Greek Yogurt"].Servings[0].Amount,
		},
		{
			dto.FoodItemId:             memberIngredients["Simple Spinach Scramble"].Id,
			dto.FoodItemDate:           time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC),
			dto.FoodItemCategory:       dto.Breakfast,
			dto.FoodItemAmount:         1,
			dto.FoodItemServingsName:   memberIngredients["Simple Spinach Scramble"].Servings[0].Name,
			dto.FoodItemServingsAmount: memberIngredients["Simple Spinach Scramble"].Servings[0].Amount,
		},
		{
			dto.FoodItemId:             memberIngredients["Simple Italian Chicken"].Id,
			dto.FoodItemDate:           time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC),
			dto.FoodItemCategory:       dto.Lunch,
			dto.FoodItemAmount:         1,
			dto.FoodItemServingsName:   memberIngredients["Simple Italian Chicken"].Servings[0].Name,
			dto.FoodItemServingsAmount: memberIngredients["Simple Italian Chicken"].Servings[0].Amount,
		},
		{
			dto.FoodItemId:             memberIngredients["Whole Milk Greek Yogurt"].Id,
			dto.FoodItemDate:           time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC),
			dto.FoodItemCategory:       dto.Snacks,
			dto.FoodItemAmount:         170,
			dto.FoodItemServingsName:   memberIngredients["Whole Milk Greek Yogurt"].Servings[0].Name,
			dto.FoodItemServingsAmount: memberIngredients["Whole Milk Greek Yogurt"].Servings[0].Amount,
		},
		{
			dto.FoodItemId:             memberIngredients["Organic Whole Milk"].Id,
			dto.FoodItemDate:           time.Date(2026, time.November, 18, 0, 0, 0, 0, time.UTC),
			dto.FoodItemCategory:       dto.Breakfast,
			dto.FoodItemAmount:         240,
			dto.FoodItemServingsName:   memberIngredients["Organic Whole Milk"].Servings[0].Name,
			dto.FoodItemServingsAmount: memberIngredients["Organic Whole Milk"].Servings[0].Amount,
		},
		{
			dto.FoodItemId:             memberIngredients["Enriched Long Grain White Rice"].Id,
			dto.FoodItemDate:           time.Date(2026, time.November, 18, 0, 0, 0, 0, time.UTC),
			dto.FoodItemCategory:       dto.Lunch,
			dto.FoodItemAmount:         45,
			dto.FoodItemServingsName:   memberIngredients["Enriched Long Grain White Rice"].Servings[0].Name,
			dto.FoodItemServingsAmount: memberIngredients["Enriched Long Grain White Rice"].Servings[0].Amount,
		},
		{
			dto.FoodItemId:             memberIngredients["Chicken Breast"].Id,
			dto.FoodItemDate:           time.Date(2026, time.November, 18, 0, 0, 0, 0, time.UTC),
			dto.FoodItemCategory:       dto.Dinner,
			dto.FoodItemAmount:         300,
			dto.FoodItemServingsName:   memberIngredients["Chicken Breast"].Servings[0].Name,
			dto.FoodItemServingsAmount: memberIngredients["Chicken Breast"].Servings[0].Amount,
		},
	}

	for _, params := range foodItemParams {
		res := client.CreateFoodItem(token, params)
		memberFoodItems = append(memberFoodItems, helpers.GetResponseObject[dto.FoodItem](t, res, http.StatusCreated))
	}

	testCases := []struct {
		testName           string
		startDate, endDate time.Time
		statusCode         int
	}{
		{
			testName:   "Get food log for one day",
			startDate:  time.Date(2026, time.November, 16, 1, 1, 1, 1, time.UTC),
			endDate:    time.Date(2026, time.November, 16, 0, 0, 0, 0, time.UTC),
			statusCode: http.StatusOK,
		},
		{
			testName:   "Get food log for one day while omitting the endDate",
			startDate:  time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC),
			statusCode: http.StatusOK,
		},
		{
			testName:   "Get food long for multiple days",
			startDate:  time.Date(2026, time.November, 16, 0, 0, 0, 0, time.UTC),
			endDate:    time.Date(2026, time.November, 18, 0, 0, 0, 0, time.UTC),
			statusCode: http.StatusOK,
		},
		{
			testName:   "receive bad request when endDate is before the startDate",
			startDate:  time.Date(2026, time.November, 18, 0, 0, 0, 0, time.UTC),
			endDate:    time.Date(2026, time.November, 16, 0, 0, 0, 0, time.UTC),
			statusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			startDate := &testCase.startDate
			var endDate *time.Time
			zeroTime := time.Time{}
			if testCase.endDate != zeroTime {
				endDate = &testCase.endDate
			} else {
				endDate = startDate
			}

			res := client.GetFoodItems(token, startDate, endDate)
			if res.StatusCode != testCase.statusCode {
				t.Fatalf("expected get food items to return %d, received %d", testCase.statusCode, res.StatusCode)
			}
			if res.StatusCode == http.StatusOK {
				actualFoodItems := helpers.GetResponseObject[[]dto.FoodItem](t, res, http.StatusOK)
				for index, foodItem := range actualFoodItems {
					actualFoodItems[index] = foodItem.Normalize()
				}
				expectedFoodItems := memberFoodItems.GetItemsByRange(*startDate, *endDate)
				assert.EqualValues(t, expectedFoodItems, actualFoodItems)
			} else {
				res.Body.Close()
			}
		})
	}

}
