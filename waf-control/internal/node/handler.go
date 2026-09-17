package node

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gov-waf/waf-control/internal/httpx"
	"github.com/gov-waf/waf-control/pkg/pagination"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Routes(r chi.Router) {
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)
	items, total, err := h.svc.List(r.Context(), p.Offset(), p.Limit())
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, pagination.NewList(items, total, p.Page, p.PageSize))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	resp, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	resp, err := h.svc.Update(r.Context(), chi.URLParam(r, "id"), &req)
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, resp)
}

func writeErr(w http.ResponseWriter, err error) {
	if ae, ok := httpx.AsAppError(err); ok {
		httpx.Fail(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	httpx.Fail(w, 500, httpx.CodeInternal, "internal server error")
}
