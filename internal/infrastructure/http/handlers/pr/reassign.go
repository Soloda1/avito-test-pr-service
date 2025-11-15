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

type ReassignPRRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldUserID     string `json:"old_user_id" validate:"required"`
}

type ReassignPRResponse struct {
	PR         dto.PRDTO `json:"pr"`
	ReplacedBy string    `json:"replaced_by"`
}

func (h *PRHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req ReassignPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "invalid json body")
		return
	}
	if err := utils.Validate(req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "validation failed")
		return
	}
	prID := req.PullRequestID
	if prID == "" {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "pull_request_id is required")
		return
	}
	oldID, err := uuid.Parse(req.OldUserID)
	if err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), "invalid old_user_id")
		return
	}

	h.log.Info("Reassign request", slog.String("pr_id", prID), slog.String("old_user_id", oldID.String()))

	pr, err := h.prService.ReassignReviewer(r.Context(), prID, oldID)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrPRNotFound) || errors.Is(err, utils.ErrUserNotFound):
			_ = utils.WriteError(w, http.StatusNotFound, utils.HTTPStatusToCode(http.StatusNotFound), err.Error())
			return
		case errors.Is(err, utils.ErrAlreadyMerged), errors.Is(err, utils.ErrReviewerNotAssigned), errors.Is(err, utils.ErrNoReplacementCandidates):
			code := http.StatusConflict
			if errors.Is(err, utils.ErrAlreadyMerged) {
				_ = utils.WriteError(w, code, utils.HTTPStatusToCode(code), err.Error())
				return
			}
			if errors.Is(err, utils.ErrReviewerNotAssigned) {
				_ = utils.WriteError(w, code, utils.HTTPStatusToCode(code), err.Error())
				return
			}
			if errors.Is(err, utils.ErrNoReplacementCandidates) {
				_ = utils.WriteError(w, code, utils.HTTPStatusToCode(code), err.Error())
				return
			}
			return
		default:
			h.log.Error("Reassign failed", slog.Any("err", err), slog.String("pr_id", prID))
			_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPStatusToCode(http.StatusInternalServerError), "internal error")
			return
		}
	}

	body := dto.ToPRDTO(pr)
	replacedBy := ""
	if len(body.AssignedReviewers) > 0 {
		replacedBy = body.AssignedReviewers[len(body.AssignedReviewers)-1]
	}
	_ = utils.WriteJSON(w, http.StatusOK, ReassignPRResponse{PR: body, ReplacedBy: replacedBy})
}
