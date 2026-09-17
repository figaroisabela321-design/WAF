package node

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
	if strings.TrimSpace(req.Name) == "" {
		return nil, httpx.Validation("name is required")
	}
	if strings.TrimSpace(req.IP) == "" {
		return nil, httpx.Validation("ip is required")
	}
	status := strings.ToLower(req.Status)
	if status == "" {
		status = "unknown"
	}
	if status != "online" && status != "offline" && status != "unknown" {
		return nil, httpx.Validation("status must be online, offline or unknown")
	}
	now := time.Now().UTC()
	n := &Node{
		ID: uuid.New(), Name: strings.TrimSpace(req.Name), IP: strings.TrimSpace(req.IP),
		Region: req.Region, NodeGroup: req.NodeGroup, CorazaVersion: req.CorazaVersion,
		CRSVersion: req.CRSVersion, ConfigVersion: req.ConfigVersion, Status: status,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, httpx.Internal("failed to create node", err)
	}
	resp := ToResponse(n)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, idStr string) (*Response, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid node id")
	}
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get node", err)
	}
	if n == nil {
		return nil, httpx.NotFound("node not found")
	}
	resp := ToResponse(n)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, idStr string, req *UpdateRequest) (*Response, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid node id")
	}
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get node", err)
	}
	if n == nil {
		return nil, httpx.NotFound("node not found")
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, httpx.Validation("name is required")
		}
		n.Name = strings.TrimSpace(*req.Name)
	}
	if req.IP != nil {
		if strings.TrimSpace(*req.IP) == "" {
			return nil, httpx.Validation("ip is required")
		}
		n.IP = strings.TrimSpace(*req.IP)
	}
	if req.Region != nil {
		n.Region = *req.Region
	}
	if req.NodeGroup != nil {
		n.NodeGroup = *req.NodeGroup
	}
	if req.CorazaVersion != nil {
		n.CorazaVersion = *req.CorazaVersion
	}
	if req.CRSVersion != nil {
		n.CRSVersion = *req.CRSVersion
	}
	if req.ConfigVersion != nil {
		n.ConfigVersion = *req.ConfigVersion
	}
	if req.Status != nil {
		st := strings.ToLower(*req.Status)
		if st != "online" && st != "offline" && st != "unknown" {
			return nil, httpx.Validation("status must be online, offline or unknown")
		}
		n.Status = st
	}
	n.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, n); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.NotFound("node not found")
		}
		return nil, httpx.Internal("failed to update node", err)
	}
	resp := ToResponse(n)
	return &resp, nil
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]Response, int, error) {
	list, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, httpx.Internal("failed to list nodes", err)
	}
	return ToResponseList(list), total, nil
}
