package model

import (
	"time"

	"github.com/google/uuid"
)

type Environment string

const (
	EnvDev  Environment = "dev"
	EnvCert Environment = "cert"
	EnvProd Environment = "prod"
)

func (e Environment) Valid() bool {
	switch e {
	case EnvDev, EnvCert, EnvProd:
		return true
	}
	return false
}

type DeploymentEvent struct {
	ID          uuid.UUID   `json:"id"`
	ReleaseID   uuid.UUID   `json:"release_id"`
	Environment Environment `json:"environment"`
	CommitSHA   string      `json:"commit_sha"`
	DeployedBy  *uuid.UUID  `json:"deployed_by,omitempty"`
	DeployedAt  time.Time   `json:"deployed_at"`
	Notes       string      `json:"notes"`
}
