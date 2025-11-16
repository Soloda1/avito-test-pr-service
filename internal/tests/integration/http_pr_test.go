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

func buildPRDeps(t *testing.T) (input.PRInputPort, input.TeamInputPort, input.UserInputPort) {
	log := logger.New("test")
	u := uow.NewPostgresUOW(pgC.Pool, log)
	selector := reviewerselector.NewRandomReviewerSelector()
	prSvc := pr.NewService(u, selector, log)
	teamSvc := team.NewService(u, log)
	userSvc := user.NewService(u, log)
	return prSvc, teamSvc, userSvc
}

func insertUserHTTP(t *testing.T, id, name string, active bool) {
	_, err := pgC.Pool.Exec(testCtx, `INSERT INTO users(id, name, is_active, created_at, updated_at) VALUES ($1,$2,$3,now(),now())`, id, name, active)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
}

func insertTeamHTTP(t *testing.T, name string) string {
	row := pgC.Pool.QueryRow(testCtx, `INSERT INTO teams(id, name, created_at, updated_at) VALUES (gen_random_uuid(),$1,now(),now()) RETURNING id`, name)
	var id string
	if err := row.Scan(&id); err != nil {
		t.Fatalf("insert team: %v", err)
	}
	return id
}

func addMemberHTTP(t *testing.T, teamID, userID string) {
	_, err := pgC.Pool.Exec(testCtx, `INSERT INTO team_members(team_id, user_id) VALUES ($1,$2)`, teamID, userID)
	if err != nil {
		t.Fatalf("add member: %v", err)
	}
}

func postJSONPR(baseURL, path string, body any) (*http.Response, error) {
	b, _ := json.Marshal(body)
	return http.Post(baseURL+path, "application/json", bytes.NewReader(b))
}

