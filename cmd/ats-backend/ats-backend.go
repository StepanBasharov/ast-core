package main

import (
	"context"
	"os"
	"sync"

	_ "backend/docs"
	gintransfer "backend/internal/transfer/gin"

	"backend/internal/adapters/claude"
	"backend/internal/adapters/pdfreader"
	"backend/internal/adapters/postgres"
	"backend/internal/transfer/gin/handlers"
	usecase "backend/internal/use-case"
	"backend/pkg/log"

	"github.com/joho/godotenv"
)

// @title           ATS Backend API
// @version         1.0
// @description     Applicant Tracking System — uploads and parses CV files using Claude AI.

// @host      localhost:8080
// @BasePath  /api/v1

func main() {
	var (
		wg     sync.WaitGroup
		logger = log.NewZerologLogger()
	)

	if err := godotenv.Load(); err != nil {
		logger.Warn(".env file not found, using environment variables",
			log.FieldLogger{Key: "Error", Value: err},
		)
	}

	var (
		databaseURL      = os.Getenv("DATABASE_URL")
		apiKey           = os.Getenv("API_KEY")
		anthropicBaseURL = os.Getenv("ANTHROPIC_BASE_URL")
		port             = os.Getenv("PORT")
	)

	if port == "" {
		port = ":8080"
	}

	if err := postgres.RunMigrations(databaseURL); err != nil {
		logger.Error("failed to run migrations", log.FieldLogger{Key: "Error", Value: err})

		os.Exit(1)
	}

	aiAgent, err := claude.NewAgentClaude(apiKey, anthropicBaseURL, logger)
	if err != nil {
		logger.Error("failed to initialize ai agent", log.FieldLogger{Key: "Error", Value: err})

		os.Exit(1)
	}

	db, err := postgres.NewAdapterPostgres(context.Background(), logger, databaseURL)
	if err != nil {
		logger.Error("failed to initialize postgres", log.FieldLogger{Key: "Error", Value: err.Error()})

		os.Exit(1)
	}

	pdfReader := pdfreader.NewPDFReader(logger)

	uploadCVUseCase := usecase.NewUploadCVUseCase(pdfReader, db, aiAgent, logger)

	uploadCVHandler := handlers.NewUploadCVHandler(uploadCVUseCase, logger)

	srv := gintransfer.NewServer(uploadCVHandler, logger)

	wg.Add(1)
	go func() {
		defer wg.Done()

		if errStartSrv := srv.Run(port); errStartSrv != nil {
			logger.Error("failed to start server", log.FieldLogger{Key: "Error", Value: errStartSrv})
		}
	}()

	wg.Wait()
}
