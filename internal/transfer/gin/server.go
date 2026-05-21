package gintransfer

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"backend/internal/transfer/gin/handlers"
	"backend/pkg/log"
)

type Server struct {
	router *gin.Engine
	log    log.Logger
}

func NewServer(uploadCV *handlers.UploadCVHandler, logger log.Logger) *Server {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		cv := v1.Group("/cv")
		{
			cv.POST("/upload", uploadCV.Handle)
		}
	}

	return &Server{
		router: router,
		log:    logger,
	}
}

func (s *Server) Run(addr string) error {
	s.log.Info("starting HTTP server", log.FieldLogger{Key: "addr", Value: addr})
	return s.router.Run(addr)
}
