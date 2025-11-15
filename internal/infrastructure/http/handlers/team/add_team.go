package team

import (
	"avito-test-pr-service/internal/domain/models"
	"avito-test-pr-service/internal/utils"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type AddTeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type AddTeamRequest struct {
	TeamName string          `json:"team_name"`
	Members  []AddTeamMember `json:"members"`
}

type AddTeamResponse struct {
	Team struct {
		TeamName string          `json:"team_name"`
		Members  []AddTeamMember `json:"members"`
	} `json:"team"`
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req AddTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "invalid json body")
		return
	}
	if req.TeamName == "" {
		_ = utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "team_name is required")
		return
	}

	h.log.Info("AddTeam request", slog.String("team_name", req.TeamName))

	var usersIn []*models.User
	for _, m := range req.Members {
		var id uuid.UUID
		if m.UserID != "" {
			parsed, err := uuid.Parse(m.UserID)
			if err != nil {
				_ = utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid user_id")
				return
			}
			id = parsed
		}
		usersIn = append(usersIn, &models.User{ID: id, Name: m.Username, IsActive: m.IsActive})
	}

	team, users, err := h.teamService.CreateTeamWithMembers(r.Context(), req.TeamName, usersIn)
	if err != nil {
		_ = utils.WriteError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	var resp AddTeamResponse
	resp.Team.TeamName = team.Name
	for _, u := range users {
		resp.Team.Members = append(resp.Team.Members, AddTeamMember{
			UserID:   u.ID.String(),
			Username: u.Name,
			IsActive: u.IsActive,
		})
	}
	_ = utils.WriteJSON(w, http.StatusCreated, resp)
}
