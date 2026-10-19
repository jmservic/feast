package integration

import (
	//	"github.com/google/uuid"
	"errors"
	"github.com/jmservic/feast/integration_tests/dto"
	"github.com/jmservic/feast/integration_tests/helpers"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func CreateDecimal(value any) (decimal.Decimal, error) {
	switch v := value.(type) {
	case int:
		return decimal.NewFromInt(int64(v)), nil
	case int32:
		return decimal.NewFromInt32(v), nil
	case int64:
		return decimal.NewFromInt(v), nil
	case string:
		return decimal.NewFromString(v)
	case float32:
		return decimal.NewFromFloat32(v), nil
	case float64:
		return decimal.NewFromFloat(v), nil
	default:
		return decimal.Zero, errors.New("Unknown type!!")
	}
}

func TestCreateIngredient(t *testing.T) {
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

	//Failing cases: No Name, 0 amount, negative amount, negative protein, fats, carbs, calories, no servings
	testCases := []struct {
		testName   string
		params     map[dto.IngredientKeys]any
		statusCode int
	}{
		{
			testName: "Successfully created ingredient with all fields with a single serving, not based on weight / volume",
			params: map[dto.IngredientKeys]any{
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
				},
			},
			statusCode: http.StatusCreated,
		},
		{
			testName: "Successfully creates ingredient with all fields and amount greater than 1",
			params: map[dto.IngredientKeys]any{
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
			},
			statusCode: http.StatusCreated,
		},
		{
			testName: "Successfully creates ingredient with missing protein, fats, carbs, and calories",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Water",
				dto.IngredientAmount:   240,
				dto.IngredientUnitType: dto.Volume,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "Cups",
						Quantity: decimal.NewFromInt(240),
					},
				},
			},
			statusCode: http.StatusCreated,
		},
		{
			testName: "Fails to create ingredient with no name",
			params: map[dto.IngredientKeys]any{
				dto.IngredientProtein:  3,
				dto.IngredientCarbs:    36,
				dto.IngredientCalories: 160,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   45,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.NewFromInt(45),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with empty name",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "",
				dto.IngredientProtein:  3,
				dto.IngredientCarbs:    36,
				dto.IngredientCalories: 160,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   45,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.NewFromInt(45),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with negative protein",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  -8,
				dto.IngredientFats:     0.5,
				dto.IngredientCarbs:    23,
				dto.IngredientCalories: 130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   122,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.NewFromInt(122),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with negative fats",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  8,
				dto.IngredientFats:     -0.5,
				dto.IngredientCarbs:    23,
				dto.IngredientCalories: 130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   122,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.NewFromInt(122),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with negative carbs",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  8,
				dto.IngredientFats:     0.5,
				dto.IngredientCarbs:    -23,
				dto.IngredientCalories: 130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   122,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.NewFromInt(122),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with negative calories",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  8,
				dto.IngredientFats:     0.5,
				dto.IngredientCarbs:    23,
				dto.IngredientCalories: -130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   122,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.NewFromInt(122),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with missing amount",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  8,
				dto.IngredientFats:     0.5,
				dto.IngredientCarbs:    23,
				dto.IngredientCalories: 130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.NewFromInt(122),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with zero amount",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  8,
				dto.IngredientFats:     0.5,
				dto.IngredientCarbs:    23,
				dto.IngredientCalories: 130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   0,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.NewFromInt(122),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with negative amount",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  8,
				dto.IngredientFats:     0.5,
				dto.IngredientCarbs:    23,
				dto.IngredientCalories: 130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   -122,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.NewFromInt(122),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with missing servings",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  8,
				dto.IngredientFats:     0.5,
				dto.IngredientCarbs:    23,
				dto.IngredientCalories: 130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   122,
				dto.IngredientUnitType: dto.Weight,
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with empty servings",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  8,
				dto.IngredientFats:     0.5,
				dto.IngredientCarbs:    23,
				dto.IngredientCalories: 130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   122,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			testName: "Fails to create ingredient with multiple servings when 1 has zero quantity",
			params: map[dto.IngredientKeys]any{
				dto.IngredientName:     "Black Beans",
				dto.IngredientProtein:  8,
				dto.IngredientFats:     0.5,
				dto.IngredientCarbs:    23,
				dto.IngredientCalories: 130,
				dto.IngredientBrand:    "Food Lion",
				dto.IngredientAmount:   122,
				dto.IngredientUnitType: dto.Weight,
				dto.IngredientServings: []dto.ServingInfo{
					{
						Name:     "grams",
						Quantity: decimal.Zero,
					},
					{
						Name:     "ounces",
						Quantity: decimal.NewFromInt(28),
					},
				},
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.CreateIngredient(token, testCase.params)
			if res.StatusCode != testCase.statusCode {
				t.Fatalf("Expected create ingredient to return %d, received %d", testCase.statusCode, res.StatusCode)
			}

			if testCase.statusCode == http.StatusCreated {
				ingredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)
				//t.Log(ingredient)

				// compare values to the map
				amount, err := CreateDecimal(testCase.params[dto.IngredientAmount])
				if err != nil {
					t.Fatal(err)
				}

				if ingredient.Name != testCase.params[dto.IngredientName] {
					t.Fatalf("Expected ingredient name to be '%s', received %s'",
						testCase.params[dto.IngredientName],
						ingredient.Name)
				}

				if proteinVal, ok := testCase.params[dto.IngredientProtein]; ok {
					protein, err := CreateDecimal(proteinVal)
					if err != nil {
						t.Fatal(err)
					}
					rtnVal := ingredient.Protein.Mul(amount).Round(1)
					if !protein.Round(1).Equal(rtnVal) {
						t.Fatalf("Expected ingredient protein to be %v, received %v",
							protein, rtnVal)
					}
				} else if !ingredient.Protein.IsZero() {
					t.Fatalf("Expected the ingredient's protein to be 0, received %v",
						ingredient.Protein)
				}

				if fatsVal, ok := testCase.params[dto.IngredientFats]; ok {
					fats, err := CreateDecimal(fatsVal)
					if err != nil {
						t.Fatal(err)
					}
					rtnVal := ingredient.Fats.Mul(amount).Round(1)
					if !fats.Round(1).Equal(rtnVal) {
						t.Fatalf("Expected ingredient fats to be %v, received %v",
							fats, rtnVal)
					}
				} else if !ingredient.Protein.IsZero() {
					t.Fatalf("Expected the ingredient's fats to be 0, received %v",
						ingredient.Protein)
				}

				if carbVal, ok := testCase.params[dto.IngredientCarbs]; ok {
					carbs, err := CreateDecimal(carbVal)
					if err != nil {
						t.Fatal(err)
					}
					rtnVal := ingredient.Carbohydrates.Mul(amount).Round(1)
					if !carbs.Round(1).Equal(rtnVal) {
						t.Fatalf("Expected ingredient carbohydrates to be %v, received %v",
							carbs, rtnVal)
					}
				} else if !ingredient.Carbohydrates.IsZero() {
					t.Fatalf("Expected the ingredient's carbohydrates to be 0, received %v",
						ingredient.Carbohydrates)
				}

				if calorieVal, ok := testCase.params[dto.IngredientCalories]; ok {
					calories, err := CreateDecimal(calorieVal)
					if err != nil {
						t.Fatal(err)
					}
					rtnVal := ingredient.Calories.Mul(amount).Round(1)
					if !calories.Round(1).Equal(rtnVal) {
						t.Fatalf("Expected ingredient calories to be %v, received %v",
							calories, rtnVal)
					}
				} else if !ingredient.Calories.IsZero() {
					t.Fatalf("Expected the ingredient's calories to be 0, received %v",
						ingredient.Calories)
				}

				if unitType, ok := testCase.params[dto.IngredientUnitType]; ok {
					if unitType != ingredient.Unittype {
						t.Fatalf("Expected unit type to be %v, received %v", unitType, ingredient.Unittype)
					}

				} else if ingredient.Unittype {
					t.Fatal("Expected unit type to default to false, received true")

				}

				if brand, ok := testCase.params[dto.IngredientBrand]; ok {
					if ingredient.Brand != brand {
						t.Fatalf("Expected the brand to be %v, received %v", brand, ingredient.Brand)
					}
				} else if !(ingredient.Brand == "") {
					t.Fatalf("Expected the brand to be empty, received %v", ingredient.Brand)
				}

				if !ingredient.Amount.Equal(decimal.NewFromInt(1)) {
					t.Fatalf("Returned amount should be 1, received %v", ingredient.Amount)
				}

				servings := testCase.params[dto.IngredientServings].([]dto.ServingInfo)
				for _, serving := range servings {
					found := false
					for _, rtnServing := range ingredient.Servings {
						if rtnServing.Name == serving.Name && rtnServing.Quantity.Equal(serving.Quantity) {
							found = true
							break
						}
					}
					if !found {
						t.Fatalf("Couldn't find serving info for [%v], in [%v]", serving, ingredient.Servings)
					}
				}
			} else {
				res.Body.Close()
			}
		})
	}
}

