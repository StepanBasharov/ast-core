package usecase

import (
	"backend/internal/domain"
	"backend/internal/use-case/interfaces"
	"backend/pkg/log"
	"context"

	"github.com/google/uuid"
)

type GetCVUseCase struct {
	cvRepo interfaces.CVRepository
	log    log.Logger
}

func NewGetCVUseCase(cvRepo interfaces.CVRepository, logger log.Logger) *GetCVUseCase {
	return &GetCVUseCase{
		cvRepo: cvRepo,
		log:    logger,
	}
}

func (uc *GetCVUseCase) Execute(ctx context.Context, cvID uuid.UUID) (domain.CV, error) {
	cv, err := uc.cvRepo.GetByID(ctx, cvID)
	if err != nil {
		uc.log.Error("failed to fetch CV",
			log.FieldLogger{Key: "use-case", Value: "GetCVUseCase"},
			log.FieldLogger{Key: "Error", Value: err},
		)
	}

	return cv, err
}
