package config

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DB         DBConfig
	NATS       NATSConfig
	Server     ServerConfig
	Eureka     EurekaConfig
	AppProfile string
}

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type NATSConfig struct {
	URL      string
	User     string
	Password string
	NatsPrefix string
}

func (n NATSConfig) HostPort() (string, int) {
	rem := n.URL
	if i := strings.Index(rem, "://"); i != -1 {
		rem = rem[i+3:]
	}
	if i := strings.Index(rem, "@"); i != -1 {
		rem = rem[i+1:]
	}
	parts := strings.Split(rem, ":")
	host := parts[0]
	port := 4222
	if len(parts) > 1 {
		if p, err := strconv.Atoi(parts[1]); err == nil {
			port = p
		}
	}
	return host, port
}

type ServerConfig struct {
	Port                string
	ServiceName         string
	HTTPPort            int
	InstanceTokenSecret string
}

type EurekaConfig struct {
	ServerURL         string
	AppName           string
	HostName          string
	IPAddr            string
	Port              int
	VipAddress        string
	InstanceID        string
	HeartbeatInterval time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	appProfile := getEnv("APP_PROFILE", "dev")
	log.Printf("Loading configuration for profile: %s", appProfile)

	natsUser := getEnv("NATS_USER", "")
	natsPassword := getEnv("NATS_PASSWORD", "")

	if (appProfile == "dev" || appProfile == "staging") && natsUser == "" {
		natsUser = "auth-server"
		natsPassword = "auth-secret"
	}

	httpPort := getEnvInt("HTTP_PORT", 8085)

	cfg := &Config{
		AppProfile: appProfile,
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "root"),
			Password:        getEnv("DB_PASSWORD", "root"),
			Database:        getEnv("DB_NAME", "metrics_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
		},
		NATS: NATSConfig{
			URL:      getEnv("NATS_URL", "nats://localhost:4222"),
			User:     natsUser,
			Password: natsPassword,
			NatsPrefix: getEnv("NATS_PREFIX","dev.v1"),
		},
		Server: ServerConfig{
			Port:                getEnv("PORT", "8085"),
			ServiceName:         getEnv("SERVICE_NAME", "metrics-gateway"),
			HTTPPort:            httpPort,
			InstanceTokenSecret: getEnv("INSTANCE_TOKEN_SECRET", "6A576E5A7234753778214125442A472D4B6150645367566B5970337336763979"),
		},
		Eureka: EurekaConfig{
			ServerURL:         getEnv("EUREKA_SERVER_URL", "http://localhost:8761/eureka"),
			AppName:           getEnv("EUREKA_APP_NAME", "METRICS-SERVICE"),
			HostName:          getEnv("EUREKA_HOSTNAME", "localhost"),
			IPAddr:            getEnv("EUREKA_IP_ADDR", "127.0.0.1"),
			Port:              httpPort, // reuse the same resolved value
			VipAddress:        getEnv("EUREKA_VIP_ADDRESS", "metrics-service"),
			InstanceID:        getEnv("EUREKA_INSTANCE_ID", fmt.Sprintf("metrics-service:%d", httpPort)),
			HeartbeatInterval: getEnvDuration("EUREKA_HEARTBEAT_INTERVAL", 30*time.Second),
		},
	}

	return cfg, nil
}

func (db DBConfig) ConnectionString() string {
	return "host=" + db.Host +
		" port=" + strconv.Itoa(db.Port) +
		" user=" + db.User +
		" password=" + db.Password +
		" dbname=" + db.Database +
		" sslmode=" + db.SSLMode
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func CheckReachability(host string, port int, serviceName, profile string) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	maxAttempts := 5
	delay := 2 * time.Second

	var err error
	for i := 1; i <= maxAttempts; i++ {
		log.Printf("[%s][%s] Checking reachability of %s (attempt %d/%d)...", profile, serviceName, addr, i, maxAttempts)
		conn, dialErr := net.DialTimeout("tcp", addr, 2*time.Second)
		if dialErr == nil {
			conn.Close()
			log.Printf("[%s][%s] Reachability check successful for %s", profile, serviceName, addr)
			return nil
		}
		err = dialErr
		if i < maxAttempts {
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("FATAL: [%s] %s at %s is not reachable after %d attempts: %w", profile, serviceName, addr, maxAttempts, err)
}