func TestGetIngredient(t *testing.T) {
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
		testName      string
		ingredientsId int64
		ingredient    dto.IngredientResponse
		statusCode    int
	}{
		{
			testName:      "Successfully get ingredient by Id",
			ingredientsId: ingredient.Id,
			ingredient:    ingredient,
			statusCode:    http.StatusOK,
		},
		{
			testName:      "Unsuccessfully get nonexistent ingredient",
			ingredientsId: 1334,
			statusCode:    http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.GetIngredient(token, testCase.ingredientsId)
			if res.StatusCode != testCase.statusCode {
				buffer := make([]byte, 1000)
				res.Body.Read(buffer)
				t.Log(string(buffer))
				res.Body.Close()
				t.Fatalf("Expected get ingredient to return %d, received %d", testCase.statusCode, res.StatusCode)
			}

			if res.StatusCode == http.StatusOK {
				rtnIngredient := helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusOK)
				assert.Equal(t, ingredient, rtnIngredient, "The return ingredient should equal the previously created ingredient")
			} else {
				res.Body.Close()
			}

		})
	}
}

func TestGetIngredients(t *testing.T) {
	helpers.LoadDotEnv()

	member := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "mL",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "dto.IngredientServings",
					Quantity: decimal.Zero,
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
					Name:     "dto.IngredientServings",
					Quantity: decimal.Zero,
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
				},
			},
		},
	}
	foods := make(map[string]dto.IngredientResponse, len(foodInfos))

	res := client.CreateUser(member.name, member.email, member.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(member.email, member.password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)
	token := loginResponse.Token

	for name, infoMap := range foodInfos {
		res := client.CreateIngredient(token, infoMap)
		foods[name] = helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)
	}

	testCases := []struct {
		testName, query     string
		expectedIngredients []dto.IngredientResponse
		statusCode          int
	}{
		{
			testName: "Query Beans return the 2 bean products",
			query:    "Beans",
			expectedIngredients: []dto.IngredientResponse{
				foods["Black Beans"],
				foods["Green Beans"],
			},
			statusCode: http.StatusOK,
		},
		{
			testName: "Prefix query B returns the bean products and chicken breast",
			query:    "B",
			expectedIngredients: []dto.IngredientResponse{
				foods["Black Beans"],
				foods["Green Beans"],
				foods["Chicken Breast"],
			},
			statusCode: http.StatusOK,
		},
		{
			testName: "Query that matches the ingredients brand, returns the ingredients",
			query:    "Food Lion",
			expectedIngredients: []dto.IngredientResponse{
				foods["Sour Cream Grade A"],
				foods["Egg Whites"],
				foods["Whole Milk Greek Yogurt"],
				foods["Enriched Long Grain White Rice"],
			},
			statusCode: http.StatusOK,
		},
		{
			testName:   "Empty query returns nothing",
			statusCode: http.StatusNotFound,
		},
		{
			testName:   "Query that doesn't match any ingredients returns nothing",
			query:      "cake",
			statusCode: http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.GetIngredients(token, testCase.query)
			if res.StatusCode != testCase.statusCode {
				t.Fatalf("Expected get ingredients to return %d, received %d", testCase.statusCode, res.StatusCode)
			}
			if testCase.statusCode == http.StatusOK {
				ingredients := helpers.GetResponseObject[[]dto.IngredientResponse](t, res, http.StatusOK)
				assert.ElementsMatch(t, testCase.expectedIngredients, ingredients)

			} else {
				res.Body.Close()
			}
		})
	}
}

