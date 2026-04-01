package integration

import (
	"github.com/jmservic/feast/integration_tests/dto"
	"github.com/jmservic/feast/integration_tests/helpers"
	"net/http"
	"testing"
)

// Test for when we delete the user.... who is the owner of a household
func TestCreateNewHousehold(t *testing.T) {
	// arrange
	helpers.LoadDotEnv()
	name := "jonathan"
	email := "Jon@example.com"
	password := "very-secret!"
	householdName := "Service family"

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })

	res := dto.CreateUser(t, feastUrl, name, email, password)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}

	userCreateResponse := dto.UserCreateResponse{}
	helpers.DecodeJSONResponse(&userCreateResponse, res.Body, t)
	res.Body.Close()

	res = dto.LoginUser(t, feastUrl, email, password)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Failed to login as the test user - Status code: %d", res.StatusCode)
	}

	loginResponse := dto.UserLoginResponse{}
	helpers.DecodeJSONResponse(&loginResponse, res.Body, t)
	res.Body.Close()

	// act
	res = dto.CreateHousehold(t, feastUrl, loginResponse.Token, householdName)
	defer res.Body.Close()

	// assert
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}

	sut := dto.HouseholdCreateResponse{}
	helpers.DecodeJSONResponse(&sut, res.Body, t)

	dto.ValidateHouseholdCreateResponse(t, sut, householdName)
	// add in check for a new household member after we get to those endpoints.
}
