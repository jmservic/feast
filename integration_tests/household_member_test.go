package integration

import (
	"github.com/google/uuid"
	"github.com/jmservic/feast/integration_tests/dto"
	"github.com/jmservic/feast/integration_tests/helpers"
	"net/http"
	"testing"
)

func TestCreateHouseholdMember(t *testing.T) {
	// arrange
	helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}
	nonMember := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}
	otherOwner := UserInfo{
		name:     "joey",
		email:    "Joey@example.com",
		password: "fAcE-PaIn",
	}

	householdName := "Service family"

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(nonMember.name, nonMember.email, nonMember.password)
	createUserResponse := helpers.GetResponseObject[dto.UserCreateResponse](t, res, http.StatusCreated)

	res = client.LoginUser(owner.email, owner.password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.CreateHousehold(loginResponse.Token, householdName)
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	// create other household
	res = client.CreateUser(otherOwner.name, otherOwner.email, otherOwner.password)
	createOtherOwnerResponse := helpers.GetResponseObject[dto.UserCreateResponse](t, res, http.StatusCreated)

	res = client.LoginUser(otherOwner.email, otherOwner.password)
	otherOwnerLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.CreateHousehold(otherOwnerLoginResponse.Token, "Miller family")
	secondHouseholdResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	testCases := []struct {
		testName, memberName string
		userId               *uuid.UUID
		householdId          uuid.UUID
		status               int
		memberDiff           int
	}{
		{
			testName:    "Create User without an User Id",
			memberName:  "Johnny Jr.",
			userId:      nil,
			householdId: householdCreationResponse.Id,
			status:      http.StatusCreated,
			memberDiff:  1,
		},
		{
			testName:    "Create User with an User Id",
			memberName:  "Cassidy",
			userId:      &createUserResponse.Id,
			householdId: householdCreationResponse.Id,
			status:      http.StatusCreated,
			memberDiff:  1,
		},
		{
			testName:    "Create User with an User Id in another household",
			memberName:  "Joey",
			userId:      &createOtherOwnerResponse.Id,
			householdId: householdCreationResponse.Id,
			status:      http.StatusBadRequest,
			memberDiff:  0,
		},
		{
			testName:    "Creating Member in another household fails",
			memberName:  "Schemer",
			userId:      nil,
			householdId: secondHouseholdResponse.Id,
			status:      http.StatusForbidden,
			memberDiff:  0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			// Check members before
			res = client.GetHouseholdMembers(loginResponse.Token, testCase.householdId)
			originalMembers := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)

			// act
			res = client.CreateHouseholdMember(loginResponse.Token, testCase.memberName,
				testCase.householdId, testCase.userId)
			res.Body.Close()

			// assert
			if res.StatusCode != testCase.status {
				t.Fatalf("Expected status created for member creation, got: %d.", res.StatusCode)
			}

			//Check members after
			res = client.GetHouseholdMembers(loginResponse.Token, testCase.householdId)
			updatedMembers := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)

			if len(updatedMembers) != len(originalMembers)+testCase.memberDiff {
				t.Fatalf("Expected length of updated members list to be %d, but got %d",
					len(originalMembers)+testCase.memberDiff, len(updatedMembers))
			}

			if testCase.memberDiff == 0 { //we don't expect a new member
				return
			}

			foundMember := false
			var memberInfo dto.HouseholdMemberResponse
			for _, member := range updatedMembers {
				if member.Name == testCase.memberName {
					foundMember = true
					memberInfo = member
					break
				}
			}

			if !foundMember {
				t.Fatalf("Did find a member with a name of %s", testCase.memberName)
			}

			if testCase.userId != nil {
				//check for the invite
				res = client.GetInvites(loginResponse.Token)
				invites := helpers.GetResponseObject[[]dto.InviteResponse](t, res, http.StatusOK)
				t.Logf("length of invites: %d", len(invites))
				foundInvite := false
				for _, invite := range invites {
					t.Log(invite.HouseholdId, invite.HouseholdMemberId, invite.InviterId, invite.InviteeId)
					if invite.HouseholdId == testCase.householdId &&
						invite.InviterId == loginResponse.Id &&
						invite.InviteeId == *testCase.userId &&
						*invite.HouseholdMemberId == memberInfo.Id {
						foundInvite = true
						break
					}
				}
				if !foundInvite {
					t.Fatalf("Did not find an invite with InviterId, %v, and InviteeId, %v",
						loginResponse.Id, testCase.userId)
				}
			}
		})
	}
}

