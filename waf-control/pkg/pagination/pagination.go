package pagination

import (
	"net/http"
	"strconv"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Params holds pagination query parameters.
type Params struct {
	Page     int
	PageSize int
}

// Offset returns SQL OFFSET.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns SQL LIMIT (page_size).
func (p Params) Limit() int {
	return p.PageSize
}

// FromRequest parses page and page_size from query string.
func FromRequest(r *http.Request) Params {
	page := DefaultPage
	pageSize := DefaultPageSize
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := r.URL.Query().Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pageSize = n
		}
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return Params{Page: page, PageSize: pageSize}
}

// ListData is the envelope data for list responses.
type ListData struct {
	Items    any `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// NewList builds ListData.
func NewList(items any, total, page, pageSize int) ListData {
	if items == nil {
		items = []any{}
	}
	return ListData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
