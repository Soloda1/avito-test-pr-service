CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_prs_author_id ON prs(author_id);
CREATE INDEX idx_pr_reviewers_reviewer_id_pr_id ON pr_reviewers(reviewer_id, pr_id);