func TestUpdateHouseholdMember(t *testing.T) {
	// arrange
	helpers.LoadDotEnv()

	users := map[string]struct {
		UserInfo
		loginResponse dto.UserLoginResponse
		memberInfo    dto.HouseholdMemberResponse
	}{
		"owner": {
			UserInfo: UserInfo{
				name:     "jonathan",
				email:    "Jon@example.com",
				password: "very-secret",
			},
		},
		"member": {
			UserInfo: UserInfo{
				name:     "cassidy",
				email:    "Cass@example.com",
				password: "kalina",
			},
		},
		"thirdMember": {
			UserInfo: UserInfo{
				name:     "sean",
				email:    "Shawn@example.com",
				password: "otouto",
			},
		},
		"fourthMember": {
			UserInfo: UserInfo{
				name:     "Sha'Myah",
				email:    "ShaMyan@example.com",
				password: "ShiMai.Imouto",
			},
		},
		"secondOwner": {
			UserInfo: UserInfo{
				name:     "mark",
				email:    "Mark@example.com",
				password: "very_secret",
			},
		},
		"nonMember": {
			UserInfo: UserInfo{
				name:     "jayce",
				email:    "Jayce@example.com",
				password: "passw0rd",
			},
		},
	}

	householdName := "Service family"

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	//create and login the users
	for name, user := range users {
		//create the user
		res := client.CreateUser(user.name, user.email, user.password)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status created for %v, got: %d", user.name, res.StatusCode)
		}
		res.Body.Close()

		//login the user
		res = client.LoginUser(user.email, user.password)
		user.loginResponse = helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)
		users[name] = user

	}

	//create the different households
	res := client.CreateHousehold(users["owner"].loginResponse.Token, householdName)
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	res = client.CreateHousehold(users["secondOwner"].loginResponse.Token, "Ecivres")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created for the second household, got: %d", res.StatusCode)
	}
	res.Body.Close()

	//Invite members to household
	memberId := users["member"].loginResponse.Id
	thirdMemberId := users["thirdMember"].loginResponse.Id
	fourthMemberId := users["fourthMember"].loginResponse.Id

	res = client.CreateHouseholdMember(users["owner"].loginResponse.Token, users["member"].name, householdCreationResponse.Id,
		&memberId)
	memberResponseTemp := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)
	userTemp := users["member"]
	userTemp.memberInfo = memberResponseTemp
	users["member"] = userTemp

	res = client.CreateHouseholdMember(users["owner"].loginResponse.Token, users["thirdMember"].name, householdCreationResponse.Id,
		&thirdMemberId)
	memberResponseTemp = helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)
	userTemp = users["thirdMember"]
	userTemp.memberInfo = memberResponseTemp
	users["thirdMember"] = userTemp

	res = client.CreateHouseholdMember(users["owner"].loginResponse.Token, users["fourthMember"].name, householdCreationResponse.Id,
		&fourthMemberId)
	memberResponseTemp = helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)
	userTemp = users["fourthMember"]
	userTemp.memberInfo = memberResponseTemp
	users["fourthMember"] = userTemp

	//Accept member invite
	res = client.HandleInvite(users["member"].loginResponse.Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error accepting the member invite, expected NoContent, got %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.HandleInvite(users["thirdMember"].loginResponse.Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error accepting the member invite, expected NoContent, got %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.HandleInvite(users["fourthMember"].loginResponse.Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error accepting the member invite, expected NoContent, got %d", res.StatusCode)
	}
	res.Body.Close()

	testCases := []struct {
		testName, token, newName string
		memberId, updatedUserId  uuid.UUID
		role, responseCode       int
	}{
		{
			token:        users["owner"].loginResponse.Token,
			memberId:     users["member"].memberInfo.Id,
			role:         2,
			testName:     "Successfully update role of a member",
			responseCode: http.StatusNoContent,
		},
		{
			token:        users["owner"].loginResponse.Token,
			memberId:     users["thirdMember"].memberInfo.Id,
			testName:     "Unable to update an user's role to a nonexistent role",
			role:         200,
			responseCode: http.StatusBadRequest,
		},
		{
			token:        users["member"].loginResponse.Token,
			memberId:     users["member"].memberInfo.Id,
			newName:      "Casstadon",
			testName:     "User's able to successfully update their name",
			responseCode: http.StatusNoContent,
		},
		{
			token:        users["member"].loginResponse.Token,
			memberId:     users["member"].memberInfo.Id,
			role:         1,
			testName:     "User unable to increase their role",
			responseCode: http.StatusForbidden,
		},
		{
			token:        users["member"].loginResponse.Token,
			memberId:     users["member"].memberInfo.Id,
			role:         3,
			testName:     "User unable to decrease their role",
			responseCode: http.StatusForbidden,
		},
		{
			token:         users["member"].loginResponse.Token,
			memberId:      users["member"].memberInfo.Id,
			testName:      "User unable to change the associated userId",
			updatedUserId: users["nonMember"].loginResponse.Id,
			responseCode:  http.StatusForbidden,
		},
		{
			token:        users["member"].loginResponse.Token,
			memberId:     users["thirdMember"].memberInfo.Id,
			testName:     "Able to update a lower member",
			newName:      "Shawn",
			responseCode: http.StatusNoContent,
		},
		{
			token:        users["member"].loginResponse.Token,
			memberId:     users["thirdMember"].memberInfo.Id,
			testName:     "Unable to update lower member's role to your own",
			role:         2,
			responseCode: http.StatusForbidden,
		},
		{
			token:        users["fourthMember"].loginResponse.Token,
			memberId:     users["thirdMember"].memberInfo.Id,
			testName:     "User unable to update another member of the same role",
			newName:      "Charles",
			responseCode: http.StatusForbidden,
		},
		{
			token:        users["secondOwner"].loginResponse.Token,
			memberId:     users["member"].memberInfo.Id,
			testName:     "An owner is unable to update a member in another household",
			newName:      "Punk!!",
			responseCode: http.StatusForbidden,
		},
		{
			token:        users["nonMember"].loginResponse.Token,
			memberId:     users["member"].memberInfo.Id,
			testName:     "A non member is unable to update a member in a household",
			newName:      "honestly-I-Tried",
			responseCode: http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			//get the member's initial state
			res := client.GetHouseholdMember(testCase.token, testCase.memberId)
			initialMemberState := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)

			//send the update request
			name := testCase.newName
			userId := testCase.updatedUserId
			role := testCase.role

			if name == "" {
				name = initialMemberState.Name
			}
			if userId == uuid.Nil {
				userId = *initialMemberState.UserId
			}
			if role == 0 {
				role = initialMemberState.Role
			}

			res = client.UpdateHouseholdMember(testCase.token, testCase.memberId,
				userId, name, role)
			defer res.Body.Close()

			if res.StatusCode != testCase.responseCode {
				t.Fatalf("Expected the Update endpoint to return %d, but received %d status code",
					testCase.responseCode, res.StatusCode)
			}

			//get the updated member
			res = client.GetHouseholdMember(testCase.token, testCase.memberId)
			updatedMemberState := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)

			if testCase.responseCode == http.StatusNoContent {
				if *updatedMemberState.UserId != userId {
					t.Fatalf("Expected an userId of %v, received %v",
						userId, *updatedMemberState.UserId)
				}
				if updatedMemberState.Name != name {
					t.Fatalf("Expected a name of %v, received %v",
						name, updatedMemberState.Name)
				}
				if updatedMemberState.Role != role {
					t.Fatalf("Expected a role of %v, received %v",
						role, updatedMemberState.Role)
				}
			} else {
				if *updatedMemberState.UserId != *initialMemberState.UserId {
					t.Fatalf("Expected an userId of %v, received %v",
						*initialMemberState.UserId, *updatedMemberState.UserId)
				}
				if updatedMemberState.Name != initialMemberState.Name {
					t.Fatalf("Expected a name of %v, received %v",
						initialMemberState.Name, updatedMemberState.Name)
				}
				if updatedMemberState.Role != initialMemberState.Role {
					t.Fatalf("Expected a role of %v, received %v",
						initialMemberState.Role, updatedMemberState.Role)
				}
			}
		})
	}
}

func TestUpdateNonExistingMember(t *testing.T) {
	helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}

	nonExistentMemberId := uuid.New()
	newName := "Kelvin"
	householdName := "Service family"

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	// create owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	// login
	res = client.LoginUser(owner.email, owner.password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//create household
	res = client.CreateHousehold(loginResponse.Token, householdName)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected household creation to return status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.UpdateHouseholdMember(loginResponse.Token, nonExistentMemberId,
		uuid.Nil, newName, 1)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected the update household member response to be %d, but received %d", http.StatusNotFound,
			res.StatusCode)
	}
}

