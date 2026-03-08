package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"metrics-gateway/internal/transport"
)

// Config holds the HTTP server configuration.
type Config struct {
	Port int
}

// New creates and returns a configured *http.Server with all routes registered.
func New(
	cfg Config,
	logger *slog.Logger,
	auth *transport.AuthMiddleware,
	healthHandler *transport.HealthHandler,
	ec2Handler *transport.EC2Handler,
	rdsHandler *transport.RDSHandler,
	lambdaHandler *transport.LambdaHandler,
	s3Handler *transport.S3Handler,
	// sagemakerHandler *transport.SageMakerHandler, // TODO: uncomment when SageMaker service is ready
	// gameliftHandler  *transport.GameLiftHandler,  // TODO: uncomment when GameLift service is ready
) *http.Server {
	mux := http.NewServeMux()

	// Register all routes through the single entry point.
	transport.RegisterRoutes(mux, auth, healthHandler, ec2Handler, rdsHandler, lambdaHandler, s3Handler)

	addr := fmt.Sprintf(":%d", cfg.Port)
	logger.Info("configuring HTTP server", "addr", addr)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
