package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gov-waf/waf-control/internal/httpx"
	"github.com/gov-waf/waf-control/pkg/pagination"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	resp, err := h.svc.Login(r.Context(), &req)
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, resp)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	p := pagination.FromRequest(r)
	items, total, err := h.svc.ListUsers(r.Context(), p.Offset(), p.Limit())
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, pagination.NewList(items, total, p.Page, p.PageSize))
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	resp, err := h.svc.CreateUser(r.Context(), &req)
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, resp)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetUser(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, resp)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req UpdateUserRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	resp, err := h.svc.UpdateUser(r.Context(), chi.URLParam(r, "id"), &req)
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, resp)
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListRoles(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, items)
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest
	if !httpx.DecodeJSON(w, r, &req) {
		return
	}
	resp, err := h.svc.CreateRole(r.Context(), &req)
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, resp)
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListPermissions(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	httpx.OK(w, items)
}

func writeErr(w http.ResponseWriter, err error) {
	if ae, ok := httpx.AsAppError(err); ok {
		httpx.Fail(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	httpx.Fail(w, 500, httpx.CodeInternal, "internal server error")
}
