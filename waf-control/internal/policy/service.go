package policy

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

func boolOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

func (s *Service) Create(ctx context.Context, req *CreateRequest) (*Response, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, httpx.Validation("name is required")
	}
	mode := strings.ToLower(req.Mode)
	if mode == "" {
		mode = "observe"
	}
	if mode != "observe" && mode != "protect" {
		return nil, httpx.Validation("mode must be observe or protect")
	}
	level := strings.ToLower(req.ProtectionLevel)
	if level == "" {
		level = "balanced"
	}
	if level != "loose" && level != "balanced" && level != "strict" {
		return nil, httpx.Validation("protection_level must be loose, balanced or strict")
	}
	pl := req.ParanoiaLevel
	if pl == 0 {
		pl = 1
	}
	if pl < 1 || pl > 4 {
		return nil, httpx.Validation("paranoia_level must be between 1 and 4")
	}
	inb, outb := req.InboundThreshold, req.OutboundThreshold
	if inb == 0 {
		inb = 5
	}
	if outb == 0 {
		outb = 4
	}
	now := time.Now().UTC()
	p := &Policy{
		ID: uuid.New(), Name: strings.TrimSpace(req.Name), Mode: mode,
		ProtectionLevel: level, ParanoiaLevel: pl, InboundThreshold: inb, OutboundThreshold: outb,
		SQLInjection: boolOr(req.SQLInjection, true), XSS: boolOr(req.XSS, true),
		RCE: boolOr(req.RCE, true), LFI: boolOr(req.LFI, true),
		Scanner: boolOr(req.Scanner, true), ProtocolAttack: boolOr(req.ProtocolAttack, true),
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, httpx.Internal("failed to create policy", err)
	}
	resp := ToResponse(p)
	return &resp, nil
}

func (s *Service) Get(ctx context.Context, idStr string) (*Response, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid policy id")
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get policy", err)
	}
	if p == nil {
		return nil, httpx.NotFound("policy not found")
	}
	resp := ToResponse(p)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, idStr string, req *UpdateRequest) (*Response, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, httpx.Validation("invalid policy id")
	}
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, httpx.Internal("failed to get policy", err)
	}
	if p == nil {
		return nil, httpx.NotFound("policy not found")
	}
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, httpx.Validation("name is required")
		}
		p.Name = strings.TrimSpace(*req.Name)
	}
	if req.Mode != nil {
		m := strings.ToLower(*req.Mode)
		if m != "observe" && m != "protect" {
			return nil, httpx.Validation("mode must be observe or protect")
		}
		p.Mode = m
	}
	if req.ProtectionLevel != nil {
		l := strings.ToLower(*req.ProtectionLevel)
		if l != "loose" && l != "balanced" && l != "strict" {
			return nil, httpx.Validation("protection_level must be loose, balanced or strict")
		}
		p.ProtectionLevel = l
	}
	if req.ParanoiaLevel != nil {
		if *req.ParanoiaLevel < 1 || *req.ParanoiaLevel > 4 {
			return nil, httpx.Validation("paranoia_level must be between 1 and 4")
		}
		p.ParanoiaLevel = *req.ParanoiaLevel
	}
	if req.InboundThreshold != nil {
		p.InboundThreshold = *req.InboundThreshold
	}
	if req.OutboundThreshold != nil {
		p.OutboundThreshold = *req.OutboundThreshold
	}
	if req.SQLInjection != nil {
		p.SQLInjection = *req.SQLInjection
	}
	if req.XSS != nil {
		p.XSS = *req.XSS
	}
	if req.RCE != nil {
		p.RCE = *req.RCE
	}
	if req.LFI != nil {
		p.LFI = *req.LFI
	}
	if req.Scanner != nil {
		p.Scanner = *req.Scanner
	}
	if req.ProtocolAttack != nil {
		p.ProtocolAttack = *req.ProtocolAttack
	}
	p.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, p); err != nil {
		if err == pgx.ErrNoRows {
			return nil, httpx.NotFound("policy not found")
		}
		return nil, httpx.Internal("failed to update policy", err)
	}
	resp := ToResponse(p)
	return &resp, nil
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]Response, int, error) {
	list, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, httpx.Internal("failed to list policies", err)
	}
	return ToResponseList(list), total, nil
}
