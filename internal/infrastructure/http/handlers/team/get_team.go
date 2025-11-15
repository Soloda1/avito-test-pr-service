package team

import (
	"avito-test-pr-service/internal/utils"
	"errors"
	"log/slog"
	"net/http"
)

type GetTeamRequest struct {
	TeamName string `json:"team_name" validate:"required"`
}

type GetTeamResponse struct {
	TeamName string          `json:"team_name"`
	Members  []GetTeamMember `json:"members"`
}

type GetTeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPCodeConverter(http.StatusBadRequest), utils.ErrInvalidArgument.Error())
		return
	}

	h.log.Info("GetTeam request", slog.String("team_name", teamName))

	team, err := h.teamService.GetTeamByName(r.Context(), teamName)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrInvalidArgument):
			_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPCodeConverter(http.StatusBadRequest), err.Error())
			return
		case errors.Is(err, utils.ErrTeamNotFound):
			_ = utils.WriteError(w, http.StatusNotFound, utils.HTTPCodeConverter(http.StatusNotFound), err.Error())
			return
		default:
			h.log.Error("GetTeam service failed", slog.Any("err", err), slog.String("team_name", teamName))
			_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPCodeConverter(http.StatusInternalServerError), utils.ErrInternal.Error())
			return
		}
	}

	membersUsers, err := h.userService.ListMembersByTeamID(r.Context(), team.ID.String())
	if err != nil {
		h.log.Error("GetTeam list members by team failed", slog.Any("err", err), slog.String("team_id", team.ID.String()))
		_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPCodeConverter(http.StatusInternalServerError), utils.ErrInternal.Error())
		return
	}

	members := make([]GetTeamMember, 0, len(membersUsers))
	for _, u := range membersUsers {
		members = append(members, GetTeamMember{UserID: u.ID, Username: u.Name, IsActive: u.IsActive})
	}

	resp := GetTeamResponse{TeamName: team.Name, Members: members}
	_ = utils.WriteJSON(w, http.StatusOK, resp)
}
