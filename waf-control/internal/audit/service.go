package audit

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/gov-waf/waf-control/internal/httpx"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Write(ctx context.Context, log *Log) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	return s.repo.Create(ctx, log)
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]Response, int, error) {
	list, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, httpx.Internal("failed to list audit logs", err)
	}
	return ToResponseList(list), total, nil
}
