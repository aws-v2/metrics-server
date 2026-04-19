package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

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

	_ = messaging.BuildSubject(cfg.AppProfile, messaging.ServiceName, messaging.Version, "metric", "ingested")

	// ── 3. Initialize PostgreSQL ─────────────────────────────────────────────
	logger.Info("initializing PostgreSQL repository")
	if err := config.CheckReachability(cfg.DB.Host, cfg.DB.Port, "PostgreSQL", cfg.AppProfile); err != nil {
		logger.Error("database reachability check failed", "error", err)
		os.Exit(1)
	}

	db, err := database.NewPostgresDB(cfg.DB.ConnectionString())
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

	auth := transport.NewAuthMiddleware(logger, cfg.Server.InstanceTokenSecret)

	healthService := application.NewHealthService(logger, repo)
	healthHandler := transport.NewHealthHandler(logger, healthService)

	ec2Service := application.NewEC2Service(logger, repo)
	ec2Handler := transport.NewEC2Handler(logger, ec2Service)

	rdsService := application.NewRDSService(logger, repo)
	rdsHandler := transport.NewRDSHandler(logger, rdsService)

	lambdaService := application.NewLambdaService(logger, repo)
	lambdaHandler := transport.NewLambdaHandler(logger, lambdaService)

	s3Service := application.NewS3Service(logger, repo)
	s3Handler := transport.NewS3Handler(logger, s3Service)

	sagemakerService := application.NewSageMakerService(logger, repo)
	sagemakerHandler := transport.NewSageMakerHandler(logger, sagemakerService)

	gameliftService := application.NewGameLiftService(logger, repo)
	gameliftHandler := transport.NewGameLiftHandler(logger, gameliftService)

	docsService := application.NewDocsService("./docs")
	docsHandler := transport.NewDocsHandler(docsService)

	// ── 5. Initialize Billing ────────────────────────────────────────────────
	billingService := application.NewBillingService(logger, repo)
	billingNATSHandler := transport.NewBillingNATSHandler(logger, natsClient.Conn, billingService, cfg.AppProfile)
	if err := billingNATSHandler.Start(); err != nil {
		logger.Error("failed to start billing NATS handler", "error", err)
	}

	// ── 6. Initialize Scaling ────────────────────────────────────────────────
	ec2Scaler := scaling.NewEC2Scaler(logger, natsClient.Conn, repo, cfg.AppProfile)
	rdsScaler := scaling.NewRDSScaler(logger, natsClient.Conn, repo, cfg.AppProfile)
	lambdaScaler := scaling.NewLambdaScaler(logger, natsClient.Conn, repo, cfg.AppProfile)
	s3Scaler := scaling.NewS3Scaler(logger, natsClient.Conn, repo, cfg.AppProfile)
	sagemakerScaler := scaling.NewSageMakerScaler(logger, natsClient.Conn, repo, cfg.AppProfile)
	gameliftScaler := scaling.NewGameLiftScaler(logger, natsClient.Conn, repo, cfg.AppProfile)

	ec2Scaler.Start()
	rdsScaler.Start()
	lambdaScaler.Start()
	s3Scaler.Start()
	sagemakerScaler.Start()
	gameliftScaler.Start()
	logger.Info("all scaling background workers started")

	scalingPolicyHandler := transport.NewScalingPolicyNATSHandler(logger, natsClient.Conn, repo, cfg.AppProfile)
	if err := scalingPolicyHandler.Start(); err != nil {
		logger.Error("failed to start scaling policy NATS handler", "error", err)
	}

	// ── 7. Register with Eureka ──────────────────────────────────────────────
	for i := 1; i <= 3; i++ {
		if err := registerWithEureka(cfg.Eureka); err != nil {
			logger.Warn("Eureka registration attempt failed",
				"attempt", i,
				"error", err,
			)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}

	go sendHeartbeat(logger, cfg.Eureka)

	// ── 8. Start HTTP Server ─────────────────────────────────────────────────
	srv := server.New(server.Config{
		Port: cfg.Server.HTTPPort,
	}, logger, auth, healthHandler, ec2Handler, rdsHandler, lambdaHandler, s3Handler, sagemakerHandler, gameliftHandler, docsHandler)

	logger.Info("metrics-gateway starting",
		"port", cfg.Server.HTTPPort,
		"service", cfg.Server.ServiceName,
	)

	// Deregister from Eureka on exit
	defer func() {
		if err := deregisterFromEureka(cfg.Eureka); err != nil {
			logger.Warn("failed to deregister from Eureka", "error", err)
		}
	}()

	fmt.Printf("metrics-gateway listening on :%d\n", cfg.Server.HTTPPort)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}

// ── Eureka helpers ────────────────────────────────────────────────────────────

func registerWithEureka(cfg config.EurekaConfig) error {
	instance := map[string]any{
		"instance": map[string]any{
			"instanceId": cfg.InstanceID,
			"hostName":   cfg.HostName,
			"app":        cfg.AppName,
			"ipAddr":     cfg.IPAddr,
			"vipAddress": cfg.VipAddress,
			"status":     "UP",
			"port": map[string]any{
				"$":        cfg.Port,
				"@enabled": "true",
			},
			"dataCenterInfo": map[string]any{
				"@class": "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
				"name":   "MyOwn",
			},
			"healthCheckUrl": fmt.Sprintf("http://%s:%d/health", cfg.HostName, cfg.Port),
			"statusPageUrl":  fmt.Sprintf("http://%s:%d/health", cfg.HostName, cfg.Port),
			"homePageUrl":    fmt.Sprintf("http://%s:%d/", cfg.HostName, cfg.Port),
		},
	}

	jsonData, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal registration payload: %w", err)
	}

	eurekaURL := fmt.Sprintf("%s/apps/%s", cfg.ServerURL, cfg.AppName)
	req, err := http.NewRequest(http.MethodPost, eurekaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach Eureka: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("eureka returned status %d", resp.StatusCode)
	}

	slog.Info("registered with Eureka", "url", eurekaURL, "instance", cfg.InstanceID)
	return nil
}

func sendHeartbeat(logger *slog.Logger, cfg config.EurekaConfig) {
	ticker := time.NewTicker(cfg.HeartbeatInterval)
	defer ticker.Stop()

	eurekaURL := fmt.Sprintf("%s/apps/%s/%s", cfg.ServerURL, cfg.AppName, cfg.InstanceID)
	client := &http.Client{Timeout: 5 * time.Second}

	for range ticker.C {
		req, err := http.NewRequest(http.MethodPut, eurekaURL, nil)
		if err != nil {
			logger.Error("failed to build heartbeat request", "error", err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("heartbeat failed", "error", err)
			continue
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			logger.Warn("heartbeat rejected", "status", resp.StatusCode)
		} else {
			logger.Debug("heartbeat sent", "instance", cfg.InstanceID)
		}
		resp.Body.Close()
	}
}

func deregisterFromEureka(cfg config.EurekaConfig) error {
	eurekaURL := fmt.Sprintf("%s/apps/%s/%s", cfg.ServerURL, cfg.AppName, cfg.InstanceID)
	req, err := http.NewRequest(http.MethodDelete, eurekaURL, nil)
	if err != nil {
		return fmt.Errorf("failed to build deregistration request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach Eureka: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("deregistration returned status %d", resp.StatusCode)
	}

	slog.Info("deregistered from Eureka", "instance", cfg.InstanceID)
	return nil
}