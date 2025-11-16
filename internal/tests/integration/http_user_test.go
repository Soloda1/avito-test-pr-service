package integration

import (
	"avito-test-pr-service/internal/application/pr"
	"avito-test-pr-service/internal/application/team"
	"avito-test-pr-service/internal/application/user"
	input "avito-test-pr-service/internal/domain/ports/input"
	"avito-test-pr-service/internal/infrastructure/config"
	apihttp "avito-test-pr-service/internal/infrastructure/http" // переименован, чтобы не конфликтовать с net/http
	"avito-test-pr-service/internal/infrastructure/logger"
	"avito-test-pr-service/internal/infrastructure/persistence/postgres/uow"
	"avito-test-pr-service/internal/infrastructure/reviewerselector"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func buildServices() (input.UserInputPort, input.PRInputPort, input.TeamInputPort) {
	log := logger.New("test")
	u := uow.NewPostgresUOW(pgC.Pool, log)
	userSvc := user.NewService(u, log)
	selector := reviewerselector.NewRandomReviewerSelector()
	prSvc := pr.NewService(u, selector, log)
	teamSvc := team.NewService(u, log)
	return userSvc, prSvc, teamSvc
}

func TestUserHandlers_HTTPIntegration(t *testing.T) {
	if pgC == nil {
		t.Fatal("postgres container not initialized")
	}

	userSvc, prSvc, teamSvc := buildServices()
	log := logger.New("test")
	r := apihttp.NewRouter(log, prSvc, teamSvc, userSvc)
	cfg := &config.Config{HTTPServer: config.HTTPServer{RequestTimeout: 5 * time.Second}}
	r.Setup(cfg)
	server := httptest.NewServer(r.GetRouter())
	defer server.Close()

	baseURL := server.URL

	t.Run("SetIsActive happy", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		if err := InsertUser(testCtx, pgC.Pool, "u1", "alice", false); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		body, _ := json.Marshal(map[string]any{"user_id": "u1", "is_active": true})
		resp, err := http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("http post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("resp.Body.Close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d", resp.StatusCode)
		}
		var r struct {
			User struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				TeamName string `json:"team_name"`
				IsActive bool   `json:"is_active"`
			} `json:"user"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if r.User.UserID != "u1" || !r.User.IsActive {
			b, _ := json.Marshal(r)
			t.Fatalf("unexpected body %s", string(b))
		}
	})

	t.Run("SetIsActive user not found", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		body, _ := json.Marshal(map[string]any{"user_id": "u404", "is_active": true})
		resp, err := http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("http post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("resp.Body.Close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("want 404 got %d", resp.StatusCode)
		}
	})

	t.Run("SetIsActive validation error (empty user_id)", func(t *testing.T) {
		body, _ := json.Marshal(map[string]any{"user_id": "", "is_active": true})
		resp, err := http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("http post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("resp.Body.Close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("want 400 got %d", resp.StatusCode)
		}
	})

	t.Run("GetReviews happy (two PRs)", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		if err := InsertUser(testCtx, pgC.Pool, "u1", "alice", true); err != nil {
			t.Fatalf("insert user: %v", err)
		}
		if err := InsertUser(testCtx, pgC.Pool, "u2", "bob", true); err != nil {
			t.Fatalf("insert user2: %v", err)
		}
		teamID, err := InsertTeam(testCtx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("insert team: %v", err)
		}
		if err := AddTeamMember(testCtx, pgC.Pool, teamID, "u1"); err != nil {
			t.Fatalf("add member u1: %v", err)
		}
		if err := AddTeamMember(testCtx, pgC.Pool, teamID, "u2"); err != nil {
			t.Fatalf("add member u2: %v", err)
		}
		if err := InsertPR(testCtx, pgC.Pool, "pr-1", "title1", "u1"); err != nil {
			t.Fatalf("insert pr1: %v", err)
		}
		if err := AddPRReviewer(testCtx, pgC.Pool, "pr-1", "u2"); err != nil {
			t.Fatalf("add reviewer pr1: %v", err)
		}
		if err := InsertPR(testCtx, pgC.Pool, "pr-2", "title2", "u1"); err != nil {
			t.Fatalf("insert pr2: %v", err)
		}
		if err := AddPRReviewer(testCtx, pgC.Pool, "pr-2", "u2"); err != nil {
			t.Fatalf("add reviewer pr2: %v", err)
		}
		resp, err := http.Get(baseURL + "/users/getReview?user_id=u2")
		if err != nil {
			t.Fatalf("http get: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("resp.Body.Close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d", resp.StatusCode)
		}
		var r struct {
			UserID       string `json:"user_id"`
			PullRequests []struct {
				ID     string `json:"pull_request_id"`
				Title  string `json:"pull_request_name"`
				Author string `json:"author_id"`
				Status string `json:"status"`
			} `json:"pull_requests"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if r.UserID != "u2" || len(r.PullRequests) != 2 {
			b, _ := json.Marshal(r)
			t.Fatalf("unexpected resp %s", string(b))
		}
	})

	t.Run("GetReviews empty user_id -> internal (service invalid arg)", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/users/getReview?user_id=")
		if err != nil {
			t.Fatalf("http get: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("resp.Body.Close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("want 500 got %d", resp.StatusCode)
		}
	})

	t.Run("GetReviews user with no PRs -> empty list", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		if err := InsertUser(testCtx, pgC.Pool, "u10", "lonely", true); err != nil {
			t.Fatalf("insert lonely: %v", err)
		}
		teamID, err := InsertTeam(testCtx, pgC.Pool, "core")
		if err != nil {
			t.Fatalf("insert team: %v", err)
		}
		if err := AddTeamMember(testCtx, pgC.Pool, teamID, "u10"); err != nil {
			t.Fatalf("add member: %v", err)
		}
		resp, err := http.Get(baseURL + "/users/getReview?user_id=u10")
		if err != nil {
			t.Fatalf("http get: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Fatalf("resp.Body.Close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d", resp.StatusCode)
		}
		var r struct {
			UserID       string `json:"user_id"`
			PullRequests []struct {
				ID string `json:"pull_request_id"`
			} `json:"pull_requests"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(r.PullRequests) != 0 {
			b, _ := json.Marshal(r)
			t.Fatalf("expected empty got %s", string(b))
		}
	})
}
