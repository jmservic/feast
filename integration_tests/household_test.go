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
	sut := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	// assert
	dto.ValidateHouseholdCreateResponse(t, sut, householdName)
	// add in check for a new household member after we get to those endpoints.
}

// Test cases -
// Household owner attempts to update - successful
// Non household member attempts to update - failure
// household member who isn't the owner attempts to update - failure
func TestUpdateHousehold(t *testing.T) {
	helpers.LoadDotEnv()

	householdName := "Service family"

	owner := UserInfo{
		name:     "jonathan",
		email:    "jon@example.com",
		password: "very-secret!",
	}

	nonMember := UserInfo{
		name:     "daron",
		email:    "daron@example.com",
		password: "ab-city",
	}

	//member := UserInfo{}

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })

	// create the owner
	res := dto.CreateUser(t, feastUrl, owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// create the non member
	res = dto.CreateUser(t, feastUrl, nonMember.name, nonMember.email, nonMember.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// login as users
	//Owner
	res = dto.LoginUser(t, feastUrl, owner.email, owner.password)
	ownerTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	//non member
	res = dto.LoginUser(t, feastUrl, nonMember.email, nonMember.password)
	nonMemberTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	// create the household
	res = dto.CreateHousehold(t, feastUrl, ownerTokenRes.Token, householdName)
	householdInfo := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	testCases := []struct {
		newName, token, testName string
		householdInfo            *dto.HouseholdResponse
		responseCode             int
	}{
		{
			newName:       "Francis-Service Family",
			token:         ownerTokenRes.Token,
			testName:      "Successful update by owner",
			householdInfo: &householdInfo,
			responseCode:  http.StatusOK,
		},
		{
			newName:       "Fradulent! Family",
			token:         nonMemberTokenRes.Token,
			testName:      "Failed update by non member",
			householdInfo: &householdInfo,
			responseCode:  http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			//Test Update
			res := dto.UpdateHousehold(t, feastUrl, testCase.householdInfo.Id.String(), testCase.newName, testCase.token)
			defer res.Body.Close()
			if res.StatusCode != testCase.responseCode {
				t.Fatalf("Expected Update Household to return %d code, received %d", testCase.responseCode, res.StatusCode)
			}
			if testCase.responseCode == http.StatusOK {
				var updatedHouseholdInfo dto.HouseholdResponse
				helpers.DecodeJSONResponse(&updatedHouseholdInfo, res.Body, t)
				if updatedHouseholdInfo.Name != testCase.newName {
					t.Fatal("Expected the household name to be updated!")
				} else {
					householdInfo.Name = testCase.newName
				}
			} else {
				//Check that the name stayed the same
				getRes := dto.GetHousehold(t, feastUrl, testCase.householdInfo.Id.String(), testCase.token)
				if getRes.StatusCode != http.StatusOK {
					t.Fatalf("Expected Get Household to return %d code, received %d", http.StatusOK, getRes.StatusCode)
				}

				var householdInfo dto.HouseholdResponse
				helpers.DecodeJSONResponse(&householdInfo, res.Body, t)

				if householdInfo.Name != testCase.householdInfo.Name {
					t.Fatalf("Expected the household name to remain the same. It was changed from \"%s\" to \"%s\"", householdInfo.Name, testCase.householdInfo.Name)
				}
			}
		})
	}

}