func TestGetHouseholdMember(t *testing.T) {
	helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}

	householdName := "Service family"
	memberName := "cassidy"
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	// create owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	// login
	res = client.LoginUser(owner.email, owner.password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//create household
	res = client.CreateHousehold(loginResponse.Token, householdName)
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	//create member
	res = client.CreateHouseholdMember(loginResponse.Token, memberName, householdCreationResponse.Id, nil)
	memberCreationResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	res = client.GetHouseholdMember(loginResponse.Token, memberCreationResponse.Id)
	getMemberResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)

	client.ValidateHouseholdMemberResponse(memberCreationResponse, getMemberResponse)
}

func TestGetHouseholdMembers(t *testing.T) {
	helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}

	additionalMembers := []string{"shawn", "sha'dimond", "kamyah"}

	householdName := "Service family"
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	// create owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	// login
	res = client.LoginUser(owner.email, owner.password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//create household
	res = client.CreateHousehold(loginResponse.Token, householdName)
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	//create other members
	for _, name := range additionalMembers {
		res = client.CreateHouseholdMember(loginResponse.Token, name, householdCreationResponse.Id, nil)
		res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Expected household member %v to be created, received response of %d", name, res.StatusCode)
		}
	}

	// act finally
	res = client.GetHouseholdMembers(loginResponse.Token, householdCreationResponse.Id)
	members := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)

	// assert
	for _, name := range additionalMembers {
		found := false
		for _, member := range members {
			if member.Name == name {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Expected to find %v in the household members, but did not", name)
		}
	}
}

func TestDeleteHouseholdMember(t *testing.T) {
	helpers.LoadDotEnv()

	users := map[string]struct {
		UserInfo
		loginResponse dto.UserLoginResponse
		memberInfo    dto.HouseholdMemberResponse
	}{
		"owner": {
			UserInfo: UserInfo{
				name:     "jonathan",
				email:    "Jon@example.com",
				password: "very-secret",
			},
		},
		"member": {
			UserInfo: UserInfo{
				name:     "cassidy",
				email:    "Cass@example.com",
				password: "kalina",
			},
		},
		"thirdMember": {
			UserInfo: UserInfo{
				name:     "sean",
				email:    "Shawn@example.com",
				password: "otouto",
			},
		},
		"fourthMember": {
			UserInfo: UserInfo{
				name:     "Sha'Myah",
				email:    "ShaMyan@example.com",
				password: "ShiMai.Imouto",
			},
		},
		"secondOwner": {
			UserInfo: UserInfo{
				name:     "mark",
				email:    "Mark@example.com",
				password: "very_secret",
			},
		},
		"nonMember": {
			UserInfo: UserInfo{
				name:     "jayce",
				email:    "Jayce@example.com",
				password: "passw0rd",
			},
		},
	}

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)
	householdName := "Service family"

	//create and login the users
	for name, user := range users {
		//create the user
		res := client.CreateUser(user.name, user.email, user.password)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status created for %v, got: %d", user.name, res.StatusCode)
		}
		res.Body.Close()

		//login the user
		res = client.LoginUser(user.email, user.password)
		user.loginResponse = helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)
		users[name] = user

	}

	//create the different households
	res := client.CreateHousehold(users["owner"].loginResponse.Token, householdName)
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)
	res = client.GetHouseholdMembers(users["owner"].loginResponse.Token, householdCreationResponse.Id)
	membersTemp := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)
	userTemp := users["owner"]
	if len(membersTemp) > 0 {
		userTemp.memberInfo = membersTemp[0]
		users["owner"] = userTemp
	} else {
		t.Fatalf("Household somehow has no members...")
	}

	res = client.CreateHousehold(users["secondOwner"].loginResponse.Token, "Ecivres")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created for the second household, got: %d", res.StatusCode)
	}
	res.Body.Close()

	//Invite members to household
	memberId := users["member"].loginResponse.Id
	thirdMemberId := users["thirdMember"].loginResponse.Id
	fourthMemberId := users["fourthMember"].loginResponse.Id

	res = client.CreateHouseholdMember(users["owner"].loginResponse.Token, users["member"].name, householdCreationResponse.Id,
		&memberId)
	memberResponseTemp := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)
	userTemp = users["member"]
	userTemp.memberInfo = memberResponseTemp
	users["member"] = userTemp

	res = client.CreateHouseholdMember(users["owner"].loginResponse.Token, users["thirdMember"].name, householdCreationResponse.Id,
		&thirdMemberId)
	memberResponseTemp = helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)
	userTemp = users["thirdMember"]
	userTemp.memberInfo = memberResponseTemp
	users["thirdMember"] = userTemp

	res = client.CreateHouseholdMember(users["owner"].loginResponse.Token, users["fourthMember"].name, householdCreationResponse.Id,
		&fourthMemberId)
	memberResponseTemp = helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)
	userTemp = users["fourthMember"]
	userTemp.memberInfo = memberResponseTemp
	users["fourthMember"] = userTemp

	res = client.CreateHouseholdMember(users["owner"].loginResponse.Token, "Quincy", householdCreationResponse.Id, nil)
	nonUserMemberResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	//Accept member invite
	res = client.HandleInvite(users["member"].loginResponse.Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error accepting the member invite, expected NoContent, got %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.HandleInvite(users["thirdMember"].loginResponse.Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error accepting the member invite, expected NoContent, got %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.HandleInvite(users["fourthMember"].loginResponse.Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error accepting the member invite, expected NoContent, got %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.UpdateHouseholdMember(users["owner"].loginResponse.Token, users["member"].memberInfo.Id,
		users["member"].loginResponse.Id, users["member"].memberInfo.Name, 2)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error updating member's role, expected StatusNoContent, got %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.UpdateHouseholdMember(users["owner"].loginResponse.Token, users["thirdMember"].memberInfo.Id,
		users["thirdMember"].loginResponse.Id, users["thirdMember"].memberInfo.Name, 2)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Error updating member's role, expected StatusNoContent, got %d", res.StatusCode)
	}
	res.Body.Close()

	testCases := []struct {
		testName, token       string
		householdId, memberId uuid.UUID
		responseCode          int
	}{
		{
			testName:     "Unable to delete household member with a higher rank",
			token:        users["thirdMember"].loginResponse.Token,
			householdId:  users["member"].memberInfo.HouseholdId,
			memberId:     users["member"].memberInfo.Id,
			responseCode: http.StatusForbidden,
		},
		{
			testName:     "Unable to delete the household owner",
			token:        users["member"].loginResponse.Token,
			householdId:  users["owner"].memberInfo.HouseholdId,
			memberId:     users["owner"].memberInfo.Id,
			responseCode: http.StatusForbidden,
		},
		{
			testName:     "Unable to delete a member in another household",
			token:        users["secondOwner"].loginResponse.Token,
			householdId:  users["fourthMember"].memberInfo.HouseholdId,
			memberId:     users["fourthMember"].memberInfo.Id,
			responseCode: http.StatusForbidden,
		},
		{
			testName:     "Unable to delete a member in another household when you don't have a household",
			token:        users["nonMember"].loginResponse.Token,
			householdId:  users["fourthMember"].memberInfo.HouseholdId,
			memberId:     users["fourthMember"].memberInfo.Id,
			responseCode: http.StatusForbidden,
		},
		{
			testName:     "Unable to delete yourself lol",
			token:        users["member"].loginResponse.Token,
			householdId:  users["member"].memberInfo.HouseholdId,
			memberId:     users["member"].memberInfo.Id,
			responseCode: http.StatusForbidden,
		},
		{
			testName:     "Successfully delete household member",
			token:        users["owner"].loginResponse.Token,
			householdId:  nonUserMemberResponse.HouseholdId,
			memberId:     nonUserMemberResponse.Id,
			responseCode: http.StatusOK,
		},
		{
			testName:     "Successfully delete a household member with a linked user",
			token:        users["member"].loginResponse.Token,
			householdId:  users["fourthMember"].memberInfo.HouseholdId,
			memberId:     users["fourthMember"].memberInfo.Id,
			responseCode: http.StatusOK,
		},

		//Possibly put tests for different roles, yes, we want administrators and heads to only be able to delete.
		//currently managers are also able to delete.
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.GetHouseholdMembers(testCase.token, testCase.householdId)
			prevMembers := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)

			res = client.DeleteHouseholdMember(testCase.token, testCase.memberId)
			if res.StatusCode != testCase.responseCode {
				t.Fatalf("Expected member deletion to return %v, received %v",
					testCase.responseCode, res.StatusCode)
			}
			res.Body.Close()

			res = client.GetHouseholdMembers(testCase.token, testCase.householdId)
			currMembers := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)

			if testCase.responseCode == http.StatusOK {
				if len(currMembers) != len(prevMembers)-1 {
					t.Fatalf("Succussful Delete: expected the number of current members to be %d, got %d",
						len(prevMembers)-1, len(currMembers))
				}
				for _, member := range currMembers {
					if member.Id == testCase.memberId {
						t.Fatalf("Expected not to find member %v, but found him/her",
							testCase.memberId.String())
					}
				}
			} else {
				found := false
				if len(currMembers) != len(prevMembers) {
					t.Fatalf("Unsuccessful Delete: expected the number of current members to be %d, got %d",
						len(prevMembers), len(currMembers))
				}
				t.Logf("Length of currMembers = %d", len(currMembers))
				for _, member := range currMembers {
					t.Logf("%v", member.Id.String())
					if member.Id == testCase.memberId {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("Expected to find member %v, but did not find him/her",
						testCase.memberId.String())
				}
			}
		})
	}
}

