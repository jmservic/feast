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

	testCases := []struct {
		testName, memberName string
		userId               *uuid.UUID
	}{
		{
			testName:   "Create User without a User Id",
			memberName: "Johnny Jr.",
			userId:     nil,
		},
		{
			testName:   "Create User with a User Id",
			memberName: "Cassidy",
			userId:     &createUserResponse.Id,
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

			// assert
			if res.StatusCode != http.StatusCreated {
				t.Fatalf("Expected status created for member creation, got: %d", res.StatusCode)
			}

			//Check members after
			res = client.GetHouseholdMembers(loginResponse.Token, householdCreationResponse.Id)
			updatedMembers := helpers.GetResponseObject[[]dto.HouseholdMemberResponse](t, res, http.StatusOK)

			if len(updatedMembers) != len(originalMembers)+1 {
				t.Fatalf("Expected length of updated members list to be %d, but got %d",
					len(originalMembers)+1, len(updatedMembers))
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
	t.FailNow()
}

func TestGetHouseholdMember(t *testing.T) {
	t.FailNow()
}

func TestGetHouseholdMembers(t *testing.T) {
	t.FailNow()
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
