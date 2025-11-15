CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_prs_author_id ON prs(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer_id_pr_id ON pr_reviewers(reviewer_id, pr_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr_id_assigned_at ON pr_reviewers(pr_id, assigned_at);