func TestGetHouseholdInvites(t *testing.T) {
	helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}
	nonMember := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}

	householdName := "Service family"
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	// create owner
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	// login
	res = client.LoginUser(owner.email, owner.password)
	loginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//create household
	res = client.CreateHousehold(loginResponse.Token, householdName)
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	//create nonMember
	res = client.CreateUser(nonMember.name, nonMember.email, nonMember.password)
	createUserResponse := helpers.GetResponseObject[dto.UserCreateResponse](t, res, http.StatusCreated)

	//nonMember Login
	res = client.LoginUser(nonMember.email, nonMember.password)
	nonMemberLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//create the new member and invite the non member
	res = client.CreateHouseholdMember(loginResponse.Token, "cassidy", householdCreationResponse.Id, &createUserResponse.Id)
	createMemberResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	res = client.GetInvites(nonMemberLoginResponse.Token)
	invites := helpers.GetResponseObject[[]dto.InviteResponse](t, res, http.StatusOK)

	for _, invite := range invites {
		if invite.InviterId == loginResponse.Id && invite.InviteeId == nonMemberLoginResponse.Id && invite.HouseholdMemberId != nil &&
			*invite.HouseholdMemberId == createMemberResponse.Id && invite.HouseholdId == householdCreationResponse.Id {
			return
		}
	}

	t.Fatalf("Could not find an invite with InviterId: %v, InviteeId: %v, HouseholdId: %v, and HouseholdMemberId: %v",
		loginResponse.Id, nonMemberLoginResponse.Id, createMemberResponse.Id, householdCreationResponse.Id)
}

