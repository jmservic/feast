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
			status:      http.StatusForbidden,
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
			responseCode: http.StatusOK,
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
			responseCode: http.StatusOK,
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
			responseCode: http.StatusOK,
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

			if testCase.responseCode == http.StatusOK {
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
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Error updating member's role, expected StatusOK, got %d", res.StatusCode)
	}
	res.Body.Close()

	res = client.UpdateHouseholdMember(users["owner"].loginResponse.Token, users["thirdMember"].memberInfo.Id,
		users["thirdMember"].loginResponse.Id, users["thirdMember"].memberInfo.Name, 2)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Error updating member's role, expected StatusOK, got %d", res.StatusCode)
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

//TODO: invite to household, duplicate invites to the same user

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

	//create the new member and invite the non member
	res = client.CreateHouseholdMember(ownerLoginResponse.Token, "cassidy", householdCreationResponse.Id, &createUserAcceptResponse.Id)
	createMemberAcceptResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

	res = client.CreateHouseholdMember(ownerLoginResponse.Token, "joey", householdCreationResponse.Id, &createUserDeclineResponse.Id)
	createMemberDeclineResponse := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusCreated)

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
		memberId    uuid.UUID
		accept      bool
	}{
		{
			testName:    "Non Member Accepts Household Request",
			userId:      nonMemberAcceptLoginResponse.Id,
			token:       nonMemberAcceptLoginResponse.Token,
			householdId: householdCreationResponse.Id,
			memberId:    createMemberAcceptResponse.Id,
			accept:      true,
		},
		{
			testName:    "Non Member Declines Household Request",
			userId:      nonMemberDeclineLoginResponse.Id,
			token:       nonMemberDeclineLoginResponse.Token,
			householdId: householdCreationResponse.Id,
			memberId:    createMemberDeclineResponse.Id,
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
				if invite.HouseholdId == testCase.householdId && invite.HouseholdMemberId != nil &&
					*invite.HouseholdMemberId == testCase.memberId {
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
				if invite.HouseholdId == testCase.householdId && invite.HouseholdMemberId != nil &&
					*invite.HouseholdMemberId == testCase.memberId {
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
				if *declinedInvite.HouseholdMemberId != testCase.memberId {
					t.Errorf("Expected a household member id of %v, got %v", testCase.memberId, declinedInvite.HouseholdMemberId)
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

			res = client.GetHouseholdMember(testCase.token, testCase.memberId)
			memberInfo := helpers.GetResponseObject[dto.HouseholdMemberResponse](t, res, http.StatusOK)

			if testCase.accept && ((memberInfo.UserId != nil && *memberInfo.UserId != testCase.userId) || memberInfo.UserId == nil) {
				t.Fatalf("Expected the test case user (%v) to be the userId for the household member (%v)", testCase.userId, memberInfo.UserId)
			}
			if !testCase.accept && memberInfo.UserId != nil {
				t.Fatalf("Expected the household member to have a nil userId, got %v", memberInfo.UserId)
			}

		})
	}
}

//TODO: giving the household to someone else.
