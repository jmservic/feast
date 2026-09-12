package main

import (
	"encoding/json"

	"github.com/google/uuid"

	//	"github.com/jmservic/feast/internal/auth"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jmservic/feast/internal/constants"
	"github.com/jmservic/feast/internal/database"
	"github.com/jmservic/feast/internal/dto"
	// "time"
)

func (cfg apiConfig) handlerCreateHouseholdMember(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	householdId, err := uuid.Parse(r.PathValue("household_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	params := struct {
		Name   string     `json:"name"`
		UserId *uuid.UUID `json:"user_id"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, err)
		return
	}
	// create a repeatable read transaction
	tx, err := cfg.conn.BeginTx(r.Context(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.BeginTransactionErrStr, err)
		return
	}
	defer tx.Rollback(r.Context())

	queries := cfg.db.WithTx(tx)

	currMembers, err := queries.GetHouseholdMembers(r.Context(), householdId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberRetrievalByIdErrStr, err)
		return
	}

	currMemberIds := make(map[string]struct{})
	for _, member := range currMembers {
		currMemberIds[member.ID.String()] = struct{}{}
	}

	err = queries.CreateHouseholdMember(r.Context(), database.CreateHouseholdMemberParams{
		CreatorID:   userId,
		MemberName:  params.Name,
		UserID:      params.UserId,
		HouseholdID: householdId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberCreateErrStr, err)
		return
	}

	updatedMembers, err := queries.GetHouseholdMembers(r.Context(), householdId)
	var newMember *database.HouseholdMember

	for i, member := range updatedMembers {
		if _, ok := currMemberIds[member.ID.String()]; ok {
			continue
		}
		newMember = &updatedMembers[i]
		break
	}

	if len(updatedMembers)-len(currMembers) != 1 {
		respondWithError(w, http.StatusInternalServerError, constants.HouseholdMemberCreateErrStr, err)
		return
	}

	if newMember == nil {
		respondWithError(w, http.StatusInternalServerError, constants.GetNewInstanceErrStr, err)
		return
	}

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberRetrievalByIdErrStr, err)
		return
	}

	err = tx.Commit(r.Context())
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberRetrievalByIdErrStr, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, dto.HouseholdMemberResources{
		Id:          newMember.ID,
		Name:        newMember.Name,
		CreatedAt:   newMember.CreatedAt,
		UpdatedAt:   newMember.UpdatedAt,
		Role:        newMember.Role,
		HouseholdId: newMember.HouseholdID,
		UserId:      newMember.UserID,
	})
}

func (cfg apiConfig) handlerInviteUserToHousehold(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	params := struct {
		UserId            uuid.UUID  `json:"user_id"`
		HouseholdId       uuid.UUID  `json:"household_id"`
		HouseholdMemberId *uuid.UUID `json:"household_member_id"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if params.UserId == uuid.Nil || params.HouseholdId == uuid.Nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, nil)
		return
	}

	err := cfg.db.InviteUserToHousehold(r.Context(), database.InviteUserToHouseholdParams{
		Inviter:            userId,
		Invitee:            params.UserId,
		VHouseholdID:       params.HouseholdId,
		VHouseholdMemberID: params.HouseholdMemberId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberInviteErrStr, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (cfg apiConfig) handlerUpdateHouseholdMember(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	memberId, err := uuid.Parse(r.PathValue("member_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}
	//can change name, role, and userid

	params := struct {
		Name   string    `json:"name"`
		UserId uuid.UUID `json:"user_id"`
		Role   int       `json:"role"`
	}{}

	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, constants.JsonDecodeErrStr, err)
		return
	}

	if params.Name == "" {
		respondWithError(w, http.StatusBadRequest, constants.EmptyParameterErrStr, err)
		return
	}

	err = cfg.db.UpdateHouseholdMember(r.Context(), database.UpdateHouseholdMemberParams{
		UpdaterID:         userId,
		NewName:           params.Name,
		NewRole:           params.Role,
		NewUserID:         params.UserId,
		HouseholdMemberID: memberId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberUpdateErrStr, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg apiConfig) handlerPromoteHouseholdMemberToHead(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	memberId, err := uuid.Parse(r.PathValue("member_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	err = cfg.db.PromoteHouseholdMemberToHead(r.Context(), database.PromoteHouseholdMemberToHeadParams{
		UserID:            userId,
		HouseholdMemberID: memberId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.PromoteHouseholdMemberToHeadErrStr, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg apiConfig) handlerGetHouseholdMember(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	memberId, err := uuid.Parse(r.PathValue("member_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	memberInfo, err := cfg.db.GetHouseholdMember(r.Context(), memberId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberRetrievalByIdErrStr, err)
		return
	}
	respondWithJSON(w, http.StatusOK, dto.HouseholdMemberResources{
		Id:          memberInfo.ID,
		Name:        memberInfo.Name,
		CreatedAt:   memberInfo.CreatedAt,
		UpdatedAt:   memberInfo.UpdatedAt,
		Role:        memberInfo.Role,
		HouseholdId: memberInfo.HouseholdID,
		UserId:      memberInfo.UserID,
	})
}

func (cfg apiConfig) handlerGetHouseholdMembers(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	householdId, err := uuid.Parse(r.PathValue("household_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	householdMembers, err := cfg.db.GetHouseholdMembers(r.Context(), householdId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMembersRetrievalByHouseholdIdErrStr, err)
		return
	}

	var rtnMembers []dto.HouseholdMemberResources
	for _, memberInfo := range householdMembers {
		rtnMembers = append(rtnMembers, dto.HouseholdMemberResources{
			Id:          memberInfo.ID,
			Name:        memberInfo.Name,
			CreatedAt:   memberInfo.CreatedAt,
			UpdatedAt:   memberInfo.UpdatedAt,
			Role:        memberInfo.Role,
			HouseholdId: memberInfo.HouseholdID,
			UserId:      memberInfo.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, rtnMembers)
}

func (cfg apiConfig) handlerDeleteHouseholdMember(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	memberId, err := uuid.Parse(r.PathValue("member_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	err = cfg.db.DeleteHouseholdMember(r.Context(), database.DeleteHouseholdMemberParams{
		VUserID:           userId,
		HouseholdMemberID: memberId,
	})

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.HouseholdMemberDeleteErrStr, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (cfg apiConfig) handlerGetMemberInvites(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	// This will return invites where you are the invitee or the inviter
	invites, err := cfg.db.GetHouseholdInvites(r.Context(), userId)
	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.MemberInviteRetrievalByUserErrStr, err)
		return
	}

	var rtnInvites []dto.HouseholdInviteResources
	for _, inviteInfo := range invites {
		rtnInvites = append(rtnInvites, dto.HouseholdInviteResources{
			InviterId:         inviteInfo.InviterID,
			InviteeId:         inviteInfo.InviteeID,
			HouseholdMemberId: inviteInfo.HouseholdMemberID,
			HouseholdId:       inviteInfo.HouseholdID,
			CreatedAt:         inviteInfo.CreatedAt,
		})
	}

	respondWithJSON(w, http.StatusOK, rtnInvites)
}

func (cfg apiConfig) handlerHandleMemberInvite(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	householdId, err := uuid.Parse(r.FormValue("household_id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, constants.InvalidUUIDErrStr, err)
		return
	}

	action := strings.ToLower(r.FormValue("action"))
	switch action {
	case "accept":
		err = cfg.db.AcceptHouseholdInvite(r.Context(), database.AcceptHouseholdInviteParams{
			InviteeID:   userId,
			HouseholdID: householdId,
		})
		if err != nil {
			respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.InviteAcceptErrStr, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	case "decline":
		declinedInvite, err := cfg.db.DeclineHouseholdInvite(r.Context(), database.DeclineHouseholdInviteParams{
			InviteeID:   userId,
			HouseholdID: householdId,
		})
		if err != nil {
			respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.InviteDeclineErrStr, err)
			return
		}

		// return the declinedInvite
		respondWithJSON(w, http.StatusOK, dto.HouseholdInviteResources{
			InviterId:         declinedInvite.InviterID,
			InviteeId:         declinedInvite.InviteeID,
			HouseholdMemberId: declinedInvite.HouseholdMemberID,
			HouseholdId:       declinedInvite.HouseholdID,
			CreatedAt:         declinedInvite.CreatedAt,
		})
	default:
		respondWithError(w, http.StatusBadRequest, constants.HandleMemberInviteActionErrStr, nil)
		return
	}
}

func (cfg apiConfig) handlerLeaveHousehold(w http.ResponseWriter, r *http.Request, userId uuid.UUID) {
	err := cfg.db.LeaveHousehold(r.Context(), userId)

	if err != nil {
		respondWithError(w, mapDbErrorToHttpStatusCode(err), constants.LeaveHouseholdErrStr, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