func TestInviteUserToHousehold(t *testing.T) {
	helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}
	secondOwner := UserInfo{
		name:     "mark",
		email:    "Mark@example.com",
		password: "very_secret",
	}
	member := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}
	nonMember := UserInfo{
		name:     "joey",
		email:    "Joey@example.com",
		password: "fAcE-PaIn",
	}
	secondNonMember := UserInfo{
		name:     "elijah",
		email:    "elijah@example.com",
		password: "secret-very-much",
	}

	householdName := "Service Family"
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	// create owners
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(secondOwner.name, secondOwner.email, secondOwner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	// login
	res = client.LoginUser(owner.email, owner.password)
	ownerLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.LoginUser(secondOwner.email, secondOwner.password)
	secondOwnerLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//create households
	res = client.CreateHousehold(ownerLoginResponse.Token, householdName)
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	res = client.CreateHousehold(secondOwnerLoginResponse.Token, "Ecivres family")
	secondHouseholdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	// create member/nonmembers and login
	res = client.CreateUser(member.name, member.email, member.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(nonMember.name, nonMember.email, nonMember.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(secondNonMember.name, secondNonMember.email, secondNonMember.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.LoginUser(member.email, member.password)
	memberLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.LoginUser(nonMember.email, nonMember.password)
	nonMemberLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.LoginUser(secondNonMember.email, secondNonMember.password)
	secondNonMemberLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//Create member for the member user
	res = client.CreateHouseholdMember(ownerLoginResponse.Token, member.name, householdCreationResponse.Id,
		&memberLoginResponse.Id)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	//Accept the invite
	res = client.HandleInvite(memberLoginResponse.Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status no content, got: %d", res.StatusCode)
	}
	res.Body.Close()

	//Create a member in the second household
	res = client.CreateHouseholdMember(secondOwnerLoginResponse.Token, nonMember.name,
		secondHouseholdCreationResponse.Id, nil)
	secondHouseholdMemberInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](
		t, res, http.StatusCreated)

	//Test Cases:
	testCases := []struct {
		testName, inviterToken, inviteeToken string
		householdId, userId                  uuid.UUID
		householdMemberId                    *uuid.UUID
		responseCode, inviteNum              int
	}{
		{
			testName:     "Invite an user not in a household",
			inviterToken: ownerLoginResponse.Token,
			inviteeToken: nonMemberLoginResponse.Token,
			householdId:  householdCreationResponse.Id,
			userId:       nonMemberLoginResponse.Id,
			responseCode: http.StatusCreated,
		},
		{ //This also serves as inviting the user to multiple households
			testName:          "Invite a user to a household for a particular member",
			inviterToken:      secondOwnerLoginResponse.Token,
			inviteeToken:      nonMemberLoginResponse.Token,
			householdId:       secondHouseholdCreationResponse.Id,
			userId:            nonMemberLoginResponse.Id,
			householdMemberId: &secondHouseholdMemberInfo.Id,
			responseCode:      http.StatusCreated,
		},
		{
			testName:     "Invite a user already in a household",
			inviterToken: secondOwnerLoginResponse.Token,
			inviteeToken: memberLoginResponse.Token,
			householdId:  secondHouseholdCreationResponse.Id,
			userId:       memberLoginResponse.Id,
			responseCode: http.StatusBadRequest,
		},
		{
			testName:     "Invite a user already in the same household",
			inviterToken: ownerLoginResponse.Token,
			inviteeToken: memberLoginResponse.Token,
			householdId:  householdCreationResponse.Id,
			userId:       memberLoginResponse.Id,
			responseCode: http.StatusBadRequest,
		},
		{
			testName:     "User unable to invite to the household",
			inviterToken: memberLoginResponse.Token,
			inviteeToken: secondNonMemberLoginResponse.Token,
			householdId:  householdCreationResponse.Id,
			userId:       secondNonMemberLoginResponse.Id,
			responseCode: http.StatusForbidden,
		},
		{
			testName:     "Invite an user to a household other than your own",
			inviterToken: ownerLoginResponse.Token,
			inviteeToken: secondNonMemberLoginResponse.Token,
			householdId:  secondHouseholdCreationResponse.Id,
			userId:       secondNonMemberLoginResponse.Id,
			responseCode: http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			foundInvite := func(invites []dto.InviteResponse) bool {
				for _, invite := range invites {
					if invite.HouseholdId == testCase.householdId {
						if invite.HouseholdMemberId == nil {
							if testCase.householdMemberId == nil {
								return true
							}
						} else {
							if testCase.householdMemberId != nil &&
								*invite.HouseholdMemberId == *testCase.householdMemberId {
								return true
							}

						}
					}
				}
				return false
			}

			res := client.GetInvites(testCase.inviteeToken)
			prevInvites := helpers.GetResponseObject[[]dto.InviteResponse](t, res, http.StatusOK)

			res = client.InviteUserToHousehold(testCase.inviterToken, testCase.userId, testCase.householdId,
				testCase.householdMemberId)

			if res.StatusCode != testCase.responseCode {
				t.Fatalf("Expected the invite endpoint to return a status code of %d, received %d",
					testCase.responseCode, res.StatusCode)
			}

			res = client.GetInvites(testCase.inviteeToken)
			currInvites := helpers.GetResponseObject[[]dto.InviteResponse](t, res, http.StatusOK)

			if testCase.responseCode == http.StatusCreated {
				if len(currInvites) != len(prevInvites)+1 {
					t.Fatalf("Expected the number to be %d, but received %d",
						len(prevInvites)+1, len(currInvites))
				}

				found := foundInvite(currInvites)

				if !found {
					t.Fatalf("Failed to find the corresponding invite")
				}
			} else {
				if len(currInvites) != len(prevInvites) {
					t.Fatalf("Expected the number of invites to remain the same, but changed from %d to %d",
						len(prevInvites), len(currInvites))
				}

				found := foundInvite(currInvites)

				if found {
					t.Fatalf("Found a corresponding invite when no invite should've been created")
				}
			}

		})
	}
}

func TestInviteUserMultipleTimesToHousehold(t *testing.T) {
	helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}
	user := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}

	householdName := "Service family"
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	//Create Users
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(user.name, user.email, user.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	//Login
	res = client.LoginUser(owner.email, owner.password)
	ownerLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.LoginUser(user.email, user.password)
	userLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	// Create the household
	res = client.CreateHousehold(ownerLoginResponse.Token, householdName)
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	//Invite the user to the household
	res = client.InviteUserToHousehold(ownerLoginResponse.Token,
		userLoginResponse.Id,
		householdCreationResponse.Id,
		nil)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected to receive status created, received: %d when creating the initial invite",
			res.StatusCode)
	}

	//Invite the user again to the household
	res = client.InviteUserToHousehold(ownerLoginResponse.Token,
		userLoginResponse.Id,
		householdCreationResponse.Id,
		nil)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected to receive status bad request, received: %d when sending the second invite",
			res.StatusCode)
	}

	res = client.GetInvites(userLoginResponse.Token)
	invites := helpers.GetResponseObject[[]dto.InviteResponse](t, res, http.StatusOK)

	if len(invites) != 1 {
		t.Fatalf("Expected to have only 1 invite, have %d", len(invites))
	}

	if invites[0].HouseholdId != householdCreationResponse.Id {
		t.Fatalf("Expected the invite to be for house ID %s, received %s",
			householdCreationResponse.Id.String(),
			invites[0].HouseholdId.String())
	}
}

