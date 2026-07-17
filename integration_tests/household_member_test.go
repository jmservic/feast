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
	res.Body.Close()

	testCases := []struct {
		testName, memberName string
		userId               *uuid.UUID
		status               int
		memberDiff           int
	}{
		{
			testName:   "Create User without an User Id",
			memberName: "Johnny Jr.",
			userId:     nil,
			status:     http.StatusCreated,
			memberDiff: 1,
		},
		{
			testName:   "Create User with an User Id",
			memberName: "Cassidy",
			userId:     &createUserResponse.Id,
			status:     http.StatusCreated,
			memberDiff: 1,
		},
		{
			testName:   "Create User with an User Id in another household",
			memberName: "Joey",
			userId:     &createOtherOwnerResponse.Id,
			status:     http.StatusInternalServerError,
			memberDiff: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			// Check members before
			res = client.GetHouseholdMembers(loginResponse.Token, householdCreationResponse.Id)
			originalMembers := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)

			// act
			res = client.CreateHouseholdMember(loginResponse.Token, testCase.memberName,
				householdCreationResponse.Id, testCase.userId)
			res.Body.Close()

			// assert
			if res.StatusCode != testCase.status {
				t.Fatalf("Expected status created for member creation, got: %d.", res.StatusCode)
			}

			//Check members after
			res = client.GetHouseholdMembers(loginResponse.Token, householdCreationResponse.Id)
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
					if invite.HouseholdId == householdCreationResponse.Id &&
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
	/*helpers.LoadDotEnv()
	owner := UserInfo{
		name:     "jonathan",
		email:    "Jon@example.com",
		password: "very-secret",
	}
	Member := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}
	//managerMember
	//nonmember (not in a household)
	//nonmember (in another household)

	householdName := "Service family"

	feastUrl := helpers.GetFeastURL()
	t.Cleanup(func() { helpers.ResetDatabase(feastUrl) })
	client := dto.NewClient(t, feastUrl)
	*/

	//cases, update name, user id, role.. actually kind of need to do the other ones before I complete this one..
	// I need users at different role levels, so I can test the failing cases as well.
	t.FailNow()

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

	res = client.GetHouseholdMember(loginResponse.Token, householdCreationResponse.Id, memberCreationResponse.Id)
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
	t.FailNow()
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

			res = client.GetHouseholdMember(testCase.token, testCase.householdId, testCase.memberId)
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
