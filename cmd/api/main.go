package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"metrics-gateway/application"
	"metrics-gateway/application/scaling"
	"metrics-gateway/internal/config"
	"metrics-gateway/internal/database"
	"metrics-gateway/internal/database/postgres"
	"metrics-gateway/internal/messaging"
	"metrics-gateway/internal/server"
	"metrics-gateway/internal/transport"
	"metrics-gateway/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// ── 1. Load Configuration ────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Inject profile into logger for all subsequent logs
	logger = logger.With("profile", cfg.AppProfile)

	if cfg.Server.InstanceTokenSecret == "" {
		log.Fatal("INSTANCE_TOKEN_SECRET environment variable must be set")
	}

	// ── 2. Initialize NATS ───────────────────────────────────────────────────
	natsHost, natsPort := cfg.NATS.HostPort()
	if err := config.CheckReachability(natsHost, natsPort, "NATS", cfg.AppProfile); err != nil {
		logger.Error("NATS reachability check failed", "error", err)
		os.Exit(1)
	}

	logger.Info("connecting to NATS")
	natsClient, err := messaging.NewClient(cfg.NATS.URL, cfg.NATS.User, cfg.NATS.Password)
	if err != nil {
		logger.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer natsClient.Close()

	// Example: build a subject following the norms
	_ = messaging.BuildSubject(cfg.AppProfile, messaging.ServiceName, messaging.Version, "metric", "ingested")

	// ── 3. Initialize PostgreSQL ─────────────────────────────────────────────
	logger.Info("initializing PostgreSQL repository")
	if err := config.CheckReachability(cfg.DB.Host, cfg.DB.Port, "PostgreSQL", cfg.AppProfile); err != nil {
		logger.Error("database reachability check failed", "error", err)
		os.Exit(1)
	}

	postgresConn := cfg.DB.ConnectionString()
	db, err := database.NewPostgresDB(postgresConn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("running database migrations")
	if err := database.Migrate(db, postgres.Schema); err != nil {
		logger.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}
	logger.Info("database migration completed successfully")

	// ── 4. Initialize Layers ─────────────────────────────────────────────────
	repo := repository.NewPostgresRepository(db)

	// Auth middleware
	auth := transport.NewAuthMiddleware(logger, cfg.Server.InstanceTokenSecret)

	// Health
	healthService := application.NewHealthService(logger, repo)
	healthHandler := transport.NewHealthHandler(logger, healthService)

	// EC2
	ec2Service := application.NewEC2Service(logger, repo)
	ec2Handler := transport.NewEC2Handler(logger, ec2Service)

	// RDS
	rdsService := application.NewRDSService(logger, repo)
	rdsHandler := transport.NewRDSHandler(logger, rdsService)

	// Lambda
	lambdaService := application.NewLambdaService(logger, repo)
	lambdaHandler := transport.NewLambdaHandler(logger, lambdaService)

	// S3
	s3Service := application.NewS3Service(logger, repo)
	s3Handler := transport.NewS3Handler(logger, s3Service)

	// SageMaker (uncomment when SageMaker service is ready)
	// sagemakerService := application.NewSageMakerService(logger, repo)
	// sagemakerHandler := transport.NewSageMakerHandler(logger, sagemakerService)

	// GameLift (uncomment when GameLift service is ready)
	// gameliftService := application.NewGameLiftService(logger, repo)
	// gameliftHandler := transport.NewGameLiftHandler(logger, gameliftService)

	// ── 5. Initialize Billing Service ────────────────────────────────────────
	billingService := application.NewBillingService(logger, repo)
	billingNATSHandler := transport.NewBillingNATSHandler(logger, natsClient.Conn, billingService, cfg.AppProfile)
	if err := billingNATSHandler.Start(); err != nil {
		logger.Error("failed to start billing NATS handler", "error", err)
	}

	// ── 6. Initialize Scaling Modules ────────────────────────────────────────
	ec2Scaler := scaling.NewEC2Scaler(logger, natsClient.Conn, repo, cfg.AppProfile)
	rdsScaler := scaling.NewRDSScaler(logger, natsClient.Conn, repo, cfg.AppProfile)
	lambdaScaler := scaling.NewLambdaScaler(logger, natsClient.Conn, repo, cfg.AppProfile)
	s3Scaler := scaling.NewS3Scaler(logger, natsClient.Conn, repo, cfg.AppProfile)

	ec2Scaler.Start()
	rdsScaler.Start()
	lambdaScaler.Start()
	s3Scaler.Start()
	logger.Info("all scaling background workers started")

	scalingPolicyHandler := transport.NewScalingPolicyNATSHandler(logger, natsClient.Conn, repo, cfg.AppProfile)
	if err := scalingPolicyHandler.Start(); err != nil {
		logger.Error("failed to start scaling policy NATS handler", "error", err)
	}

	// ── 7. Start HTTP Server ─────────────────────────────────────────────────
	srv := server.New(server.Config{
		Port: cfg.Server.HTTPPort,
	}, logger, auth, healthHandler, ec2Handler, rdsHandler, lambdaHandler, s3Handler)

	logger.Info("metrics-gateway starting",
		"port", cfg.Server.HTTPPort,
		"service", cfg.Server.ServiceName,
	)

	fmt.Printf("metrics-gateway listening on :%d\n", cfg.Server.HTTPPort)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}
