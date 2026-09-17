package clickhouse

import (
	"context"
	"time"
)

// Event is a WAF security event (stored in ClickHouse in Phase 2+).
type Event struct {
	ID        string    `json:"id"`
	SiteID    string    `json:"site_id"`
	RuleID    string    `json:"rule_id"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	ClientIP  string    `json:"client_ip"`
	CreatedAt time.Time `json:"created_at"`
}

// EventStore queries WAF events from ClickHouse (Phase 2+).
type EventStore interface {
	List(ctx context.Context, page, pageSize int) ([]Event, int, error)
}

// NoopEventStore returns empty results.
type NoopEventStore struct{}

func NewNoopEventStore() *NoopEventStore { return &NoopEventStore{} }

func (n *NoopEventStore) List(ctx context.Context, page, pageSize int) ([]Event, int, error) {
	return []Event{}, 0, nil
}
