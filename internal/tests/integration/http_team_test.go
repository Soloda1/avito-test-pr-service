package integration

import (
	"avito-test-pr-service/internal/application/pr"
	"avito-test-pr-service/internal/application/team"
	"avito-test-pr-service/internal/application/user"
	input "avito-test-pr-service/internal/domain/ports/input"
	"avito-test-pr-service/internal/infrastructure/config"
	apihttp "avito-test-pr-service/internal/infrastructure/http"
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

func buildTeamDeps(t *testing.T) (input.TeamInputPort, input.UserInputPort, input.PRInputPort) {
	log := logger.New("test")
	u := uow.NewPostgresUOW(pgC.Pool, log)
	teamSvc := team.NewService(u, log)
	userSvc := user.NewService(u, log)
	selector := reviewerselector.NewRandomReviewerSelector()
	prSvc := pr.NewService(u, selector, log)
	return teamSvc, userSvc, prSvc
}

func TestTeamHandlers_HTTPIntegration(t *testing.T) {
	if pgC == nil {
		t.Fatal("postgres not init")
	}

	teamSvc, userSvc, prSvc := buildTeamDeps(t)
	log := logger.New("test")
	r := apihttp.NewRouter(log, prSvc, teamSvc, userSvc)
	cfg := &config.Config{HTTPServer: config.HTTPServer{RequestTimeout: 5 * time.Second}}
	r.Setup(cfg)
	server := httptest.NewServer(r.GetRouter())
	defer server.Close()
	baseURL := server.URL

	postJSON := func(path string, body any) (*http.Response, error) {
		b, _ := json.Marshal(body)
		return http.Post(baseURL+path, "application/json", bytes.NewReader(b))
	}

	t.Run("AddTeam happy", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		req := map[string]any{
			"team_name": "core",
			"members":   []map[string]any{{"user_id": "u1", "username": "alice", "is_active": true}, {"user_id": "u2", "username": "bob", "is_active": false}},
		}
		resp, err := postJSON("/team/add", req)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("want 201 got %d", resp.StatusCode)
		}
		var r struct {
			Team struct {
				TeamName string `json:"team_name"`
				Members  []struct {
					UserID   string `json:"user_id"`
					Username string `json:"username"`
					IsActive bool   `json:"is_active"`
				} `json:"members"`
			} `json:"team"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if r.Team.TeamName != "core" || len(r.Team.Members) != 2 {
			t.Fatalf("unexpected resp %+v", r)
		}
	})

	t.Run("AddTeam duplicate -> 409", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		req := map[string]any{"team_name": "core", "members": []map[string]any{{"user_id": "u1", "username": "alice"}}}
		resp1, err := postJSON("/team/add", req)
		if err != nil {
			t.Fatalf("post1: %v", err)
		}
		resp1.Body.Close()
		resp2, err := postJSON("/team/add", req)
		if err != nil {
			t.Fatalf("post2: %v", err)
		}
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusConflict {
			t.Fatalf("want 409 got %d", resp2.StatusCode)
		}
	})

	t.Run("AddTeam validation error (missing team_name)", func(t *testing.T) {
		req := map[string]any{"members": []map[string]any{{"user_id": "u1", "username": "alice"}}}
		resp, err := postJSON("/team/add", req)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("want 400 got %d", resp.StatusCode)
		}
	})

	t.Run("GetTeam not found", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		resp, err := http.Get(baseURL + "/team/get?team_name=absent")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("want 404 got %d", resp.StatusCode)
		}
	})

	t.Run("GetTeam happy with members", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}

		req := map[string]any{"team_name": "core", "members": []map[string]any{{"user_id": "u1", "username": "alice", "is_active": true}, {"user_id": "u2", "username": "bob", "is_active": true}}}
		resp, err := postJSON("/team/add", req)
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		resp.Body.Close()

		getResp, err := http.Get(baseURL + "/team/get?team_name=core")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		defer getResp.Body.Close()
		if getResp.StatusCode != http.StatusOK {
			t.Fatalf("want 200 got %d", getResp.StatusCode)
		}
		var r struct {
			TeamName string `json:"team_name"`
			Members  []struct {
				UserID   string `json:"user_id"`
				Username string `json:"username"`
				IsActive bool   `json:"is_active"`
			} `json:"members"`
		}
		if err := json.NewDecoder(getResp.Body).Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if r.TeamName != "core" || len(r.Members) != 2 {
			t.Fatalf("unexpected resp %+v", r)
		}
	})
}
