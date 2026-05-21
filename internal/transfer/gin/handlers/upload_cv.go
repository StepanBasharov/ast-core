package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	usecase "backend/internal/use-case"
	"backend/pkg/log"
)

type UploadCVHandler struct {
	uploadCVUseCase *usecase.UploadCVUseCase
	log             log.Logger
}

func NewUploadCVHandler(uploadCVUseCase *usecase.UploadCVUseCase, logger log.Logger) *UploadCVHandler {
	return &UploadCVHandler{
		uploadCVUseCase: uploadCVUseCase,
		log:             logger,
	}
}

// Handle godoc
// @Summary      Upload a CV
// @Description  Accepts a PDF file, extracts text, parses structured data via Claude AI, and saves to the database.
// @Tags         cv
// @Accept       multipart/form-data
// @Produce      json
// @Param        cv   formData  file  true  "PDF file of the candidate's CV"
// @Success      201  {object}  map[string]string  "CV uploaded successfully"
// @Failure      400  {object}  map[string]string  "cv file is required"
// @Failure      422  {object}  map[string]string  "failed to process CV"
// @Router       /cv/upload [post]
func (h *UploadCVHandler) Handle(ctx *gin.Context) {
	file, _, err := ctx.Request.FormFile("cv")
	if err != nil {
		h.log.Error("failed to read uploaded file",
			log.FieldLogger{Key: "err", Value: err},
		)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cv file is required"})
		return
	}
	defer file.Close()

	if err = h.uploadCVUseCase.Execute(ctx.Request.Context(), file); err != nil {
		if errors.Is(err, usecase.ErrUploadCV) {
			h.log.Error("failed to process CV upload",
				log.FieldLogger{Key: "err", Value: err},
			)
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to process CV"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "CV uploaded successfully"})
}
