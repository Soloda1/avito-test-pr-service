package utils

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

func HTTPCodeConverter(status int, errs ...error) string {
	if status == http.StatusConflict && len(errs) > 0 && errs[0] != nil {
		err := errs[0]
		switch {
		case errors.Is(err, ErrAlreadyMerged):
			return "PR_MERGED"
		case errors.Is(err, ErrReviewerNotAssigned):
			return "NOT_ASSIGNED"
		case errors.Is(err, ErrNoReplacementCandidates):
			return "NO_CANDIDATE"
		case errors.Is(err, ErrPRExists):
			return "PR_EXISTS"
		case errors.Is(err, ErrTeamExists):
			return "TEAM_EXISTS"
		}
	}
	switch status {
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	default:
		return "INTERNAL"
	}
}

func WriteJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorResponse{Error: ErrorDetails{Code: code, Message: message}}
	return json.NewEncoder(w).Encode(resp)
}
