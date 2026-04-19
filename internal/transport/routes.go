package transport

import "net/http"

// prefix is the versioned API base path.
const prefix = "/api/v1/metrics"


// RegisterRoutes is the single source of truth for all HTTP route registration.
// Every route group gets its own handler file; this function wires them to the mux.
// The auth middleware is applied to all routes except health.
func RegisterRoutes(
	mux *http.ServeMux,
	auth *AuthMiddleware,
	healthHandler *HealthHandler,
	ec2Handler *EC2Handler,
	rdsHandler *RDSHandler,
	lambdaHandler *LambdaHandler,
	s3Handler *S3Handler,
	sagemakerHandler *SageMakerHandler,
	gameliftHandler  *GameLiftHandler,
	docsHandler *DocsHandler,

) {
	// ── Health (no auth) ─────────────────────────────────────────────────────
	mux.HandleFunc("GET "+prefix+"/health", healthHandler.Health)

	// ── EC2 Metrics ──────────────────────────────────────────────────────────
	mux.HandleFunc("POST "+prefix+"/ec2/ingest", auth.Wrap(ec2Handler.Ingest))
	mux.HandleFunc("GET "+prefix+"/ec2/{instanceId}", auth.Wrap(ec2Handler.GetByInstance))
	mux.HandleFunc("GET "+prefix+"/ec2", auth.Wrap(ec2Handler.List))

	// ── RDS Metrics ──────────────────────────────────────────────────────────
	mux.HandleFunc("POST "+prefix+"/rds/ingest", auth.Wrap(rdsHandler.Ingest))
	mux.HandleFunc("GET "+prefix+"/rds/{instanceId}", auth.Wrap(rdsHandler.GetByInstance))
	mux.HandleFunc("GET "+prefix+"/rds", auth.Wrap(rdsHandler.List))

	// ── Lambda Metrics ───────────────────────────────────────────────────────
	mux.HandleFunc("POST "+prefix+"/lambda/ingest", auth.Wrap(lambdaHandler.Ingest))
	mux.HandleFunc("GET "+prefix+"/lambda/{functionId}", auth.Wrap(lambdaHandler.GetByFunction))
	mux.HandleFunc("GET "+prefix+"/lambda", auth.Wrap(lambdaHandler.List))

	// ── S3 Metrics ───────────────────────────────────────────────────────────
	mux.HandleFunc("POST "+prefix+"/s3/ingest", auth.Wrap(s3Handler.Ingest))
	mux.HandleFunc("GET "+prefix+"/s3/{bucketId}", auth.Wrap(s3Handler.GetByBucket))
	mux.HandleFunc("GET "+prefix+"/s3", auth.Wrap(s3Handler.List))

	// ── SageMaker Metrics (uncomment when service is ready) ─────────────────
	mux.HandleFunc("POST "+prefix+"/sagemaker/ingest", auth.Wrap(sagemakerHandler.Ingest))
	mux.HandleFunc("GET "+prefix+"/sagemaker/{endpointId}", auth.Wrap(sagemakerHandler.GetByEndpoint))
	mux.HandleFunc("GET "+prefix+"/sagemaker", auth.Wrap(sagemakerHandler.List))

	// ── GameLift Metrics (uncomment when service is ready) ──────────────────
	mux.HandleFunc("POST "+prefix+"/gamelift/ingest", auth.Wrap(gameliftHandler.Ingest))
	mux.HandleFunc("GET "+prefix+"/gamelift/{fleetId}", auth.Wrap(gameliftHandler.GetByFleet))
	mux.HandleFunc("GET "+prefix+"/gamelift", auth.Wrap(gameliftHandler.List))





	// ── Public Docs ───────────────────────────────────────────────────────────
	mux.HandleFunc("GET "+prefix+"/docs",       docsHandler.GetPublicManifest)
	mux.HandleFunc("GET "+prefix+"/docs/{slug}", docsHandler.GetPublicDoc)

	// ── Internal Docs ─────────────────────────────────────────────────────────
	mux.HandleFunc("GET "+prefix+"/internal/docs",       docsHandler.GetInternalManifest)
	mux.HandleFunc("GET "+prefix+"/internal/docs/{slug}", docsHandler.GetInternalDoc)

}
