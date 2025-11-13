CREATE TABLE users (
   id UUID PRIMARY KEY,
   name TEXT NOT NULL,
   is_active BOOLEAN NOT NULL DEFAULT TRUE,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE teams (
   id UUID PRIMARY KEY,
   name TEXT NOT NULL UNIQUE,
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE team_members (
  team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  PRIMARY KEY (team_id, user_id)
);

CREATE TABLE prs (
     id UUID PRIMARY KEY,
     title TEXT NOT NULL,
     author_id UUID NOT NULL REFERENCES users(id),
     status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
     created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
     merged_at TIMESTAMPTZ NULL,
     updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE pr_reviewers (
  pr_id UUID NOT NULL REFERENCES prs(id) ON DELETE CASCADE,
  reviewer_id UUID NOT NULL REFERENCES users(id),
  assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (pr_id, reviewer_id)
);