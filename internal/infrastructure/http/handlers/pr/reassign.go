package pr

import (
	"avito-test-pr-service/internal/infrastructure/http/handlers/dto"
	"avito-test-pr-service/internal/utils"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
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
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), utils.ErrInvalidJSON.Error())
		return
	}
	if err := utils.Validate(req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), utils.ErrValidationFailed.Error())
		return
	}
	prID := req.PullRequestID
	oldID := req.OldUserID

	if prID == "" || oldID == "" {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), utils.ErrInvalidArgument.Error())
		return
	}

	h.log.Info("Reassign request", slog.String("pr_id", prID), slog.String("old_user_id", oldID))

	pr, err := h.prService.ReassignReviewer(r.Context(), prID, oldID)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrPRNotFound) || errors.Is(err, utils.ErrUserNotFound):
			_ = utils.WriteError(w, http.StatusNotFound, utils.HTTPStatusToCode(http.StatusNotFound), utils.ErrNotFound.Error())
			return
		case errors.Is(err, utils.ErrAlreadyMerged) || errors.Is(err, utils.ErrReviewerNotAssigned) || errors.Is(err, utils.ErrNoReplacementCandidates):
			status := http.StatusConflict
			codeStr := utils.HTTPStatusToCode(status, err)
			_ = utils.WriteError(w, status, codeStr, err.Error())
			return
		default:
			h.log.Error("Reassign failed", slog.Any("err", err), slog.String("pr_id", prID))
			_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPStatusToCode(http.StatusInternalServerError), utils.ErrInternal.Error())
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
