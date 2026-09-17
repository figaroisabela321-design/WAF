package site

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/gov-waf/waf-control/internal/httpx"
)

// Service contains site business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ValidateCreate validates create request fields.
func ValidateCreate(req *CreateRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return httpx.Validation("name is required")
	}
	if strings.TrimSpace(req.Domain) == "" {
		return httpx.Validation("domain is required")
	}
	if req.UpstreamPort < 1 || req.UpstreamPort > 65535 {
		return httpx.Validation("upstream_port must be between 1 and 65535")
	}
	proto := strings.ToLower(req.Protocol)
	if proto != "http" && proto != "https" {
		return httpx.Validation("protocol must be http or https")
	}
	mode := strings.ToLower(req.ProtectionMode)
	if mode != "" && mode != "observe" && mode != "protect" {
		return httpx.Validation("protection_mode must be observe or protect")
	}
	status := strings.ToLower(req.Status)
	if status != "" && status != "enabled" && status != "disabled" {
		return httpx.Validation("status must be enabled or disabled")
	}
	return nil
}

func (s *Service) Create(ctx context.Context, req *CreateRequest) (*Response, error) {
	if err := ValidateCreate(req); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	site := &Site{
		ID:             uuid.New(),
		Name:           strings.TrimSpace(req.Name),
		Domain:         strings.TrimSpace(req.Domain),
		UpstreamHost:   strings.TrimSpace(req.UpstreamHost),
		UpstreamPort:   req.UpstreamPort,
		Protocol:       strings.ToLower(req.Protocol),
		ProtectionMode: defaultStr(strings.ToLower(req.ProtectionMode), "observe"),
		Status:         defaultStr(strings.ToLower(req.Status), "enabled"),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if site.UpstreamHost == "" {
		site.UpstreamHost = site.Domain
	}
	if req.PolicyID != nil && *req.PolicyID != "" {
		id, err := uuid.Parse(*req.PolicyID)
		if err != nil {
			return nil, httpx.Validation("invalid policy_id")
		}
		site.PolicyID = &id
	}
	if req.NodeGroupID != nil && *req.NodeGroupID != "" {
		id, err := uuid.Parse(*req.NodeGroupID)
		if err != nil {
			return nil, httpx.Validation("invalid node_group_id")
		}
		site.NodeGroupID = &id
	}
	if err := s.repo.Create(ctx, site); err != nil {
		return nil, httpx.Internal("failed to create site", err)
	}
	resp := ToResponse(site)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, idStr string) (*Response, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid site id")
	}
	site, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get site", err)
	}
	if site == nil {
		return nil, httpx.NotFound("site not found")
	}
	resp := ToResponse(site)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, idStr string, req *UpdateRequest) (*Response, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid site id")
	}
	site, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get site", err)
	}
	if site == nil {
		return nil, httpx.NotFound("site not found")
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, httpx.Validation("name is required")
		}
		site.Name = strings.TrimSpace(*req.Name)
	}
	if req.Domain != nil {
		if strings.TrimSpace(*req.Domain) == "" {
			return nil, httpx.Validation("domain is required")
		}
		site.Domain = strings.TrimSpace(*req.Domain)
	}
	if req.UpstreamHost != nil {
		site.UpstreamHost = strings.TrimSpace(*req.UpstreamHost)
	}
	if req.UpstreamPort != nil {
		if *req.UpstreamPort < 1 || *req.UpstreamPort > 65535 {
			return nil, httpx.Validation("upstream_port must be between 1 and 65535")
		}
		site.UpstreamPort = *req.UpstreamPort
	}
	if req.Protocol != nil {
		p := strings.ToLower(*req.Protocol)
		if p != "http" && p != "https" {
			return nil, httpx.Validation("protocol must be http or https")
		}
		site.Protocol = p
	}
	if req.ProtectionMode != nil {
		m := strings.ToLower(*req.ProtectionMode)
		if m != "observe" && m != "protect" {
			return nil, httpx.Validation("protection_mode must be observe or protect")
		}
		site.ProtectionMode = m
	}
	if req.Status != nil {
		st := strings.ToLower(*req.Status)
		if st != "enabled" && st != "disabled" {
			return nil, httpx.Validation("status must be enabled or disabled")
		}
		site.Status = st
	}
	if req.PolicyID != nil {
		if *req.PolicyID == "" {
			site.PolicyID = nil
		} else {
			pid, err := uuid.Parse(*req.PolicyID)
			if err != nil {
				return nil, httpx.Validation("invalid policy_id")
			}
			site.PolicyID = &pid
		}
	}
	if req.NodeGroupID != nil {
		if *req.NodeGroupID == "" {
			site.NodeGroupID = nil
		} else {
			nid, err := uuid.Parse(*req.NodeGroupID)
			if err != nil {
				return nil, httpx.Validation("invalid node_group_id")
			}
			site.NodeGroupID = &nid
		}
	}
	site.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, site); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.NotFound("site not found")
		}
		return nil, httpx.Internal("failed to update site", err)
	}
	resp := ToResponse(site)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httpx.Validation("invalid site id")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if err == pgx.ErrNoRows {
			return httpx.NotFound("site not found")
		}
		return httpx.Internal("failed to delete site", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]Response, int, error) {
	list, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, httpx.Internal("failed to list sites", err)
	}
	return ToResponseList(list), total, nil
}

func defaultStr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
