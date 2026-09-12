package integration

import (
	"github.com/jmservic/feast/integration_tests/dto"
	"github.com/jmservic/feast/integration_tests/helpers"
	"io"
	"net/http"
	"testing"
)

func TestCreateHousehold(t *testing.T) {
	// arrange
	helpers.LoadDotEnv()
	name := "jonathan"
	email := "Jon@example.com"
	password := "very-secret!"
	householdName := "Service family"

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	res := client.CreateUser(name, email, password)

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(email, password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	// act
	res = client.CreateHousehold(loginResponse.Token, householdName)
	sut := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	// assert
	dto.ValidateHouseholdCreateResponse(t, sut, householdName)
	// add in check for a new household member after we get to those endpoints.
}

// Add tests for the owner and a random member trying to create a new household (should fail)
func TestCreateHouseholdForUserInAHousehold(t *testing.T) {
	// arrange
	helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}
	member := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	// create and login users
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(member.name, member.email, member.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(owner.email, owner.password)
	ownerLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.LoginUser(member.email, member.password)
	memberLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	// create the household
	res = client.CreateHousehold(ownerLoginResponse.Token, "Service Family")
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	// invite the member to the household
	res = client.InviteUserToHousehold(ownerLoginResponse.Token, memberLoginResponse.Id, householdCreationResponse.Id, nil)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected the invite to return status created, got :%d", res.StatusCode)
	}

	// accept the invite
	res = client.HandleInvite(memberLoginResponse.Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(res.Body)
		bodyStr := string(body)
		t.Fatalf("Expected the handle invite to return no content, got :%d. msg: %s", res.StatusCode, bodyStr)

	}

	//Attempt to create the households
	res = client.CreateHousehold(ownerLoginResponse.Token, "Evicres Household")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected the create household to return bad request, but received %d", res.StatusCode)
	}

	res = client.CreateHousehold(memberLoginResponse.Token, "Evicres Household")
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected the create household to return bad request, but received %d", res.StatusCode)
	}
}

func TestGetHousehold(t *testing.T) {
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
	client := dto.NewClient(t, feastUrl)

	// create the owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// create the non member
	res = client.CreateUser(nonMember.name, nonMember.email, nonMember.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// login as users
	//Owner
	res = client.LoginUser(owner.email, owner.password)
	ownerTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	//non member
	res = client.LoginUser(nonMember.email, nonMember.password)
	nonMemberTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	// create the household
	res = client.CreateHousehold(ownerTokenRes.Token, householdName)
	householdInfo := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	testCases := []struct {
		testName, token string
		householdInfo   *dto.HouseholdResponse
	}{
		{
			token:         ownerTokenRes.Token,
			testName:      "Owner gets household information",
			householdInfo: &householdInfo,
		},
		{
			token:         nonMemberTokenRes.Token,
			testName:      "Non member gets household information",
			householdInfo: &householdInfo,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.GetHousehold(testCase.householdInfo.Id.String(), testCase.token)
			householdResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusOK)
			if householdResponse.Id != testCase.householdInfo.Id || householdResponse.Name != testCase.householdInfo.Name {
				t.Fatal("household response differs from stored household info!")
			}
		})
	}
}

func TestUpdateHousehold(t *testing.T) {
	helpers.LoadDotEnv()

	householdName := "Service family"

	owner := UserInfo{
		name:     "jonathan",
		email:    "jon@example.com",
		password: "very-secret!",
	}

	otherOwner := UserInfo{
		name:     "joey",
		email:    "Joey@example.com",
		password: "fAcE-PaIn",
	}

	nonMember := UserInfo{
		name:     "daron",
		email:    "daron@example.com",
		password: "ab-city",
	}

	member := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	// create the owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(otherOwner.name, otherOwner.email, otherOwner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(member.name, member.email, member.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// create the non member
	res = client.CreateUser(nonMember.name, nonMember.email, nonMember.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// login as users
	//Owner
	res = client.LoginUser(owner.email, owner.password)
	ownerTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	res = client.LoginUser(otherOwner.email, otherOwner.password)
	otherOwnerTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	res = client.LoginUser(member.email, member.password)
	memberLoginRes := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//non member
	res = client.LoginUser(nonMember.email, nonMember.password)
	nonMemberTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	// create the household
	res = client.CreateHousehold(ownerTokenRes.Token, householdName)
	householdInfo := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	res = client.CreateHousehold(otherOwnerTokenRes.Token, "Evicres Family")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected household creation to return status created, received :%d", res.StatusCode)
	}
	res.Body.Close()

	//Invite, accept and promote member
	res = client.CreateHouseholdMember(ownerTokenRes.Token, member.name, householdInfo.Id, &memberLoginRes.Id)
	memberInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	res = client.HandleInvite(memberLoginRes.Token, householdInfo.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the handle invite to return status no content, received :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.UpdateHouseholdMember(ownerTokenRes.Token, memberInfo.Id, memberLoginRes.Id,
		memberInfo.Name, 2)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the update household member to return status no content, received :%d", res.StatusCode)
	}
	res.Body.Close()

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
		{
			newName:       "LET'S GOOO FAM",
			token:         memberLoginRes.Token,
			testName:      "Failed update by administrator member",
			householdInfo: &householdInfo,
			responseCode:  http.StatusForbidden,
		},
		{
			newName:       "I own this household now",
			token:         otherOwnerTokenRes.Token,
			testName:      "Failed update by another household owner",
			householdInfo: &householdInfo,
			responseCode:  http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			//Test Update
			res := client.UpdateHousehold(testCase.householdInfo.Id.String(), testCase.newName, testCase.token)
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
				getRes := client.GetHousehold(testCase.householdInfo.Id.String(), testCase.token)
				defer getRes.Body.Close()
				if getRes.StatusCode != http.StatusOK {
					t.Fatalf("Expected Get Household to return %d code, received %d", http.StatusOK, getRes.StatusCode)
				}

				var householdInfo dto.HouseholdResponse
				helpers.DecodeJSONResponse(&householdInfo, getRes.Body, t)

				if householdInfo.Name != testCase.householdInfo.Name {
					t.Fatalf("Expected the household name to remain the same. It was changed from \"%s\" to \"%s\"", testCase.householdInfo.Name, householdInfo.Name)
				}
			}
		})
	}
}

