// Package interfaces defines the port interfaces that use-cases depend on.
// Adapters must implement these interfaces; use-cases must never import concrete adapter packages.
package interfaces

import (
	"context"

	"backend/internal/domain"
)

// Agent is the port for AI-powered CV parsing.
// Implementations are expected to populate all structured fields of the provided CV in-place.
type Agent interface {
	FillOutCv(ctx context.Context, cv *domain.CV) error
}
