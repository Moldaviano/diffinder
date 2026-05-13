package model

import (
	"time"

	"github.com/google/uuid"
)

type ReleaseStatus string

const (
	ReleaseDraft    ReleaseStatus = "draft"
	ReleaseInDev    ReleaseStatus = "in_dev"
	ReleaseInCert   ReleaseStatus = "in_cert"
	ReleaseApproved ReleaseStatus = "approved"
	ReleaseInProd   ReleaseStatus = "in_prod"
	ReleaseRejected ReleaseStatus = "rejected"
)

func (s ReleaseStatus) Valid() bool {
	switch s {
	case ReleaseDraft, ReleaseInDev, ReleaseInCert,
		ReleaseApproved, ReleaseInProd, ReleaseRejected:
		return true
	}
	return false
}

type Release struct {
	ID          uuid.UUID     `json:"id"`
	ProjectID   uuid.UUID     `json:"project_id"`
	BranchName  string        `json:"branch_name"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Status      ReleaseStatus `json:"status"`
	CreatedBy   *uuid.UUID    `json:"created_by,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
