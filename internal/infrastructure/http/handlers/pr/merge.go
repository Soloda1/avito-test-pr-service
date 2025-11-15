package pr

import (
	"avito-test-pr-service/internal/infrastructure/http/handlers/dto"
	"avito-test-pr-service/internal/utils"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

type MergePRResponse struct {
	PR dto.PRDTO `json:"pr"`
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), utils.ErrInvalidJSON.Error())
		return
	}
	if err := utils.Validate(req); err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), utils.ErrValidationFailed.Error())
		return
	}
	prID := req.PullRequestID

	h.log.Info("MergePR request", slog.String("pr_id", prID))

	pr, err := h.prService.MergePR(r.Context(), prID)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrPRNotFound):
			_ = utils.WriteError(w, http.StatusNotFound, utils.HTTPStatusToCode(http.StatusNotFound), utils.ErrPRNotFound.Error())
			return
		default:
			h.log.Error("MergePR failed", slog.Any("err", err), slog.String("pr_id", prID))
			_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPStatusToCode(http.StatusInternalServerError), utils.ErrInternal.Error())
			return
		}
	}
	_ = utils.WriteJSON(w, http.StatusOK, MergePRResponse{PR: dto.ToPRDTO(pr)})
}
