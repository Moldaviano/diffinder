package model

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	RepositoryURL string    `json:"repository_url"`
	WebhookToken  string    `json:"webhook_token,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
