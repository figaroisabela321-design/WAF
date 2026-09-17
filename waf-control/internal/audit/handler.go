package audit

import (
	"net/http"

	"github.com/gov-waf/waf-control/internal/httpx"
	"github.com/gov-waf/waf-control/pkg/pagination"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)
	items, total, err := h.svc.List(r.Context(), p.Offset(), p.Limit())
	if err != nil {
		if ae, ok := httpx.AsAppError(err); ok {
			httpx.Fail(w, ae.HTTPStatus, ae.Code, ae.Message)
			return
		}
		httpx.Fail(w, 500, httpx.CodeInternal, "internal server error")
		return
	}
	httpx.OK(w, pagination.NewList(items, total, p.Page, p.PageSize))
}
