// Tipi mirror dei modelli Go. Tutti i timestamp sono stringhe ISO.

export type UserRole = 'admin' | 'developer' | 'viewer';
export type ReleaseStatus = 'draft' | 'in_dev' | 'in_cert' | 'approved' | 'in_prod' | 'rejected';
export type Environment = 'dev' | 'cert' | 'prod';
export type PRStatus = 'open' | 'merged' | 'blocked' | 'closed';

export interface User {
  id: string;
  username: string;
  email: string;
  role: UserRole;
  created_at: string;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  repository_url: string;
  webhook_token?: string;
  created_at: string;
  updated_at: string;
}

export interface ProjectStats {
  project_id: string;
  active_releases: number;
  last_activity?: string;
}

export interface Release {
  id: string;
  project_id: string;
  branch_name: string;
  title: string;
  description: string;
  status: ReleaseStatus;
  created_by?: string;
  created_at: string;
  updated_at: string;
}

export interface DeploymentEvent {
  id: string;
  release_id: string;
  environment: Environment;
  commit_sha: string;
  deployed_by?: string;
  deployed_at: string;
  notes: string;
}

export interface CommitSnapshot {
  id: string;
  release_id: string;
  commit_sha: string;
  commit_message: string;
  author: string;
  committed_at: string;
  captured_at: string;
}

export interface PullRequest {
  id: string;
  release_id: string;
  pr_url: string;
  pr_number: number;
  head_commit_sha: string;
  base_branch: string;
  status: PRStatus;
  opened_at: string;
  merged_at?: string;
}

export interface CertificationCheck {
  id: string;
  pull_request_id: string;
  head_commit_sha: string;
  cert_commit_sha: string;
  passed: boolean;
  checked_at: string;
  details: string;
}

export interface Paged<T> {
  items: T[];
  page: number;
  limit: number;
  total: number;
}

export interface Summary {
  total_releases: number;
  in_cert: number;
  blocked_prs: number;
  deployments_today: number;
}

export interface StatusCount { status: string; count: number; }

export interface ActivityItem {
  type: string;
  release_id: string;
  title: string;
  environment?: string;
  commit_sha?: string;
  at: string;
}

export interface ApiError { error: string; code: string; }

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  user: User;
}
