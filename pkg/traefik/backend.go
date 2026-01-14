package traefik

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/tazhate/k8s-internal-loadbalancer/pkg/circuitbreaker"
	"github.com/tazhate/k8s-internal-loadbalancer/pkg/config"
)

// Backend manages Traefik backend configuration
type Backend struct {
	apiURL         string
	routerName     string
	serviceName    string
	lbMethod       string
	client         *http.Client
	circuitBreaker *circuitbreaker.CircuitBreaker
}

// New creates a new Traefik backend manager
func New(cfg *config.Config) *Backend {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	cb := circuitbreaker.New(
		cfg.CBMaxRequests,
		cfg.CBInterval,
		cfg.CBTimeout,
		cfg.CBConsecutiveFailures,
	)

	return &Backend{
		apiURL:         cfg.TraefikAPIURL,
		routerName:     cfg.RouterName,
		serviceName:    cfg.ServiceName,
		lbMethod:       cfg.LoadBalancerMethod,
		client:         client,
		circuitBreaker: cb,
	}
}

// UpdateBackends updates the Traefik backend servers
func (b *Backend) UpdateBackends(ctx context.Context, backends []string) error {
	return b.circuitBreaker.Execute(func() error {
		return b.updateBackendsInternal(ctx, backends)
	})
}

// updateBackendsInternal performs the actual backend update
func (b *Backend) updateBackendsInternal(ctx context.Context, backends []string) error {
	// Build servers slice
	servers := make([]map[string]string, 0, len(backends))
	for _, backend := range backends {
		servers = append(servers, map[string]string{
			"address": backend,
		})
	}

	// Build configuration
	config := map[string]any{
		"tcp": map[string]any{
			"routers": map[string]any{
				b.routerName: map[string]any{
					"entryPoints": []string{"tcp"},
					"rule":        "HostSNI(`*`)",
					"service":     b.serviceName,
				},
			},
			"services": map[string]any{
				b.serviceName: map[string]any{
					"loadBalancer": map[string]any{
						"method":  b.lbMethod,
						"servers": servers,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", b.apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	slog.Info("Updated Traefik configuration",
		"backend_count", len(backends),
		"circuit_breaker_state", b.circuitBreaker.State())

	return nil
}

// HealthCheck checks if Traefik API is accessible
func (b *Backend) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", b.apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMethodNotAllowed {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// CircuitBreakerStats returns circuit breaker statistics
func (b *Backend) CircuitBreakerStats() map[string]interface{} {
	return b.circuitBreaker.Stats()
}
