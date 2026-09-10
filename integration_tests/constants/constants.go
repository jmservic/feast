package constants

const (
	HouseholdsPath              = "/api/households"
	UsersPath                   = "/api/users"
	LeaveHouseholdPath          = HouseholdMembersPath + "/leave"
	LoginPath                   = "/api/login"
	RefreshPath                 = "/api/refresh"
	HouseholdMembersPath        = "/api/members"
	HouseholdIdMembersPath      = HouseholdsPath + "/%v/members"
	HouseholdMembersPromotePath = HouseholdMembersPath + "/%v/promote"
	InvitesPath                 = HouseholdMembersPath + "/invites"
)