func TestPRHandlers_HTTPIntegration(t *testing.T) {
	if pgC == nil {
		t.Fatal("postgres not init")
	}

	prSvc, teamSvc, userSvc := buildPRDeps(t)
	log := logger.New("test")
	r := apihttp.NewRouter(log, prSvc, teamSvc, userSvc)
	cfg := &config.Config{HTTPServer: config.HTTPServer{RequestTimeout: 5 * time.Second}}
	r.Setup(cfg)
	server := httptest.NewServer(r.GetRouter())
	defer server.Close()
	baseURL := server.URL

	t.Run("CreatePR happy", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUserHTTP(t, "u1", "author", true)
		insertUserHTTP(t, "u2", "rev1", true)
		teamID := insertTeamHTTP(t, "core")
		addMemberHTTP(t, teamID, "u1")
		addMemberHTTP(t, teamID, "u2")
		resp, err := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-1", "pull_request_name": "title", "author_id": "u1"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("want 201 got %d", resp.StatusCode)
		}
		var r struct {
			PR struct {
				PullRequestID     string   `json:"pull_request_id"`
				AssignedReviewers []string `json:"assigned_reviewers"`
				Status            string   `json:"status"`
			} `json:"pr"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if r.PR.PullRequestID != "pr-1" || r.PR.Status != "OPEN" {
			t.Fatalf("bad resp %+v", r)
		}
		reviewersDB, err := GetPRReviewers(testCtx, pgC.Pool, "pr-1")
		if err != nil {
			t.Fatalf("db reviewers: %v", err)
		}
		if !EqualStringSets(reviewersDB, r.PR.AssignedReviewers) {
			t.Fatalf("reviewers mismatch db=%v json=%v", reviewersDB, r.PR.AssignedReviewers)
		}
		if HasDuplicates(reviewersDB) {
			t.Fatalf("duplicates in db reviewers %v", reviewersDB)
		}
	})

	t.Run("CreatePR duplicate id -> 409", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUserHTTP(t, "u1", "author", true)
		teamID := insertTeamHTTP(t, "core")
		addMemberHTTP(t, teamID, "u1")
		first, _ := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-dup", "pull_request_name": "t", "author_id": "u1"})
		if err := first.Body.Close(); err != nil {
			t.Fatalf("close first body: %v", err)
		}
		second, err := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-dup", "pull_request_name": "t", "author_id": "u1"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := second.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if second.StatusCode != http.StatusConflict {
			t.Fatalf("want 409 got %d", second.StatusCode)
		}
	})

	t.Run("CreatePR author not found -> 404", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		resp, err := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-x", "pull_request_name": "t", "author_id": "missing"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("want 404 got %d", resp.StatusCode)
		}
	})

	t.Run("MergePR happy", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUserHTTP(t, "u1", "author", true)
		teamID := insertTeamHTTP(t, "core")
		addMemberHTTP(t, teamID, "u1")
		createResp, _ := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-m", "pull_request_name": "t", "author_id": "u1"})
		if err := createResp.Body.Close(); err != nil {
			t.Fatalf("close createResp body: %v", err)
		}
		mergeResp, err := postJSONPR(baseURL, "/pullRequest/merge", map[string]any{"pull_request_id": "pr-m"})
		if err != nil {
			t.Fatalf("merge post: %v", err)
		}
		defer func() {
			if err := mergeResp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if mergeResp.StatusCode != http.StatusOK {
			t.Fatalf("want 200 got %d", mergeResp.StatusCode)
		}
		var r struct {
			PR struct {
				Status   string     `json:"status"`
				MergedAt *time.Time `json:"mergedAt"`
			} `json:"pr"`
		}
		if err := json.NewDecoder(mergeResp.Body).Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if r.PR.Status != "MERGED" || r.PR.MergedAt == nil {
			t.Fatalf("not merged %+v", r)
		}
	})

	t.Run("MergePR not found -> 404", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		resp, err := postJSONPR(baseURL, "/pullRequest/merge", map[string]any{"pull_request_id": "missing"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("want 404 got %d", resp.StatusCode)
		}
	})

	t.Run("Reassign happy", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUserHTTP(t, "u1", "author", true)
		insertUserHTTP(t, "u2", "old", true)
		insertUserHTTP(t, "u3", "new", false)
		teamID := insertTeamHTTP(t, "core")
		addMemberHTTP(t, teamID, "u1")
		addMemberHTTP(t, teamID, "u2")
		addMemberHTTP(t, teamID, "u3")
		createResp, _ := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-r", "pull_request_name": "t", "author_id": "u1"})
		if err := createResp.Body.Close(); err != nil {
			t.Fatalf("close createResp body: %v", err)
		}

		_, err := pgC.Pool.Exec(testCtx, `UPDATE users SET is_active=true WHERE id='u3'`)
		if err != nil {
			t.Fatalf("activate u3: %v", err)
		}

		resp, err := postJSONPR(baseURL, "/pullRequest/reassign", map[string]any{"pull_request_id": "pr-r", "old_user_id": "u2"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("want 200 got %d", resp.StatusCode)
		}
		var r struct {
			PR struct {
				AssignedReviewers []string `json:"assigned_reviewers"`
			} `json:"pr"`
			ReplacedBy string `json:"replaced_by"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if r.ReplacedBy != "u3" {
			t.Fatalf("expected replaced_by u3 got %s", r.ReplacedBy)
		}
		for _, rv := range r.PR.AssignedReviewers {
			if rv == "u2" {
				t.Fatalf("old reviewer still present %+v", r.PR.AssignedReviewers)
			}
		}
		reviewersDB, err := GetPRReviewers(testCtx, pgC.Pool, "pr-r")
		if err != nil {
			t.Fatalf("db reviewers: %v", err)
		}
		if !EqualStringSets(reviewersDB, r.PR.AssignedReviewers) {
			t.Fatalf("reviewers mismatch db=%v json=%v", reviewersDB, r.PR.AssignedReviewers)
		}
		if HasDuplicates(reviewersDB) {
			t.Fatalf("duplicates after reassign %v", reviewersDB)
		}
		if len(reviewersDB) != len(r.PR.AssignedReviewers) {
			t.Fatalf("length mismatch db=%d json=%d", len(reviewersDB), len(r.PR.AssignedReviewers))
		}
	})

	t.Run("Reassign PR not found -> 404", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		resp, err := postJSONPR(baseURL, "/pullRequest/reassign", map[string]any{"pull_request_id": "missing", "old_user_id": "u2"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("want 404 got %d", resp.StatusCode)
		}
	})

	t.Run("Reassign old reviewer not assigned -> 409", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUserHTTP(t, "u1", "author", true)
		teamID := insertTeamHTTP(t, "core")
		addMemberHTTP(t, teamID, "u1")
		createResp, _ := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-z", "pull_request_name": "t", "author_id": "u1"})
		if err := createResp.Body.Close(); err != nil {
			t.Fatalf("close createResp body: %v", err)
		}
		resp, err := postJSONPR(baseURL, "/pullRequest/reassign", map[string]any{"pull_request_id": "pr-z", "old_user_id": "uX"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("want 409 got %d", resp.StatusCode)
		}
	})

	t.Run("Reassign merged -> 409", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUserHTTP(t, "u1", "author", true)
		insertUserHTTP(t, "u2", "old", true)
		teamID := insertTeamHTTP(t, "core")
		addMemberHTTP(t, teamID, "u1")
		addMemberHTTP(t, teamID, "u2")
		createResp, _ := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-mg", "pull_request_name": "t", "author_id": "u1"})
		if err := createResp.Body.Close(); err != nil {
			t.Fatalf("close createResp body: %v", err)
		}

		mergeResp, _ := postJSONPR(baseURL, "/pullRequest/merge", map[string]any{"pull_request_id": "pr-mg"})
		if err := mergeResp.Body.Close(); err != nil {
			t.Fatalf("close mergeResp body: %v", err)
		}
		resp, err := postJSONPR(baseURL, "/pullRequest/reassign", map[string]any{"pull_request_id": "pr-mg", "old_user_id": "u2"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("want 409 got %d", resp.StatusCode)
		}
	})

	t.Run("Reassign no candidates -> 409", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUserHTTP(t, "u1", "author", true)
		insertUserHTTP(t, "u2", "old", true)
		teamID := insertTeamHTTP(t, "core")
		addMemberHTTP(t, teamID, "u1")
		addMemberHTTP(t, teamID, "u2")
		createResp, _ := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-nc", "pull_request_name": "t", "author_id": "u1"})
		if err := createResp.Body.Close(); err != nil {
			t.Fatalf("close createResp body: %v", err)
		}
		resp, err := postJSONPR(baseURL, "/pullRequest/reassign", map[string]any{"pull_request_id": "pr-nc", "old_user_id": "u2"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("want 409 got %d", resp.StatusCode)
		}
	})

	t.Run("CreatePR validation error missing id", func(t *testing.T) {
		resp, err := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_name": "t", "author_id": "u1"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("want 400 got %d", resp.StatusCode)
		}
	})
	t.Run("Reassign validation error missing old_user_id", func(t *testing.T) {
		resp, err := postJSONPR(baseURL, "/pullRequest/reassign", map[string]any{"pull_request_id": "pr-x"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("want 400 got %d", resp.StatusCode)
		}
	})
	t.Run("Merge validation error missing id", func(t *testing.T) {
		resp, err := postJSONPR(baseURL, "/pullRequest/merge", map[string]any{})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("want 400 got %d", resp.StatusCode)
		}
	})
	t.Run("CreatePR more than 2 candidates -> max 2 assigned", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUserHTTP(t, "u1", "author", true)
		insertUserHTTP(t, "u2", "rev2", true)
		insertUserHTTP(t, "u3", "rev3", true)
		insertUserHTTP(t, "u4", "rev4", true)
		teamID := insertTeamHTTP(t, "core")
		addMemberHTTP(t, teamID, "u1")
		addMemberHTTP(t, teamID, "u2")
		addMemberHTTP(t, teamID, "u3")
		addMemberHTTP(t, teamID, "u4")
		resp, err := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-many", "pull_request_name": "title", "author_id": "u1"})
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("want 201 got %d", resp.StatusCode)
		}
		var r struct {
			PR struct {
				PullRequestID     string   `json:"pull_request_id"`
				AssignedReviewers []string `json:"assigned_reviewers"`
				Status            string   `json:"status"`
			} `json:"pr"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if r.PR.PullRequestID != "pr-many" || r.PR.Status != "OPEN" {
			t.Fatalf("meta mismatch %+v", r)
		}
		if len(r.PR.AssignedReviewers) != 2 {
			t.Fatalf("expected 2 reviewers got %d (%+v)", len(r.PR.AssignedReviewers), r.PR.AssignedReviewers)
		}
		allowed := map[string]struct{}{"u2": {}, "u3": {}, "u4": {}}
		seen := make(map[string]struct{})
		for _, rv := range r.PR.AssignedReviewers {
			if rv == "u1" {
				t.Fatalf("author assigned: %+v", r.PR.AssignedReviewers)
			}
			if _, ok := allowed[rv]; !ok {
				t.Fatalf("unexpected reviewer %s", rv)
			}
			if _, dup := seen[rv]; dup {
				t.Fatalf("duplicate reviewer %s", rv)
			}
			seen[rv] = struct{}{}
		}
		reviewersDB, err := GetPRReviewers(testCtx, pgC.Pool, "pr-many")
		if err != nil {
			t.Fatalf("db reviewers: %v", err)
		}
		if len(reviewersDB) != 2 {
			t.Fatalf("db should have 2 reviewers got %d %v", len(reviewersDB), reviewersDB)
		}
		if !EqualStringSets(reviewersDB, r.PR.AssignedReviewers) {
			t.Fatalf("mismatch db=%v json=%v", reviewersDB, r.PR.AssignedReviewers)
		}
		if HasDuplicates(reviewersDB) {
			t.Fatalf("duplicates in db reviewers %v", reviewersDB)
		}
	})

	t.Run("MergePR idempotent returns same merged_at on second call", func(t *testing.T) {
		if err := TruncateAll(testCtx, pgC.Pool); err != nil {
			t.Fatalf("truncate: %v", err)
		}
		insertUserHTTP(t, "u1", "author", true)
		teamID := insertTeamHTTP(t, "core")
		addMemberHTTP(t, teamID, "u1")
		// Create PR
		createResp, err := postJSONPR(baseURL, "/pullRequest/create", map[string]any{"pull_request_id": "pr-idem", "pull_request_name": "title", "author_id": "u1"})
		if err != nil {
			t.Fatalf("create post: %v", err)
		}
		if err := createResp.Body.Close(); err != nil {
			t.Fatalf("close createResp body: %v", err)
		}
		// First merge
		firstMergeResp, err := postJSONPR(baseURL, "/pullRequest/merge", map[string]any{"pull_request_id": "pr-idem"})
		if err != nil {
			t.Fatalf("first merge post: %v", err)
		}
		if firstMergeResp.StatusCode != http.StatusOK {
			t.Fatalf("first merge status want 200 got %d", firstMergeResp.StatusCode)
		}
		var firstPayload struct {
			PR struct {
				Status   string     `json:"status"`
				MergedAt *time.Time `json:"mergedAt"`
			} `json:"pr"`
		}
		if err := json.NewDecoder(firstMergeResp.Body).Decode(&firstPayload); err != nil {
			t.Fatalf("decode first merge: %v", err)
		}
		if err := firstMergeResp.Body.Close(); err != nil {
			t.Fatalf("close firstMergeResp body: %v", err)
		}
		if firstPayload.PR.Status != "MERGED" || firstPayload.PR.MergedAt == nil {
			t.Fatalf("first merge invalid payload %+v", firstPayload)
		}
		initialMergedAtJSON := firstPayload.PR.MergedAt.UTC().Truncate(time.Microsecond)
		// Fetch DB state after first merge
		statusDB1, mergedAtDB1, err := GetPRStatusMerged(testCtx, pgC.Pool, "pr-idem")
		if err != nil {
			t.Fatalf("db fetch after first merge: %v", err)
		}
		if statusDB1 != "MERGED" || mergedAtDB1 == nil {
			t.Fatalf("db state after first merge invalid status=%s mergedAt=%v", statusDB1, mergedAtDB1)
		}
		initialMergedAtDB := mergedAtDB1.UTC().Truncate(time.Microsecond)
		if !initialMergedAtDB.Equal(initialMergedAtJSON) {
			t.Fatalf("JSON vs DB merged_at differ first merge json=%v db=%v", initialMergedAtJSON, initialMergedAtDB)
		}
		// Pause
		time.Sleep(25 * time.Millisecond)
		// Second merge (idempotent)
		secondMergeResp, err := postJSONPR(baseURL, "/pullRequest/merge", map[string]any{"pull_request_id": "pr-idem"})
		if err != nil {
			t.Fatalf("second merge post: %v", err)
		}
		defer func() {
			if err := secondMergeResp.Body.Close(); err != nil {
				t.Logf("body close: %v", err)
			}
		}()
		if secondMergeResp.StatusCode != http.StatusOK {
			t.Fatalf("second merge status want 200 got %d", secondMergeResp.StatusCode)
		}
		var secondPayload struct {
			PR struct {
				Status   string     `json:"status"`
				MergedAt *time.Time `json:"mergedAt"`
			} `json:"pr"`
		}
		if err := json.NewDecoder(secondMergeResp.Body).Decode(&secondPayload); err != nil {
			t.Fatalf("decode second merge: %v", err)
		}
		if err := secondMergeResp.Body.Close(); err != nil {
			t.Fatalf("close secondMergeResp body: %v", err)
		}
		if secondPayload.PR.Status != "MERGED" || secondPayload.PR.MergedAt == nil {
			t.Fatalf("second merge invalid payload %+v", secondPayload)
		}
		secondMergedAtJSON := secondPayload.PR.MergedAt.UTC().Truncate(time.Microsecond)
		// DB state after second merge
		statusDB2, mergedAtDB2, err := GetPRStatusMerged(testCtx, pgC.Pool, "pr-idem")
		if err != nil {
			t.Fatalf("db fetch after second merge: %v", err)
		}
		if statusDB2 != "MERGED" || mergedAtDB2 == nil {
			t.Fatalf("db state second merge invalid status=%s mergedAt=%v", statusDB2, mergedAtDB2)
		}
		secondMergedAtDB := mergedAtDB2.UTC().Truncate(time.Microsecond)
		// Assertions: timestamps unchanged
		if !secondMergedAtDB.Equal(initialMergedAtDB) {
			t.Fatalf("DB merged_at changed first=%v second=%v", initialMergedAtDB, secondMergedAtDB)
		}
		if !secondMergedAtJSON.Equal(initialMergedAtJSON) {
			t.Fatalf("JSON merged_at changed first=%v second=%v", initialMergedAtJSON, secondMergedAtJSON)
		}
		if !secondMergedAtDB.Equal(secondMergedAtJSON) {
			t.Fatalf("DB vs JSON differ on second merge db=%v json=%v", secondMergedAtDB, secondMergedAtJSON)
		}
	})
}