func TestUpdateIngredient(t *testing.T) {
	helpers.LoadDotEnv()

	member := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "mL",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "dto.IngredientServings",
					Quantity: decimal.Zero,
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
					Name:     "dto.IngredientServings",
					Quantity: decimal.Zero,
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
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
					Name:     "grams",
					Quantity: decimal.NewFromInt(1),
				},
			},
		},
	}
	foods := make(map[string]dto.IngredientResponse, len(foodInfos))

	res := client.CreateUser(member.name, member.email, member.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(member.email, member.password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)
	token := loginResponse.Token

	for name, infoMap := range foodInfos {
		res := client.CreateIngredient(token, infoMap)
		foods[name] = helpers.GetResponseObject[dto.IngredientResponse](t, res, http.StatusCreated)
	}

	testCases := []struct {
		testName           string
		ingredientId       int64
		expectedIngredient dto.IngredientResponse
		payload            map[string]any
		statusCode         int
	}{
		{
			testName:     "Successfully updated Ingredient Name",
			ingredientId: foods["Sour Cream Grade A"].Id,
			expectedIngredient: updateIngredientResponse(foods["Sour Cream Grade A"], map[dto.IngredientKeys]any{
				dto.IngredientName: "Sour Cream Grade B",
			}, decimal.Zero),
			statusCode: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {

		})
	}
}

func updateIngredientResponse(ing dto.IngredientResponse, updates map[dto.IngredientKeys]any, amount decimal.Decimal) dto.IngredientResponse {
	for key, value := range updates {
		switch key {
		case dto.IngredientName:
			ing.Name = value.(string)
		case dto.IngredientFats:
			decimalVal, _ := CreateDecimal(value)
			ing.Fats = decimalVal.Div(amount)
		case dto.IngredientCarbs:
			decimalVal, _ := CreateDecimal(value)
			ing.Carbohydrates = decimalVal.Div(amount)
		case dto.IngredientProtein:
			decimalVal, _ := CreateDecimal(value)
			ing.Protein = decimalVal.Div(amount)
		case dto.IngredientCalories:
			decimalVal, _ := CreateDecimal(value)
			ing.Calories = decimalVal.Div(amount)
		case dto.IngredientUnitType:
			ing.Unittype = dto.Measurement(value.(bool))
		case dto.IngredientServings:
			ing.Servings = value.([]dto.ServingInfo)
		case dto.IngredientBrand:
			ing.Brand = value.(string)
		}
	}

	return ing
}
