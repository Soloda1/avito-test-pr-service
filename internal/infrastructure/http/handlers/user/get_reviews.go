package user

import (
	"avito-test-pr-service/internal/utils"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type GetReviewsRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

type GetReviewsResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

type PullRequestShort struct {
	ID     string `json:"pull_request_id"`
	Title  string `json:"pull_request_name"`
	Author string `json:"author_id"`
	Status string `json:"status"`
}

func (h *UserHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	req := GetReviewsRequest{UserID: r.URL.Query().Get("user_id")}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		_ = utils.WriteError(w, http.StatusBadRequest, utils.HTTPStatusToCode(http.StatusBadRequest), err.Error())
		return
	}

	h.log.Info("GetReviews request", slog.String("user_id", req.UserID))

	prs, err := h.prService.ListPRsByAssignee(r.Context(), userID, nil)
	if err != nil {
		h.log.Error("GetReviews service failed", slog.String("user_id", req.UserID), slog.Any("err", err))
		_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPStatusToCode(http.StatusInternalServerError), err.Error())
		return
	}

	resp := GetReviewsResponse{UserID: req.UserID, PullRequests: []PullRequestShort{}}
	for _, p := range prs {
		resp.PullRequests = append(resp.PullRequests, PullRequestShort{
			ID:     p.ID,
			Title:  p.Title,
			Author: p.AuthorID.String(),
			Status: string(p.Status),
		})
	}

	_ = utils.WriteJSON(w, http.StatusOK, resp)
}
