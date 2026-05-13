package model

import (
	"time"

	"github.com/google/uuid"
)

type PRStatus string

const (
	PROpen    PRStatus = "open"
	PRMerged  PRStatus = "merged"
	PRBlocked PRStatus = "blocked"
	PRClosed  PRStatus = "closed"
)

func (s PRStatus) Valid() bool {
	switch s {
	case PROpen, PRMerged, PRBlocked, PRClosed:
		return true
	}
	return false
}

type PullRequest struct {
	ID            uuid.UUID  `json:"id"`
	ReleaseID     uuid.UUID  `json:"release_id"`
	PRURL         string     `json:"pr_url"`
	PRNumber      int        `json:"pr_number"`
	HeadCommitSHA string     `json:"head_commit_sha"`
	BaseBranch    string     `json:"base_branch"`
	Status        PRStatus   `json:"status"`
	OpenedAt      time.Time  `json:"opened_at"`
	MergedAt      *time.Time `json:"merged_at,omitempty"`
}
