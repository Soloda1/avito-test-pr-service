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
	req := GetTeamRequest{TeamName: r.URL.Query().Get("team_name")}
	if req.TeamName == "" {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "team_name is required")
		return
	}

	h.log.Info("GetTeam request", slog.String("team_name", req.TeamName))

	team, err := h.teamService.GetTeamByName(r.Context(), req.TeamName)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrInvalidArgument):
			_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "invalid team_name")
			return
		case errors.Is(err, utils.ErrTeamNotFound):
			_ = utils.WriteError(w, http.StatusNotFound, utils.HTTPStatusToCode(http.StatusNotFound), "team not found")
			return
		default:
			h.log.Error("GetTeam service failed", "err", err, "team_name", req.TeamName)
			_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPStatusToCode(http.StatusInternalServerError), "internal error")
			return
		}
	}

	users, err := h.userService.ListUsers(r.Context())
	if err != nil {
		h.log.Error("GetTeam list users failed", "err", err, "team_name", req.TeamName)
		_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPStatusToCode(http.StatusInternalServerError), "internal error")
		return
	}

	members := make([]GetTeamMember, 0)
	for _, u := range users {
		teamName, err := h.userService.GetUserTeamName(r.Context(), u.ID)
		if err != nil {
			if errors.Is(err, utils.ErrUserNoTeam) || errors.Is(err, utils.ErrTeamNotFound) {
				continue
			}
			h.log.Error("GetTeam get user team failed", "err", err, "user_id", u.ID)
			_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPStatusToCode(http.StatusInternalServerError), "internal error")
			return
		}
		if teamName == team.Name {
			members = append(members, GetTeamMember{
				UserID:   u.ID.String(),
				Username: u.Name,
				IsActive: u.IsActive,
			})
		}
	}

	resp := GetTeamResponse{TeamName: team.Name, Members: members}
	_ = utils.WriteJSON(w, http.StatusOK, resp)
}
