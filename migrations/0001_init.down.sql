DROP TRIGGER IF EXISTS trg_releases_updated_at ON releases;
DROP TRIGGER IF EXISTS trg_projects_updated_at ON projects;
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS certification_checks;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS commit_snapshots;
DROP TABLE IF EXISTS deployment_events;
DROP TABLE IF EXISTS releases;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS pr_status;
DROP TYPE IF EXISTS environment;
DROP TYPE IF EXISTS release_status;
DROP TYPE IF EXISTS user_role;
