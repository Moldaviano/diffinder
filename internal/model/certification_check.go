package model

import (
	"time"

	"github.com/google/uuid"
)

type CertificationCheck struct {
	ID            uuid.UUID `json:"id"`
	PullRequestID uuid.UUID `json:"pull_request_id"`
	HeadCommitSHA string    `json:"head_commit_sha"`
	CertCommitSHA string    `json:"cert_commit_sha"`
	Passed        bool      `json:"passed"`
	CheckedAt     time.Time `json:"checked_at"`
	Details       string    `json:"details"`
}
