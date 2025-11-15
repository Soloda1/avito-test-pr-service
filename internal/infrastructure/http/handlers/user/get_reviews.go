package user

import (
	"avito-test-pr-service/internal/utils"
	"log/slog"
	"net/http"
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
	userID := r.URL.Query().Get("user_id")

	h.log.Info("GetReviews request", slog.String("user_id", userID))

	prs, err := h.prService.ListPRsByAssignee(r.Context(), userID, nil)
	if err != nil {
		h.log.Error("GetReviews service failed", slog.String("user_id", userID), slog.Any("err", err))
		_ = utils.WriteError(w, http.StatusInternalServerError, utils.HTTPCodeConverter(http.StatusInternalServerError), utils.ErrInternal.Error())
		return
	}

	resp := GetReviewsResponse{UserID: userID, PullRequests: []PullRequestShort{}}
	for _, p := range prs {
		resp.PullRequests = append(resp.PullRequests, PullRequestShort{
			ID:     p.ID,
			Title:  p.Title,
			Author: p.AuthorID,
			Status: string(p.Status),
		})
	}

	_ = utils.WriteJSON(w, http.StatusOK, resp)
}