// TODO: Add Direct invites Accept / Decline
func TestHandleHouseholdInvites(t *testing.T) {
	helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}
	secondOwner := UserInfo{
		name:     "mark",
		email:    "Mark@example.com",
		password: "very_secret",
	}
	nonMemberAccept := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}

	nonMemberDecline := UserInfo{
		name:     "joey",
		email:    "Joey@example.com",
		password: "fAcE-PaIn",
	}

	nonMemberAcceptDirectInv := UserInfo{
		name:     "shawn",
		email:    "shawn@example.com",
		password: "super-secret",
	}

	nonMemberDeclineDirectInv := UserInfo{
		name:     "Sha'myah",
		email:    "themtwo@example.com",
		password: "also-supersecret",
	}

	householdName := "Service family"
	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)

	// create owners
	res := client.CreateUser(owner.name, owner.email, owner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateUser(secondOwner.name, secondOwner.email, secondOwner.password)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status created, got: %d", res.StatusCode)
	}
	res.Body.Close()

	// login
	res = client.LoginUser(owner.email, owner.password)
	ownerLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.LoginUser(secondOwner.email, secondOwner.password)
	secondOwnerLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//create households
	res = client.CreateHousehold(ownerLoginResponse.Token, householdName)
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	res = client.CreateHousehold(secondOwnerLoginResponse.Token, "Ecivres family")
	secondHouseholdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	//create nonMembers and login
	res = client.CreateUser(nonMemberAccept.name, nonMemberAccept.email, nonMemberAccept.password)
	createUserAcceptResponse := helpers.GetResponseObject[dto.UserCreateResponse](t, res, http.StatusCreated)

	res = client.LoginUser(nonMemberAccept.email, nonMemberAccept.password)
	nonMemberAcceptLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.CreateUser(nonMemberDecline.name, nonMemberDecline.email, nonMemberDecline.password)
	createUserDeclineResponse := helpers.GetResponseObject[dto.UserCreateResponse](t, res, http.StatusCreated)

	res = client.LoginUser(nonMemberDecline.email, nonMemberDecline.password)
	nonMemberDeclineLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.CreateUser(nonMemberAcceptDirectInv.name, nonMemberAcceptDirectInv.email, nonMemberAcceptDirectInv.password)
	createUserAcceptDirResponse := helpers.GetResponseObject[dto.UserCreateResponse](t, res, http.StatusCreated)

	res = client.LoginUser(nonMemberAcceptDirectInv.email, nonMemberAcceptDirectInv.password)
	nonMemberAcceptDirLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	res = client.CreateUser(nonMemberDeclineDirectInv.name, nonMemberDeclineDirectInv.email, nonMemberDeclineDirectInv.password)
	createUserDeclineDirResponse := helpers.GetResponseObject[dto.UserCreateResponse](t, res, http.StatusCreated)

	res = client.LoginUser(nonMemberDeclineDirectInv.email, nonMemberDeclineDirectInv.password)
	nonMemberDeclineDirLoginResponse := helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)

	//create the new member and invite the non member
	res = client.CreateHouseholdMember(ownerLoginResponse.Token, "cassidy", householdCreationResponse.Id, &createUserAcceptResponse.Id)
	createMemberAcceptResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	res = client.CreateHouseholdMember(ownerLoginResponse.Token, "joey", householdCreationResponse.Id, &createUserDeclineResponse.Id)
	createMemberDeclineResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	//direct invites
	res = client.InviteUserToHousehold(ownerLoginResponse.Token, createUserAcceptDirResponse.Id, householdCreationResponse.Id, nil)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected invite user to household to return status created, received:%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.InviteUserToHousehold(ownerLoginResponse.Token, createUserDeclineDirResponse.Id, householdCreationResponse.Id, nil)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected invite user to household to return status created, received:%d", res.StatusCode)
	}
	res.Body.Close()

	//Create additional invites from different households
	res = client.CreateHouseholdMember(secondOwnerLoginResponse.Token, "CjIngram", secondHouseholdCreationResponse.Id, &createUserAcceptResponse.Id)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected the second household member / invite to cassidy to return status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	res = client.CreateHouseholdMember(secondOwnerLoginResponse.Token, "SecondLyphe", secondHouseholdCreationResponse.Id, &createUserDeclineResponse.Id)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected the second household member / invite to joey to return status created, got :%d", res.StatusCode)
	}
	res.Body.Close()

	testCases := []struct {
		testName    string
		userId      uuid.UUID
		token       string
		householdId uuid.UUID
		memberId    *uuid.UUID
		accept      bool
	}{
		{
			testName:    "Non Member Accepts Household Request",
			userId:      nonMemberAcceptLoginResponse.Id,
			token:       nonMemberAcceptLoginResponse.Token,
			householdId: householdCreationResponse.Id,
			memberId:    &createMemberAcceptResponse.Id,
			accept:      true,
		},
		{
			testName:    "Non Member Declines Household Request",
			userId:      nonMemberDeclineLoginResponse.Id,
			token:       nonMemberDeclineLoginResponse.Token,
			householdId: householdCreationResponse.Id,
			memberId:    &createMemberDeclineResponse.Id,
			accept:      false,
		},
		{
			testName:    "Non Member Accepts Household Request (Direct Invite)",
			userId:      nonMemberAcceptDirLoginResponse.Id,
			token:       nonMemberAcceptDirLoginResponse.Token,
			householdId: householdCreationResponse.Id,
			accept:      true,
		},
		{
			testName:    "Non Member Decliens HOusehold Request (Direct Invite)",
			userId:      nonMemberDeclineDirLoginResponse.Id,
			token:       nonMemberDeclineDirLoginResponse.Token,
			householdId: householdCreationResponse.Id,
			accept:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			//Check and make sure we have an invite from the household
			res := client.GetInvites(testCase.token)
			invites := helpers.GetResponseObject[[]dto.InviteResponse](t, res, http.StatusOK)
			inviteFound := false
			for _, invite := range invites {
				if invite.HouseholdId == testCase.householdId &&
					((invite.HouseholdMemberId != nil && testCase.memberId != nil && *invite.HouseholdMemberId == *testCase.memberId) ||
						(invite.HouseholdMemberId == nil && testCase.memberId == nil)) {
					inviteFound = true
				}
			}

			if !inviteFound {
				t.Fatal("Could not find an invite with the expected household and household member")
			}

			res = client.HandleInvite(testCase.token, testCase.householdId, testCase.accept)

			if testCase.accept && res.StatusCode != http.StatusNoContent {
				t.Fatalf("Expected to receive %v when accepting an invite, but received %v", http.StatusNoContent, res.StatusCode)
				res.Body.Close()
			}

			inviteRes := client.GetInvites(testCase.token)
			updatedInvites := helpers.GetResponseObject[[]dto.InviteResponse](t, inviteRes, http.StatusOK)

			inviteFound = false
			for _, invite := range updatedInvites {
				if invite.HouseholdId == testCase.householdId &&
					((invite.HouseholdMemberId != nil && testCase.memberId != nil && *invite.HouseholdMemberId == *testCase.memberId) ||
						(invite.HouseholdMemberId == nil && testCase.memberId == nil)) {
					inviteFound = true
				}
			}

			if inviteFound {
				t.Fatal("Found the invite with the expected household and household member after handling it")
			}

			if !testCase.accept {
				declinedInvite := helpers.GetResponseObject[dto.InviteResponse](t, res, http.StatusOK)

				if declinedInvite.HouseholdId != testCase.householdId {
					t.Errorf("Expected an household id of %v, got %v", testCase.householdId, declinedInvite.HouseholdId)
				}
				if testCase.memberId != nil {
					if *declinedInvite.HouseholdMemberId != *testCase.memberId {
						t.Errorf("Expected a household member id of %v, got %v", testCase.memberId, declinedInvite.HouseholdMemberId)
					}
				} else {
					if declinedInvite.HouseholdMemberId != nil {
						t.Error("Expected the household member id to be nil")
					}
				}
				if declinedInvite.InviterId != ownerLoginResponse.Id {
					t.Errorf("Expected an inviter id of %v, got %v", ownerLoginResponse.Id, declinedInvite.InviterId)
				}
				if declinedInvite.InviteeId != testCase.userId {
					t.Errorf("Expected an invitee id of %v, got %v", testCase.userId, declinedInvite.InviteeId)
				}
				if len(updatedInvites) != len(invites)-1 {
					t.Errorf("Expected there to be %d invites for the users. There are %d", len(invites)-1, len(updatedInvites))
				}

				if t.Failed() {
					t.FailNow()
				}
			} else {
				if len(updatedInvites) != 0 {
					t.Fatalf("Expected there to be no invites, but there are %d", len(updatedInvites))
				}
			}

			if testCase.memberId != nil {
				res = client.GetHouseholdMember(testCase.token, *testCase.memberId)
				memberInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)

				if testCase.accept && ((memberInfo.UserId != nil && *memberInfo.UserId != testCase.userId) || memberInfo.UserId == nil) {
					t.Fatalf("Expected the test case user (%v) to be the userId for the household member (%v)", testCase.userId, memberInfo.UserId)
				}
				if !testCase.accept && memberInfo.UserId != nil {
					t.Fatalf("Expected the household member to have a nil userId, got %v", memberInfo.UserId)
				}
			} else if testCase.accept {
				found := false
				res = client.GetHouseholdMembers(testCase.token, testCase.householdId)
				members := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)

				for _, member := range members {
					if member.UserId != nil && *member.UserId == testCase.userId {
						found = true
						break
					}
				}

				if !found {
					t.Fatal("For an accepted direct invite, expected a household member to have the test case user id linked")
				}
			}

		})
	}
}

