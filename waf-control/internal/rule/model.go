package rule

import (
	"time"

	"github.com/google/uuid"
)

type Rule struct {
	ID        uuid.UUID
	RuleID    string
	Name      string
	Category  string
	Source    string
	Severity  string
	Enabled   bool
	Scope     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
