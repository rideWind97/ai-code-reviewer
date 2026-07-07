CREATE TABLE IF NOT EXISTS provider_installations (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  external_id TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (provider, external_id)
);

CREATE TABLE IF NOT EXISTS repositories (
  id BIGSERIAL PRIMARY KEY,
  provider TEXT NOT NULL,
  full_name TEXT NOT NULL,
  default_branch TEXT NOT NULL DEFAULT 'main',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (provider, full_name)
);

CREATE TABLE IF NOT EXISTS change_requests (
  id BIGSERIAL PRIMARY KEY,
  repository_id BIGINT NOT NULL REFERENCES repositories(id),
  number INTEGER NOT NULL,
  base_sha TEXT NOT NULL,
  head_sha TEXT NOT NULL,
  state TEXT NOT NULL DEFAULT 'open',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (repository_id, number, head_sha)
);

CREATE TABLE IF NOT EXISTS review_runs (
  id UUID PRIMARY KEY,
  change_request_id BIGINT NOT NULL REFERENCES change_requests(id),
  state TEXT NOT NULL,
  failure_reason TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  finished_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS patch_sets (
  id BIGSERIAL PRIMARY KEY,
  review_run_id UUID NOT NULL REFERENCES review_runs(id),
  base_sha TEXT NOT NULL,
  head_sha TEXT NOT NULL,
  files JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS review_units (
  id TEXT PRIMARY KEY,
  review_run_id UUID NOT NULL REFERENCES review_runs(id),
  review_path TEXT NOT NULL,
  file_path TEXT NOT NULL,
  payload JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS review_findings (
  id BIGSERIAL PRIMARY KEY,
  review_run_id UUID NOT NULL REFERENCES review_runs(id),
  fingerprint TEXT NOT NULL,
  severity TEXT NOT NULL,
  category TEXT NOT NULL,
  file_path TEXT NOT NULL,
  line INTEGER NOT NULL,
  message TEXT NOT NULL,
  evidence_level TEXT NOT NULL,
  confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
  payload JSONB NOT NULL,
  UNIQUE (review_run_id, fingerprint)
);

CREATE TABLE IF NOT EXISTS published_comments (
  id BIGSERIAL PRIMARY KEY,
  review_run_id UUID NOT NULL REFERENCES review_runs(id),
  provider_comment_id TEXT NOT NULL DEFAULT '',
  comment_type TEXT NOT NULL,
  file_path TEXT NOT NULL DEFAULT '',
  line INTEGER NOT NULL DEFAULT 0,
  body TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