// Test cases -
// Household owner attempts to delete - successful
// Non household member attempts to delete - failure
// household member who isn't the owner attempts to delete - failure
func TestDeleteHousehold(t *testing.T) {
	helpers.LoadDotEnv()

	householdName := "Service family"

	owner := UserInfo{
		name:     "jonathan",
		email:    "jon@example.com",
		password: "very-secret!",
	}

	otherOwner := UserInfo{
		name:     "joey",
		email:    "Joey@example.com",
		password: "fAcE-PaIn",
	}

	nonMember := UserInfo{
		name:     "daron",
		email:    "daron@example.com",
		password: "ab-city",
	}

	member := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	// create the owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(otherOwner.name, otherOwner.email, otherOwner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(member.name, member.email, member.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// create the non member
	res = client.CreateUser(nonMember.name, nonMember.email, nonMember.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	// login as users
	//Owner
	res = client.LoginUser(owner.email, owner.password)
	ownerTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	res = client.LoginUser(otherOwner.email, otherOwner.password)
	otherOwnerTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	res = client.LoginUser(member.email, member.password)
	memberLoginRes := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//non member
	res = client.LoginUser(nonMember.email, nonMember.password)
	nonMemberTokenRes := helpers.GetResponseObject[dto.TokenResponse](t, res, http.StatusOK)

	// create the household
	res = client.CreateHousehold(ownerTokenRes.Token, householdName)
	householdInfo := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	res = client.CreateHousehold(otherOwnerTokenRes.Token, "Evicres Family")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected household creation to return status created, received :%d", res.StatusCode)
	}
	res.Body.Close()

	//Invite, accept and promote member
	res = client.CreateHouseholdMember(ownerTokenRes.Token, member.name, householdInfo.Id, &memberLoginRes.Id)
	memberInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	res = client.HandleInvite(memberLoginRes.Token, householdInfo.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the handle invite to return status no content, received :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.UpdateHouseholdMember(ownerTokenRes.Token, memberInfo.Id, memberLoginRes.Id,
		memberInfo.Name, 2)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the update household member to return status no content, received :%d", res.StatusCode)
	}
	res.Body.Close()

	testCases := []struct {
		token, testName string
		householdInfo   *dto.HouseholdResponse
		responseCode    int
	}{
		{
			token:         nonMemberTokenRes.Token,
			testName:      "Failed delete by non member",
			householdInfo: &householdInfo,
			responseCode:  http.StatusForbidden,
		},
		{
			token:         memberLoginRes.Token,
			testName:      "Failed delete by administrator member",
			householdInfo: &householdInfo,
			responseCode:  http.StatusForbidden,
		},
		{
			token:         otherOwnerTokenRes.Token,
			testName:      "Failed delete by other household owner",
			householdInfo: &householdInfo,
			responseCode:  http.StatusForbidden,
		},
		{
			token:         ownerTokenRes.Token,
			testName:      "Successful delete by owner",
			householdInfo: &householdInfo,
			responseCode:  http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.DeleteHousehold(testCase.householdInfo.Id.String(), testCase.token)
			defer res.Body.Close()
			if res.StatusCode != testCase.responseCode {
				t.Fatalf("Expected Delete Household to return %d code, received %d", testCase.responseCode, res.StatusCode)
			}

			getRes := client.GetHousehold(testCase.householdInfo.Id.String(), testCase.token)
			defer getRes.Body.Close()

			switch res.StatusCode {
			case http.StatusOK: //Household should be deleted
				if getRes.StatusCode != http.StatusNotFound {
					t.Fatalf("Expected Not Found status code, received: %d", getRes.StatusCode)
				}
			case http.StatusForbidden: //Household should still exist
				if getRes.StatusCode != http.StatusOK {
					t.Fatalf("Expected OK status code, received: %d", getRes.StatusCode)
				}
			default:
				t.Fatalf("Unexpected response code: %d", res.StatusCode)
			}

		})
	}
}
