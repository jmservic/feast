package integration

import (
	//	"github.com/google/uuid"
	"errors"
	"net/http"
	"testing"

	"github.com/jmservic/feast/integration_tests/dto"
	"github.com/jmservic/feast/integration_tests/helpers"
	"github.com/shopspring/decimal"
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
				t.Log(ingredient)

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
