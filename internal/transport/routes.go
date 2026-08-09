package router

import (
	"metrics-gateway/internal/transport/handler"
	"metrics-gateway/internal/transport/middleware"
	"net/http"
)

// prefix is the versioned API base path.
const prefix = "/api/v1/metrics"

// RegisterRoutes is the single source of truth for all HTTP route registration.
// Every route group gets its own handler file; this function wires them to the mux.
// The auth middleware is applied to all routes except health.
func RegisterRoutes(
	mux *http.ServeMux,
	auth *middleware.AuthMiddleware,
	healthHandler *handler.HealthHandler,
	ec2Handler *handler.EC2Handler,
	rdsHandler *handler.RDSHandler,
	lambdaHandler *handler.LambdaHandler,
	s3Handler *handler.S3Handler,
	sagemakerHandler *handler.SageMakerHandler,
	gameliftHandler *handler.GameLiftHandler,
	docsHandler *handler.DocsHandler,

) {
	// ── Health (no auth) ─────────────────────────────────────────────────────
	mux.HandleFunc("GET "+prefix+"/health", healthHandler.Health)

	// ── EC2 Metrics ──────────────────────────────────────────────────────────
	mux.HandleFunc("POST "+prefix+"/ec2/ingest", middleware.Wrap(ec2Handler.Ingest))
	mux.HandleFunc("GET "+prefix+"/ec2/{instanceId}", middleware.Wrap(ec2Handler.GetByInstance))
	mux.HandleFunc("GET "+prefix+"/ec2", middleware.Wrap(ec2Handler.List))

	// ── RDS Metrics ──────────────────────────────────────────────────────────
	mux.HandleFunc("POST "+prefix+"/rds/ingest", middleware.Wrap(rdsHandler.Ingest))
	mux.HandleFunc("GET "+prefix+"/rds/{instanceId}", middleware.Wrap(rdsHandler.GetByInstance))
	mux.HandleFunc("GET "+prefix+"/rds", middleware.Wrap(rdsHandler.List))

	// ── Lambda Metrics ───────────────────────────────────────────────────────
	mux.HandleFunc("POST "+prefix+"/lambda/ingest", middleware.Wrap(lambdaHandler.Ingest))
	mux.HandleFunc("GET "+prefix+"/lambda/{functionId}", middleware.Wrap(lambdaHandler.GetByFunction))
	mux.HandleFunc("GET "+prefix+"/lambda", middleware.Wrap(lambdaHandler.List))

	// ── S3 Metrics ───────────────────────────────────────────────────────────
	mux.HandleFunc("POST "+prefix+"/s3/ingest", middleware.Wrap(s3Handler.Ingest))
	mux.HandleFunc("GET "+prefix+"/s3/{bucketId}", middleware.Wrap(s3Handler.GetByBucket))
	mux.HandleFunc("GET "+prefix+"/s3", middleware.Wrap(s3Handler.List))

	// ── SageMaker Metrics (uncomment when service is ready) ─────────────────
	mux.HandleFunc("POST "+prefix+"/sagemaker/ingest", middleware.Wrap(sagemakerHandler.Ingest))
	mux.HandleFunc("GET "+prefix+"/sagemaker/{endpointId}", middleware.Wrap(sagemakerHandler.GetByEndpoint))
	mux.HandleFunc("GET "+prefix+"/sagemaker", middleware.Wrap(sagemakerHandler.List))

	// ── GameLift Metrics (uncomment when service is ready) ──────────────────
	mux.HandleFunc("POST "+prefix+"/gamelift/ingest", middleware.Wrap(gameliftHandler.Ingest))
	mux.HandleFunc("GET "+prefix+"/gamelift/{fleetId}", middleware.Wrap(gameliftHandler.GetByFleet))
	mux.HandleFunc("GET "+prefix+"/gamelift", middleware.Wrap(gameliftHandler.List))

	// ── Public Docs ───────────────────────────────────────────────────────────
	mux.HandleFunc("GET "+prefix+"/docs", docsHandler.GetManifest)
	mux.HandleFunc("GET "+prefix+"/docs/{slug}", docsHandler.GetDoc)

}
