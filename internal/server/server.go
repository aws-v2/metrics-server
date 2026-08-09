package server

import (
	"fmt"
	"log/slog"
	router "metrics-gateway/internal/transport"
	"metrics-gateway/internal/transport/handler"
	"metrics-gateway/internal/transport/middleware"
	"net/http"
)

// Config holds the HTTP server configuration.
type Config struct {
	Port int
}

// New creates and returns a configured *http.Server with all routes registered.
func New(
	cfg Config,
	logger *slog.Logger,
	auth *middleware.AuthMiddleware,
	healthHandler *handler.HealthHandler,
	ec2Handler *handler.EC2Handler,
	rdsHandler *handler.RDSHandler,
	lambdaHandler *handler.LambdaHandler,
	s3Handler *handler.S3Handler,
	sagemakerHandler *handler.SageMakerHandler,
	gameliftHandler *handler.GameLiftHandler,
	docsHandler *handler.DocsHandler,

) *http.Server {
	mux := http.NewServeMux()

	// Register all routes through the single entry point.
	router.RegisterRoutes(mux, auth, healthHandler, ec2Handler, rdsHandler, lambdaHandler, s3Handler, sagemakerHandler, gameliftHandler, docsHandler)

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("configuring HTTP server", "addr", addr)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
