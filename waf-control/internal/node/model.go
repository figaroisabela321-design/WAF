package node

import (
	"time"

	"github.com/google/uuid"
)

type Node struct {
	ID            uuid.UUID
	Name          string
	IP            string
	Region        string
	NodeGroup     string
	CorazaVersion string
	CRSVersion    string
	ConfigVersion string
	Status        string
	LastHeartbeat *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
