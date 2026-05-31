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
	nonMember := UserInfo{
		name:     "cassidy",
		email:    "Cass@example.com",
		password: "kalina",
	}

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

	res = client.CreateHouseholdMember(loginResponse.Token, "cassidy", householdCreationResponse.Id, nil)
	memberCreationResponse := helpers.GetResponseObject[dto.MemberCreateResponse(t, res, http.StatusCreated)
	t.FailNow()
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
	t.FailNow()
}

func TestHandleHouseholdInvites(t *testing.T) {
	t.FailNow()
}
