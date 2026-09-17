package event

import (
	"context"
	"net/http"
	"time"

	"github.com/gov-waf/waf-control/internal/httpx"
	"github.com/gov-waf/waf-control/pkg/pagination"
)

// Event is a WAF security event (ClickHouse in Phase 2+).
type Event struct {
	ID        string    `json:"id"`
	SiteID    string    `json:"site_id"`
	RuleID    string    `json:"rule_id"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// EventStore abstracts event storage.
type EventStore interface {
	List(ctx context.Context, page, pageSize int) ([]Event, int, error)
}

// NoopStore returns empty lists (ClickHouse not wired).
type NoopStore struct{}

func NewNoopStore() *NoopStore { return &NoopStore{} }

func (n *NoopStore) List(ctx context.Context, page, pageSize int) ([]Event, int, error) {
	return []Event{}, 0, nil
}

type Handler struct{ store EventStore }

func NewHandler(store EventStore) *Handler { return &Handler{store: store} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)
	items, total, err := h.store.List(r.Context(), p.Page, p.PageSize)
	if err != nil {
		httpx.Fail(w, 500, httpx.CodeInternal, "internal server error")
		return
	}
	if items == nil {
		items = []Event{}
	}
	httpx.WriteJSON(w, 200, 0, "ok (ClickHouse not wired in Phase 1)", pagination.NewList(items, total, p.Page, p.PageSize))
}
