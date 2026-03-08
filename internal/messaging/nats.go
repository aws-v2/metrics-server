package messaging

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// ─── Subject naming norms ────────────────────────────────────────────────────
// Pattern: <env>.<service>.<version>.<domain>.<action_type>
//
//   Events  (past tense) : dev.metrics.v1.metric.ingested
//   Commands (imperative) : dev.metrics.v1.metric.ingest
//
// ─── ARN norms ───────────────────────────────────────────────────────────────
// arn:serw:<service>:<region>:<account-id>:<resource-type>/<resource-id>
// Example: arn:serw:metrics:eu-north-1:user-123:metric/cpu-usage-001

const (
	ServiceName = "metrics"
	Version     = "v1"
)

// BuildSubject constructs a NATS subject following the naming convention:
//
//	<env>.<service>.<version>.<domain>.<action>
func BuildSubject(env, service, version, domain, action string) string {
	return strings.Join([]string{env, service, version, domain, action}, ".")
}

// Client wraps a NATS connection for the metrics-gateway service.
type Client struct {
	Conn *nats.Conn
}

// NewClient connects to NATS with optional user/password credentials.
func NewClient(url, user, password string) (*Client, error) {
	opts := []nats.Option{
		nats.Name("metrics-gateway"),
		nats.Timeout(5 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(60),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.Printf("[NATS] disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("[NATS] reconnected to %s", nc.ConnectedUrl())
		}),
	}

	if user != "" && password != "" {
		opts = append(opts, nats.UserInfo(user, password))
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS at %s: %w", url, err)
	}

	log.Printf("[NATS] connected to %s", nc.ConnectedUrl())
	return &Client{Conn: nc}, nil
}

// Close drains and closes the NATS connection gracefully.
func (c *Client) Close() {
	if c.Conn != nil {
		_ = c.Conn.Drain()
		c.Conn.Close()
		log.Println("[NATS] connection closed")
	}
}
