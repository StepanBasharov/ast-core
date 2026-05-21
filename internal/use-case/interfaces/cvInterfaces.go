package interfaces

import (
	"context"

	"backend/internal/domain"

	"github.com/google/uuid"
)

// CVRepository is the port for persisting and retrieving CV entities.
type CVRepository interface {
	Create(ctx context.Context, cv *domain.CV) error
	Update(ctx context.Context, cv *domain.CV) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.CV, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// CVReader is the port for extracting plain text from a CV file.
// Implementations receive a file path and return the raw text content.
type CVReader interface {
	Read(path string) (string, error)
}
