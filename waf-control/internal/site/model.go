package site

import (
	"time"

	"github.com/google/uuid"
)

// Site is the DB model.
type Site struct {
	ID             uuid.UUID  `json:"-"`
	Name           string     `json:"-"`
	Domain         string     `json:"-"`
	UpstreamHost   string     `json:"-"`
	UpstreamPort   int        `json:"-"`
	Protocol       string     `json:"-"`
	ProtectionMode string     `json:"-"`
	PolicyID       *uuid.UUID `json:"-"`
	NodeGroupID    *uuid.UUID `json:"-"`
	Status         string     `json:"-"`
	CreatedAt      time.Time  `json:"-"`
	UpdatedAt      time.Time  `json:"-"`
}
