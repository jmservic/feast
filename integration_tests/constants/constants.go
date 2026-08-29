package constants

const (
	HouseholdsPath         = "/api/households"
	UsersPath              = "/api/users"
	LoginPath              = "/api/login"
	RefreshPath            = "/api/refresh"
	HouseholdMembersPath   = "/api/members"
	HouseholdIdMembersPath = HouseholdsPath + "/%v/members"
	InvitesPath            = HouseholdMembersPath + "/invites"
)
