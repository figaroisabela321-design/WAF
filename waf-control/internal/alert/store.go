package alert

import (
	"context"
	"net/http"
	"time"

	"github.com/gov-waf/waf-control/internal/httpx"
	"github.com/gov-waf/waf-control/pkg/pagination"
)

// Alert is a security alert (Phase 2+).
type Alert struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Severity  string    `json:"severity"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// AlertStore abstracts alert storage.
type AlertStore interface {
	List(ctx context.Context, page, pageSize int) ([]Alert, int, error)
}

// NoopStore returns empty lists.
type NoopStore struct{}

func NewNoopStore() *NoopStore { return &NoopStore{} }

func (n *NoopStore) List(ctx context.Context, page, pageSize int) ([]Alert, int, error) {
	return []Alert{}, 0, nil
}

type Handler struct{ store AlertStore }

func NewHandler(store AlertStore) *Handler { return &Handler{store: store} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)
	items, total, err := h.store.List(r.Context(), p.Page, p.PageSize)
	if err != nil {
		httpx.Fail(w, 500, httpx.CodeInternal, "internal server error")
		return
	}
	if items == nil {
		items = []Alert{}
	}
	httpx.WriteJSON(w, 200, 0, "ok (alerts not wired in Phase 1)", pagination.NewList(items, total, p.Page, p.PageSize))
}
