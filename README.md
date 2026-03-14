# metrics-gateway

Skeleton HTTP server for the metrics ingestion gateway.

## Norms

### NATS Subject Pattern
```
<env>.<service>.<version>.<domain>.<action_type>
```
- **Events** (past tense): `dev.metrics.v1.metric.ingested`
- **Commands** (imperative): `dev.metrics.v1.metric.ingest`

### ARN Format
```
arn:serw:<service>:<region>:<account-id>:<resource-type>/<resource-id>
```
Example: `arn:serw:metrics:eu-north-1:user-123:metric/cpu-001`

## Build

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o metrics-gateway ./cmd/api
```

## Run

```bash
go run cmd/api/main.go
```

The server starts on **port 8085** by default.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `root` | Database user |
| `DB_PASSWORD` | `root` | Database password |
| `DB_NAME` | `metrics_db` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `NATS_URL` | `nats://auth-server:auth-secret@localhost:4222` | NATS server URL |
| `HTTP_PORT` | `8085` | HTTP server port |

## Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Returns `{"status":"ok"}` with 200 OK |

## Project Structure

```
metrics-gateway/
  cmd/api/main.go                        # Entry point
  go.mod
  internal/
    config/config.go                     # Environment-based configuration
    database/
      postgres.go                        # PostgreSQL connection & migration
      postgres/schema.go                 # SQL schema definitions
    messaging/nats.go                    # NATS client & subject builder
    server/server.go                     # HTTP server setup
    transport/
      routes.go                          # All route registration
      health_handler.go                  # /health endpoint handler
  application/
    interfaces.go                        # All service & repository interfaces
    health_service.go                    # Health service implementation
  repository/
    repository.go                        # PostgreSQL repository implementation
```
