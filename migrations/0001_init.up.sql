-- Enable pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- ENUM TYPES
-- ============================================================
CREATE TYPE user_role AS ENUM ('admin', 'developer', 'viewer');

CREATE TYPE release_status AS ENUM (
    'draft',
    'in_dev',
    'in_cert',
    'approved',
    'in_prod',
    'rejected'
);

CREATE TYPE environment AS ENUM ('dev', 'cert', 'prod');

CREATE TYPE pr_status AS ENUM ('open', 'merged', 'blocked', 'closed');

-- ============================================================
-- USERS
-- ============================================================
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      TEXT NOT NULL UNIQUE,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role          user_role NOT NULL DEFAULT 'developer',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);

-- ============================================================
-- PROJECTS
-- ============================================================
CREATE TABLE projects (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT NOT NULL UNIQUE,
    description    TEXT NOT NULL DEFAULT '',
    repository_url TEXT NOT NULL DEFAULT '',
    webhook_token  TEXT NOT NULL DEFAULT '', -- token per identificare il progetto sul webhook
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_name ON projects (name);

-- ============================================================
-- RELEASES
-- ============================================================
CREATE TABLE releases (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    branch_name TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      release_status NOT NULL DEFAULT 'draft',
    created_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_releases_project ON releases (project_id);
CREATE INDEX idx_releases_status  ON releases (status);
CREATE INDEX idx_releases_created ON releases (created_at DESC);
CREATE UNIQUE INDEX idx_releases_project_branch ON releases (project_id, branch_name);

-- ============================================================
-- DEPLOYMENT EVENTS
-- ============================================================
CREATE TABLE deployment_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    release_id   UUID NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    environment  environment NOT NULL,
    commit_sha   TEXT NOT NULL,
    deployed_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    deployed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes        TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_deployments_release     ON deployment_events (release_id);
CREATE INDEX idx_deployments_env         ON deployment_events (environment);
CREATE INDEX idx_deployments_deployed_at ON deployment_events (deployed_at DESC);

-- ============================================================
-- COMMIT SNAPSHOTS (presi quando si entra in CERT)
-- ============================================================
CREATE TABLE commit_snapshots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    release_id      UUID NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    commit_sha      TEXT NOT NULL,
    commit_message  TEXT NOT NULL DEFAULT '',
    author          TEXT NOT NULL DEFAULT '',
    committed_at    TIMESTAMPTZ NOT NULL,
    captured_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_snapshots_release ON commit_snapshots (release_id);
CREATE UNIQUE INDEX idx_snapshots_release_sha ON commit_snapshots (release_id, commit_sha);

-- ============================================================
-- PULL REQUESTS
-- ============================================================
CREATE TABLE pull_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    release_id      UUID NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    pr_url          TEXT NOT NULL,
    pr_number       INTEGER NOT NULL,
    head_commit_sha TEXT NOT NULL,
    base_branch     TEXT NOT NULL,
    status          pr_status NOT NULL DEFAULT 'open',
    opened_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at       TIMESTAMPTZ
);

CREATE INDEX idx_pr_release ON pull_requests (release_id);
CREATE INDEX idx_pr_status  ON pull_requests (status);
CREATE UNIQUE INDEX idx_pr_number_release ON pull_requests (release_id, pr_number);

-- ============================================================
-- CERTIFICATION CHECKS
-- ============================================================
CREATE TABLE certification_checks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pull_request_id UUID NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    head_commit_sha TEXT NOT NULL,
    cert_commit_sha TEXT NOT NULL DEFAULT '',
    passed          BOOLEAN NOT NULL,
    checked_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    details         TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_checks_pr      ON certification_checks (pull_request_id);
CREATE INDEX idx_checks_passed  ON certification_checks (passed);
CREATE INDEX idx_checks_checked ON certification_checks (checked_at DESC);

-- ============================================================
-- TRIGGER: aggiorna updated_at su projects e releases
-- ============================================================
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_releases_updated_at
    BEFORE UPDATE ON releases
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
