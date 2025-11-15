package user

import (
	"avito-test-pr-service/internal/utils"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type SetIsActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveResponse struct {
	User struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		TeamName string `json:"team_name"`
		IsActive bool   `json:"is_active"`
	} `json:"user"`
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "invalid json body")
		return
	}
	if err := utils.Validate(req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed")
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "invalid user_id")
		return
	}

	h.log.Info("SetIsActive request", slog.String("user_id", req.UserID), slog.Bool("is_active", req.IsActive))

	if err := h.userService.UpdateUserActive(r.Context(), userID, req.IsActive); err != nil {
		switch {
		case errors.Is(err, utils.ErrUserNotFound):
			_ = utils.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		case errors.Is(err, utils.ErrInvalidArgument):
			_ = utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		default:
			h.log.Error("SetIsActive service failed", slog.String("user_id", req.UserID), slog.Any("err", err))
			_ = utils.WriteError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
			return
		}
	}

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, utils.ErrUserNotFound) {
			_ = utils.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		h.log.Error("GetUser after SetIsActive failed", slog.String("user_id", req.UserID), slog.Any("err", err))
		_ = utils.WriteError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	teamName, err := h.userService.GetUserTeamName(r.Context(), userID)
	if err != nil {
		h.log.Error("GetUserTeamName failed", slog.String("user_id", req.UserID), slog.Any("err", err))
		_ = utils.WriteError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	var resp SetIsActiveResponse
	resp.User.UserID = user.ID.String()
	resp.User.Username = user.Name
	resp.User.TeamName = teamName
	resp.User.IsActive = user.IsActive

	_ = utils.WriteJSON(w, http.StatusOK, resp)
}
