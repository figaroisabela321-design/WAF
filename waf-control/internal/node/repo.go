package node

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, n *Node) error
	GetByID(ctx context.Context, id uuid.UUID) (*Node, error)
	Update(ctx context.Context, n *Node) error
	List(ctx context.Context, offset, limit int) ([]Node, int, error)
}
