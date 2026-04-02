package integration

import (
	"github.com/jmservic/feast/integration_tests/dto"
	"github.com/jmservic/feast/integration_tests/helpers"
	"net/http"
	"testing"
)

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
	res.Body.Close()

	res = dto.LoginUser(t, feastUrl, email, password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

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

// Test cases -
// Household owner attempts to update - successful
// Non household member attempts to update - failure
// household member who isn't the owner attempts to update - failure
/*func TestUpdateHousehold(t *testing.T) {
	helpers.LoadDotEnv()

	owner := UserInfo{
		name: "jonathan",
		email: "jon@example.com",
		password: "very-secret!",
	}
	nonMember := UserInfo{
		name: "daron",
		email: "daron@example.com",
		password: "ab-city",
	}
	//member := UserInfo{}

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })

} */
