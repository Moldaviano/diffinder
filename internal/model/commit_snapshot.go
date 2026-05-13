package model

import (
	"time"

	"github.com/google/uuid"
)

type CommitSnapshot struct {
	ID            uuid.UUID `json:"id"`
	ReleaseID     uuid.UUID `json:"release_id"`
	CommitSHA     string    `json:"commit_sha"`
	CommitMessage string    `json:"commit_message"`
	Author        string    `json:"author"`
	CommittedAt   time.Time `json:"committed_at"`
	CapturedAt    time.Time `json:"captured_at"`
}
