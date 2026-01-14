package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

// Config holds the application configuration
type Config struct {
	// Kubernetes configuration
	PodLabels    string
	PodNamespace string

	// Traefik configuration
	TraefikAPIURL      string
	LoadBalancerMethod string
	RouterName         string
	ServiceName        string

	// Health check configuration
	HealthCheckPath string

	// Logging
	LogLevel  string
	LogFormat string // json or text

	// Version information
	Version   string
	BuildTime string
	VCSRef    string

	// Update configuration
	UpdateInterval time.Duration

	// Circuit breaker configuration
	CBInterval time.Duration
	CBTimeout  time.Duration

	// Traefik and health check ports
	BackendPort     int
	HealthCheckPort int

	// Circuit breaker thresholds
	CBMaxRequests         uint32
	CBConsecutiveFailures uint32

	// Update configuration
	UseWatch bool // Use Kubernetes watch API instead of polling
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		// Defaults
		BackendPort:           3333,
		LoadBalancerMethod:    "leastconn",
		RouterName:            "relay-router",
		ServiceName:           "relay-service",
		UpdateInterval:        time.Second,
		UseWatch:              true,
		HealthCheckPort:       8081,
		HealthCheckPath:       "/health",
		CBMaxRequests:         5,
		CBInterval:            time.Minute,
		CBTimeout:             30 * time.Second,
		CBConsecutiveFailures: 5,
		LogLevel:              "info",
		LogFormat:             "json",
	}

	// Required fields
	cfg.PodLabels = os.Getenv("POD_LABELS")
	if cfg.PodLabels == "" {
		return nil, fmt.Errorf("POD_LABELS environment variable is required")
	}

	cfg.TraefikAPIURL = os.Getenv("TRAEFIK_API_URL")
	if cfg.TraefikAPIURL == "" {
		return nil, fmt.Errorf("TRAEFIK_API_URL environment variable is required")
	}

	// Validate Traefik API URL
	if _, err := url.Parse(cfg.TraefikAPIURL); err != nil {
		return nil, fmt.Errorf("invalid TRAEFIK_API_URL: %w", err)
	}

	// Namespace
	cfg.PodNamespace = os.Getenv("POD_NAMESPACE")
	if cfg.PodNamespace == "" {
		// Try to read from service account
		namespaceBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err == nil {
			cfg.PodNamespace = strings.TrimSpace(string(namespaceBytes))
		}
		if cfg.PodNamespace == "" {
			return nil, fmt.Errorf("POD_NAMESPACE could not be determined")
		}
	}

	// Optional: Update interval
	if intervalStr := os.Getenv("UPDATE_INTERVAL"); intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UPDATE_INTERVAL: %w", err)
		}
		cfg.UpdateInterval = interval
	}

	// Optional: Use watch API
	if useWatchStr := os.Getenv("USE_WATCH"); useWatchStr != "" {
		cfg.UseWatch = useWatchStr == "true" || useWatchStr == "1"
	}

	// Optional: Backend port
	if portStr := os.Getenv("BACKEND_PORT"); portStr != "" {
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
			return nil, fmt.Errorf("invalid BACKEND_PORT: %w", err)
		}
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("BACKEND_PORT must be between 1 and 65535")
		}
		cfg.BackendPort = port
	}

	// Optional: Load balancer method
	if method := os.Getenv("LB_METHOD"); method != "" {
		cfg.LoadBalancerMethod = method
	}

	// Optional: Log level
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.LogLevel = strings.ToLower(level)
	}

	// Optional: Log format
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.LogFormat = strings.ToLower(format)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.PodLabels == "" {
		return fmt.Errorf("PodLabels is required")
	}
	if c.TraefikAPIURL == "" {
		return fmt.Errorf("TraefikAPIURL is required")
	}
	if c.PodNamespace == "" {
		return fmt.Errorf("PodNamespace is required")
	}
	if c.BackendPort < 1 || c.BackendPort > 65535 {
		return fmt.Errorf("BackendPort must be between 1 and 65535")
	}
	if c.UpdateInterval < time.Second {
		return fmt.Errorf("UpdateInterval must be at least 1 second")
	}
	return nil
}
