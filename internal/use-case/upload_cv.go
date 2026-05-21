// Package usecase contains the application use-cases.
// Use-cases orchestrate domain logic and depend only on port interfaces, never on concrete adapters.
package usecase

import (
	"context"
	"errors"
	"io"
	"os"

	"backend/internal/domain"
	"backend/internal/use-case/interfaces"
	"backend/pkg/log"
)

// ErrUploadCV is returned when the CV upload pipeline fails at any stage.
var ErrUploadCV = errors.New("error uploading CV")

// UploadCVUseCase orchestrates the full CV ingestion pipeline:
// read PDF bytes → write to temp file → extract text → AI parsing → persist to repository.
type UploadCVUseCase struct {
	cvReader interfaces.CVReader
	cvRepo   interfaces.CVRepository
	agent    interfaces.Agent
	log      log.Logger
}

// NewUploadCVUseCase constructs an UploadCVUseCase with all required dependencies injected.
func NewUploadCVUseCase(cvReader interfaces.CVReader, cvRepo interfaces.CVRepository, agent interfaces.Agent, logger log.Logger) *UploadCVUseCase {
	return &UploadCVUseCase{
		cvReader: cvReader,
		cvRepo:   cvRepo,
		agent:    agent,
		log:      logger,
	}
}

// Execute runs the CV upload pipeline for the given reader (typically a multipart file).
// The PDF bytes are written to a temp file because the PDF reader requires a file path.
// The temp file is always cleaned up via defer regardless of the outcome.
func (uc *UploadCVUseCase) Execute(ctx context.Context, uploadedCV io.Reader) error {
	data, err := io.ReadAll(uploadedCV)
	if err != nil {
		uc.log.Error("failed to read CV from input stream",
			log.FieldLogger{Key: "use-case", Value: "UploadCVUseCase"},
			log.FieldLogger{Key: "Error", Value: err},
		)

		return errors.Join(ErrUploadCV, err)
	}

	tmpFile, err := os.CreateTemp("", "*.tmp")
	if err != nil {
		uc.log.Error("failed to create temporary file",
			log.FieldLogger{Key: "use-case", Value: "UploadCVUseCase"},
			log.FieldLogger{Key: "Error", Value: err},
		)

		return errors.Join(ErrUploadCV, err)
	}

	defer func() {
		if errCloseTmpFile := tmpFile.Close(); errCloseTmpFile != nil {
			uc.log.Error("failed to close temporary file",
				log.FieldLogger{Key: "use-case", Value: "UploadCVUseCase"},
				log.FieldLogger{Key: "Error", Value: errCloseTmpFile},
			)
		}

		if errRemoveTmpFile := os.Remove(tmpFile.Name()); errRemoveTmpFile != nil {
			uc.log.Error("failed to remove temporary file",
				log.FieldLogger{Key: "use-case", Value: "UploadCVUseCase"},
				log.FieldLogger{Key: "Error", Value: errRemoveTmpFile},
			)
		}
	}()

	if _, errWrite := tmpFile.Write(data); errWrite != nil {
		uc.log.Error("failed to write CV data to temporary file",
			log.FieldLogger{Key: "use-case", Value: "UploadCVUseCase"},
			log.FieldLogger{Key: "Error", Value: errWrite},
		)

		return errors.Join(ErrUploadCV, errWrite)
	}

	rawCV, err := uc.cvReader.Read(tmpFile.Name())
	if err != nil {
		uc.log.Error("failed to parse CV from file",
			log.FieldLogger{Key: "use-case", Value: "UploadCVUseCase"},
			log.FieldLogger{Key: "Error", Value: err},
		)

		return errors.Join(ErrUploadCV, err)
	}

	cv := domain.CV{RawText: rawCV}

	if err = uc.agent.FillOutCv(ctx, &cv); err != nil {
		uc.log.Error("failed to fill out cv",
			log.FieldLogger{Key: "use-case", Value: "UploadCVUseCase"},
			log.FieldLogger{Key: "Error", Value: err},
		)

		return errors.Join(ErrUploadCV, err)
	}

	if err = uc.cvRepo.Create(ctx, &cv); err != nil {
		uc.log.Error("failed to save CV to repository",
			log.FieldLogger{Key: "use-case", Value: "UploadCVUseCase"},
			log.FieldLogger{Key: "Error", Value: err},
		)

		return errors.Join(ErrUploadCV, err)
	}

	return nil
}
