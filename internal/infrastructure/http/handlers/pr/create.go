package pr

import (
	"avito-test-pr-service/internal/infrastructure/http/handlers/dto"
	"avito-test-pr-service/internal/utils"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
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
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "invalid json body")
		return
	}
	if err := utils.Validate(req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "validation failed")
		return
	}
	prID, err := uuid.Parse(req.PullRequestID)
	if err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid pull_request_id")
		return
	}
	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid author_id")
		return
	}

	h.log.Info("CreatePR request", slog.String("pr_id", prID.String()), slog.String("author_id", authorID.String()))

	pr, err := h.prService.CreatePR(r.Context(), authorID, req.PullRequestName)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrPRExists):
			_ = utils.WriteError(w, http.StatusConflict, "PR_EXISTS", "PR id already exists")
			return
		case errors.Is(err, utils.ErrUserNotFound), errors.Is(err, utils.ErrTeamNotFound):
			_ = utils.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		default:
			h.log.Error("CreatePR failed", slog.Any("err", err), slog.String("author_id", authorID.String()))
			_ = utils.WriteError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
			return
		}
	}

	resp := PRResponse{PR: dto.ToPRDTO(pr)}
	_ = utils.WriteJSON(w, http.StatusCreated, resp)
}
