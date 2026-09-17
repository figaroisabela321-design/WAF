package audit

import (
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID         uuid.UUID
	ActorID    *uuid.UUID
	ActorName  string
	Method     string
	Path       string
	Resource   string
	ResourceID string
	StatusCode int
	IP         string
	UserAgent  string
	CreatedAt  time.Time
}
