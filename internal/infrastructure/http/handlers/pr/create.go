package pr

import (
	"avito-test-pr-service/internal/infrastructure/http/handlers/dto"
	"avito-test-pr-service/internal/utils"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
}

type PRResponse struct {
	PR dto.PRDTO `json:"pr"`
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPCodeConverter(http.StatusBadRequest), utils.ErrInvalidJSON.Error())
		return
	}
	if err := utils.Validate(req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPCodeConverter(http.StatusBadRequest), err.Error())
		return
	}
	prID := req.PullRequestID
	authorID := req.AuthorID

	if prID == "" || authorID == "" {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPCodeConverter(http.StatusBadRequest), utils.ErrInvalidArgument.Error())
		return
	}

	h.log.Info("CreatePR request", slog.String("pr_id", prID), slog.String("author_id", authorID))

	pr, err := h.prService.CreatePR(r.Context(), prID, authorID, req.PullRequestName)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrPRExists):
			_ = utils.WriteError(w, http.StatusConflict, utils.HTTPCodeConverter(http.StatusConflict, err), err.Error())
			return
		case errors.Is(err, utils.ErrUserNotFound) || errors.Is(err, utils.ErrTeamNotFound):
			_ = utils.WriteError(w, http.StatusNotFound, utils.HTTPCodeConverter(http.StatusNotFound), err.Error())
			return
		default:
			h.log.Error("CreatePR failed", slog.Any("err", err), slog.String("author_id", authorID))
			_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPCodeConverter(http.StatusInternalServerError), utils.ErrInternal.Error())
			return
		}
	}

	resp := PRResponse{PR: dto.ToPRDTO(pr)}
	_ = utils.WriteJSON(w, http.StatusCreated, resp)
}
