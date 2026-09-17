package policy

import (
	"time"

	"github.com/google/uuid"
)

type Policy struct {
	ID                uuid.UUID
	Name              string
	Mode              string
	ProtectionLevel   string
	ParanoiaLevel     int
	InboundThreshold  int
	OutboundThreshold int
	SQLInjection      bool
	XSS               bool
	RCE               bool
	LFI               bool
	Scanner           bool
	ProtocolAttack    bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
