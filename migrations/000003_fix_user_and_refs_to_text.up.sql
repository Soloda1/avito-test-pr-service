BEGIN;
ALTER TABLE IF EXISTS pr_reviewers DROP CONSTRAINT IF EXISTS pr_reviewers_reviewer_id_fkey;
ALTER TABLE IF EXISTS prs DROP CONSTRAINT IF EXISTS prs_author_id_fkey;
ALTER TABLE IF EXISTS team_members DROP CONSTRAINT IF EXISTS team_members_user_id_fkey;

ALTER TABLE users ALTER COLUMN id TYPE TEXT USING id::text;
ALTER TABLE team_members ALTER COLUMN user_id TYPE TEXT USING user_id::text;
ALTER TABLE prs ALTER COLUMN author_id TYPE TEXT USING author_id::text;
ALTER TABLE pr_reviewers ALTER COLUMN reviewer_id TYPE TEXT USING reviewer_id::text;

ALTER TABLE team_members
  ADD CONSTRAINT team_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE prs
  ADD CONSTRAINT prs_author_id_fkey FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE pr_reviewers
  ADD CONSTRAINT pr_reviewers_reviewer_id_fkey FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE CASCADE;
COMMIT;

