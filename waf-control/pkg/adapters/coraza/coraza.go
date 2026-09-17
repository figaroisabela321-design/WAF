package coraza

import "context"

// Engine is the Coraza WAF engine interface (Phase 2+).
type Engine interface {
	// EvaluateRequest evaluates a request against loaded rules (stub).
	EvaluateRequest(ctx context.Context, siteID string, rawRequest []byte) (blocked bool, ruleIDs []string, err error)
	// ReloadRules reloads CRS / custom rules (stub).
	ReloadRules(ctx context.Context) error
}

// NoopEngine is a no-op Coraza adapter for Phase 1.
type NoopEngine struct{}

func NewNoop() *NoopEngine { return &NoopEngine{} }

func (n *NoopEngine) EvaluateRequest(ctx context.Context, siteID string, rawRequest []byte) (bool, []string, error) {
	return false, nil, nil
}

func (n *NoopEngine) ReloadRules(ctx context.Context) error { return nil }
