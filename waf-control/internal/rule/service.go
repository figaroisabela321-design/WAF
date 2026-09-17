package rule

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/gov-waf/waf-control/internal/httpx"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, req *CreateRequest) (*Response, error) {
	if strings.TrimSpace(req.RuleID) == "" {
		return nil, httpx.Validation("rule_id is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, httpx.Validation("name is required")
	}
	source := strings.ToLower(req.Source)
	if source == "" {
		source = "custom"
	}
	if source != "crs" && source != "custom" && source != "exclusion" {
		return nil, httpx.Validation("source must be crs, custom or exclusion")
	}
	scope := strings.ToLower(req.Scope)
	if scope == "" {
		scope = "global"
	}
	if scope != "global" && scope != "site" {
		return nil, httpx.Validation("scope must be global or site")
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	severity := req.Severity
	if severity == "" {
		severity = "medium"
	}
	now := time.Now().UTC()
	rule := &Rule{
		ID: uuid.New(), RuleID: strings.TrimSpace(req.RuleID), Name: strings.TrimSpace(req.Name),
		Category: req.Category, Source: source, Severity: severity, Enabled: enabled, Scope: scope,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, rule); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, httpx.Conflict("rule_id already exists")
		}
		return nil, httpx.Internal("failed to create rule", err)
	}
	resp := ToResponse(rule)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, idStr string) (*Response, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid rule id")
	}
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get rule", err)
	}
	if rule == nil {
		return nil, httpx.NotFound("rule not found")
	}
	resp := ToResponse(rule)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, idStr string, req *UpdateRequest) (*Response, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid rule id")
	}
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get rule", err)
	}
	if rule == nil {
		return nil, httpx.NotFound("rule not found")
	}
	if req.RuleID != nil {
		if strings.TrimSpace(*req.RuleID) == "" {
			return nil, httpx.Validation("rule_id is required")
		}
		rule.RuleID = strings.TrimSpace(*req.RuleID)
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, httpx.Validation("name is required")
		}
		rule.Name = strings.TrimSpace(*req.Name)
	}
	if req.Category != nil {
		rule.Category = *req.Category
	}
	if req.Source != nil {
		src := strings.ToLower(*req.Source)
		if src != "crs" && src != "custom" && src != "exclusion" {
			return nil, httpx.Validation("source must be crs, custom or exclusion")
		}
		rule.Source = src
	}
	if req.Severity != nil {
		rule.Severity = *req.Severity
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Scope != nil {
		sc := strings.ToLower(*req.Scope)
		if sc != "global" && sc != "site" {
			return nil, httpx.Validation("scope must be global or site")
		}
		rule.Scope = sc
	}
	rule.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, rule); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.NotFound("rule not found")
		}
		return nil, httpx.Internal("failed to update rule", err)
	}
	resp := ToResponse(rule)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httpx.Validation("invalid rule id")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err == pgx.ErrNoRows {
			return httpx.NotFound("rule not found")
		}
		return httpx.Internal("failed to delete rule", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]Response, int, error) {
	list, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, httpx.Internal("failed to list rules", err)
	}
	return ToResponseList(list), total, nil
}
