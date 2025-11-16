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

func buildServices(t *testing.T) (input.UserInputPort, input.PRInputPort, input.TeamInputPort) {
	log := logger.New("test")
	u := uow.NewPostgresUOW(pgC.Pool, log)
	userSvc := user.NewService(u, log)
	selector := reviewerselector.NewRandomReviewerSelector()
	prSvc := pr.NewService(u, selector, log)
	teamSvc := team.NewService(u, log)
	return userSvc, prSvc, teamSvc
}

func prepareUser(t *testing.T, id, name string, active bool) {
	_, err := pgC.Pool.Exec(testCtx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, id, name, active)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
}

func prepareTeamWithMembers(t *testing.T, name string, userIDs ...string) string {
	row := pgC.Pool.QueryRow(testCtx, `INSERT INTO teams(id, name, created_at, updated_at) VALUES (gen_random_uuid(),$1,now(),now()) RETURNING id`, name)
	var id string
	if err := row.Scan(&id); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	for _, uid := range userIDs {
		_, err := pgC.Pool.Exec(testCtx, `INSERT INTO team_members(team_id, user_id) VALUES ($1,$2)`, id, uid)
		if err != nil {
			t.Fatalf("insert team member: %v", err)
		}
	}
	return id
}

func createPR(t *testing.T, prID, authorID, title string, reviewerIDs ...string) {
	_, err := pgC.Pool.Exec(testCtx, `INSERT INTO prs(id, title, author_id, status, created_at, updated_at) VALUES ($1,$2,$3,'OPEN',now(),now())`, prID, title, authorID)
	if err != nil {
		t.Fatalf("insert pr: %v", err)
	}
	for _, rid := range reviewerIDs {
		_, err := pgC.Pool.Exec(testCtx, `INSERT INTO pr_reviewers(pr_id, reviewer_id, assigned_at) VALUES ($1,$2,now())`, prID, rid)
		if err != nil {
			t.Fatalf("insert reviewer: %v", err)
		}
	}
}

func TestUserHandlers_HTTPIntegration(t *testing.T) {
	if pgC == nil {
		t.Fatal("postgres container not initialized")
	}

	userSvc, prSvc, teamSvc := buildServices(t)
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
		prepareUser(t, "u1", "alice", false)
		body, _ := json.Marshal(map[string]any{"user_id": "u1", "is_active": true})
		resp, err := http.Post(baseURL+"/users/setIsActive", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("http post: %v", err)
		}
		defer resp.Body.Close()
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
		defer resp.Body.Close()
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
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("want 400 got %d", resp.StatusCode)
		}
	})

	t.Run("GetReviews happy (two PRs)", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		prepareUser(t, "u1", "alice", true)
		prepareUser(t, "u2", "bob", true)
		prepareTeamWithMembers(t, "core", "u1", "u2")
		createPR(t, "pr-1", "u1", "title1", "u2")
		createPR(t, "pr-2", "u1", "title2", "u2")
		resp, err := http.Get(baseURL + "/users/getReview?user_id=u2")
		if err != nil {
			t.Fatalf("http get: %v", err)
		}
		defer resp.Body.Close()
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
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("want 500 got %d", resp.StatusCode)
		}
	})

	t.Run("GetReviews user with no PRs -> empty list", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		prepareUser(t, "u10", "lonely", true)
		prepareTeamWithMembers(t, "core", "u10")
		resp, err := http.Get(baseURL + "/users/getReview?user_id=u10")
		if err != nil {
			t.Fatalf("http get: %v", err)
		}
		defer resp.Body.Close()
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