func TestHandlerPromoteHouseholdMemberToHead(t *testing.T) {
	helpers.LoadDotEnv()

	users := map[string]struct {
		UserInfo
		dto.UserLoginResponse
		memberInfo dto.HouseholdMemberResponse
	}{
		"owner": {
			UserInfo: UserInfo{
				name:     "jonathan",
				email:    "Jon@example.com",
				password: "very-secret",
			},
		},
		"secondOwner": {
			UserInfo: UserInfo{
				name:     "mark",
				email:    "Mark@example.com",
				password: "very_secret",
			},
		},
		"member": {
			UserInfo: UserInfo{
				name:     "cassidy",
				email:    "Cass@example.com",
				password: "kalina",
			},
		},
		"thirdMember": {
			UserInfo: UserInfo{
				name:     "sean",
				email:    "Shawn@example.com",
				password: "otouto",
			},
		},
		"nonMember": {
			UserInfo: UserInfo{
				name:     "Tim",
				email:    "Jimothy@example.com",
				password: "oddball",
			},
		},
	}

	feastUrl := helpers.GetFeastURL()
	client := dto.NewClient(t, feastUrl)
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })

	// Create the users and login users
	for key, user := range users {
		res := client.CreateUser(user.name, user.email, user.password)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Expected the create user to return status created, but received %d", res.StatusCode)
		}
		res.Body.Close()

		res = client.LoginUser(user.email, user.password)
		user.UserLoginResponse = helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)
		users[key] = user
	}

	// Create the households
	res := client.CreateHousehold(users["owner"].Token, "Service Family")
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	res = client.CreateHousehold(users["secondOwner"].Token, "Ecivres Family")
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected the second household response code to be created, but received %d", res.StatusCode)
	}

	// Invite members
	memberInfo := users["member"]
	res = client.CreateHouseholdMember(users["owner"].Token, memberInfo.name, householdCreationResponse.Id, &memberInfo.Id)
	memberCreationResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	thirdMemberInfo := users["thirdMember"]
	res = client.CreateHouseholdMember(users["owner"].Token, thirdMemberInfo.name, householdCreationResponse.Id, &thirdMemberInfo.Id)
	secondMemberCreationResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	res = client.CreateHouseholdMember(users["owner"].Token, "Thrall", householdCreationResponse.Id, nil)
	nonuserMemberCreationResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	// Accept invites
	res = client.HandleInvite(users["member"].Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the handle invite to return no content, but received %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.HandleInvite(users["thirdMember"].Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the handle invite to return no content, but received %d", res.StatusCode)
	}
	res.Body.Close()

	// Update 1 member to admin
	res = client.UpdateHouseholdMember(users["owner"].Token, memberCreationResponse.Id, memberInfo.Id, memberCreationResponse.Name, 2)
	res.Body.Close()

	// Get the head of household
	res = client.GetHouseholdMembers(users["owner"].Token, householdCreationResponse.Id)
	members := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)
	var householdHead dto.HouseholdMemberResponse
	for _, member := range members {
		if member.Role == 1 {
			householdHead = member
			break
		}
	}

	testCases := []struct {
		testName, token       string
		householdId, memberId uuid.UUID
		statusCode            int
		fakeMember            bool
	}{
		{
			testName:    "Member other than the head tries to promote another member",
			token:       users["member"].Token,
			householdId: householdCreationResponse.Id,
			memberId:    secondMemberCreationResponse.Id,
			statusCode:  http.StatusForbidden,
		},
		{
			testName:    "Member other than the head tries to promote themself",
			token:       users["member"].Token,
			householdId: householdCreationResponse.Id,
			memberId:    memberCreationResponse.Id,
			statusCode:  http.StatusForbidden,
		},
		{
			testName:    "Member of another household tries to promote a member",
			token:       users["secondOwner"].Token,
			householdId: householdCreationResponse.Id,
			memberId:    memberCreationResponse.Id,
			statusCode:  http.StatusBadRequest,
		},
		{
			testName:    "non member tries to promote a member",
			token:       users["nonMember"].Token,
			householdId: householdCreationResponse.Id,
			memberId:    memberCreationResponse.Id,
			statusCode:  http.StatusNotFound,
		},
		{
			testName:    "The Head tries to promote a non-existent member",
			token:       users["owner"].Token,
			householdId: householdCreationResponse.Id,
			memberId:    uuid.New(),
			statusCode:  http.StatusNotFound,
			fakeMember:  true,
		},
		{
			testName:    "The head promotes themself",
			token:       users["owner"].Token,
			householdId: householdCreationResponse.Id,
			memberId:    householdHead.Id,
			statusCode:  http.StatusNoContent,
		},
		{
			testName:    "The head promotes a member that doesn't have a user",
			token:       users["owner"].Token,
			householdId: householdCreationResponse.Id,
			memberId:    nonuserMemberCreationResponse.Id,
			statusCode:  http.StatusBadRequest,
		},
		{
			testName:    "The head promotes a member",
			token:       users["owner"].Token,
			householdId: householdCreationResponse.Id,
			memberId:    memberCreationResponse.Id,
			statusCode:  http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			res := client.GetHouseholdMembers(testCase.token, testCase.householdId)
			members := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)
			var originalHeadId uuid.UUID
			for _, member := range members {
				if member.Role == 1 {
					originalHeadId = member.Id
					break
				}
			}
			res = client.PromoteHouseholdMemberToHead(testCase.token, testCase.memberId)
			if res.StatusCode != testCase.statusCode {
				t.Fatalf("Expected promote member to head to return %d, received %d",
					testCase.statusCode, res.StatusCode)
			}
			res.Body.Close()

			res = client.GetHouseholdMember(testCase.token, originalHeadId)
			originalHeadInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)

			if testCase.fakeMember {
				if originalHeadInfo.Role != 1 {
					t.Fatal("Expected the original head of household to remain the same")
				}
				return
			}
			res = client.GetHouseholdMember(testCase.token, testCase.memberId)
			member := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)

			if member.Id == originalHeadInfo.Id {
				if member.Role != 1 {
					t.Fatalf("Expected the household head to remain the same, is actually %d", member.Role)
				}
			} else if testCase.statusCode == http.StatusNoContent {
				if member.Role != 1 {
					t.Fatalf("Expected the member to be the household head, is actually %d", member.Role)
				}
				if originalHeadInfo.Role == 1 {
					t.Fatal("Expected the original head member to be demoted")
				}
			} else {
				if member.Role == 1 {
					t.Fatal("Expected the member to not be the household head")
				}

				if originalHeadInfo.Role != 1 {
					t.Fatal("Expected the original head of household to remain the same")
				}
			}
		})
	}
}

