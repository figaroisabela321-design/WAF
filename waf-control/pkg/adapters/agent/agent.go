package agent

import "context"

// Client talks to WAF data-plane agents (Phase 2+).
type Client interface {
	PushConfig(ctx context.Context, nodeID string, config []byte) error
	HealthCheck(ctx context.Context, nodeID string) (online bool, err error)
}

// NoopClient is a no-op agent client for Phase 1.
type NoopClient struct{}

func NewNoop() *NoopClient { return &NoopClient{} }

func (n *NoopClient) PushConfig(ctx context.Context, nodeID string, config []byte) error {
	return nil
}

func (n *NoopClient) HealthCheck(ctx context.Context, nodeID string) (bool, error) {
	return false, nil
}