func TestLeaveHousehold(t *testing.T) {
	helpers.LoadDotEnv()

	users := map[string]struct {
		UserInfo
		dto.UserLoginResponse
		memberInfo dto.HouseholdMemberResponse
	}{
		"owner": {
			UserInfo: UserInfo{
				name:     "jonathan",
				email:    "Jon@example.com",
				password: "very-secret",
			},
		},
		"member": {
			UserInfo: UserInfo{
				name:     "cassidy",
				email:    "Cass@example.com",
				password: "kalina",
			},
		},
		"nonMember": {
			UserInfo: UserInfo{
				name:     "Tim",
				email:    "Jimothy@example.com",
				password: "oddball",
			},
		},
	}

	feastUrl := helpers.GetFeastURL()
	client := dto.NewClient(t, feastUrl)
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })

	// Create the users and login users
	for key, user := range users {
		res := client.CreateUser(user.name, user.email, user.password)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Expected the create user to return status created, but received %d", res.StatusCode)
		}
		res.Body.Close()

		res = client.LoginUser(user.email, user.password)
		user.UserLoginResponse = helpers.GetResponseObject[dto.UserLoginResponse](t, res, http.StatusOK)
		users[key] = user
	}

	// Create the households
	res := client.CreateHousehold(users["owner"].Token, "Service Family")
	householdCreationResponse := helpers.GetResponseObject[dto.HouseholdResponse](t, res, http.StatusCreated)

	// Invite the member
	memberInfo := users["member"]
	res = client.CreateHouseholdMember(users["owner"].Token, memberInfo.name, householdCreationResponse.Id, &memberInfo.Id)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected to get status created, received %d when creating the member", res.StatusCode)
	}
	res.Body.Close()

	// Accept invites
	res = client.HandleInvite(users["member"].Token, householdCreationResponse.Id, true)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected the handle invite to return no content, but received %d", res.StatusCode)
	}
	res.Body.Close()

	//test cases
	//Household head attempts to leave
	//Member attempts to leave
	//Non Member attempts to leave

	testCases := []struct {
		testName, token string
		householdId     *uuid.UUID
		userId          uuid.UUID
		statusCode      int
	}{
		{
			testName:    "Household head attempts to leave",
			token:       users["owner"].Token,
			householdId: &householdCreationResponse.Id,
			userId:      users["owner"].Id,
			statusCode:  http.StatusBadRequest,
		},
		{
			testName:    "Member attempts to leave household",
			token:       users["member"].Token,
			householdId: &householdCreationResponse.Id,
			userId:      users["member"].Id,
			statusCode:  http.StatusNoContent,
		},
		{
			testName:   "Non member attempts to leave household",
			token:      users["nonMember"].Token,
			userId:     users["nonMember"].Id,
			statusCode: http.StatusNoContent,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			var member dto.HouseholdMemberResponse
			if testCase.householdId != nil {
				res := client.GetHouseholdMembers(testCase.token, *testCase.householdId)
				members := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)
				for _, m := range members {
					if *m.UserId == testCase.userId {
						member = m
						break
					}
				}
			}

			res := client.LeaveHousehold(testCase.token)
			if res.StatusCode != testCase.statusCode {
				t.Fatalf("Expected the leave household endpoint to return %d, received %d",
					testCase.statusCode, res.StatusCode)
			}

			if testCase.householdId != nil {
				res = client.GetHouseholdMember(testCase.token, member.Id)
				updatedMember := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)
				if testCase.statusCode == http.StatusNoContent {
					if updatedMember.UserId != nil {
						t.Fatal("Expected the user to be removed from the member")
					}
				} else {
					if *updatedMember.UserId != testCase.userId {
						t.Fatal("Expected the user to remain connected to the member")
					}
				}
			}
		})
	}
}
