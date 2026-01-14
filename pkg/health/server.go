package health

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/tazhate/k8s-internal-loadbalancer/pkg/interfaces"
)

// Server provides health and readiness endpoints
type Server struct {
	mu       sync.RWMutex
	checkers []interfaces.HealthChecker
	server   *http.Server
	port     int
	ready    bool
}

// NewServer creates a new health check server
func NewServer(port int) *Server {
	return &Server{
		port:     port,
		checkers: make([]interfaces.HealthChecker, 0),
		ready:    false,
	}
}

// AddChecker adds a health checker
func (s *Server) AddChecker(checker interfaces.HealthChecker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkers = append(s.checkers, checker)
}

// SetReady sets the readiness status
func (s *Server) SetReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = ready
}

// Start starts the health check server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthHandler)
	mux.HandleFunc("/readyz", s.readyHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	slog.Info("Starting health check server", "port", s.port)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Health server shutdown error", "error", err)
		}
	}()

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("health server error: %w", err)
	}

	return nil
}

// healthHandler handles liveness probes
func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Liveness check: just return OK if the server is running
	response := map[string]string{
		"status": "ok",
		"check":  "liveness",
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// readyHandler handles readiness probes
func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	ready := s.ready
	checkers := s.checkers
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if !ready {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "not_ready",
			"reason": "application not initialized",
		})
		return
	}

	// Run all health checkers
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	results := make(map[string]string)
	allHealthy := true

	for _, checker := range checkers {
		if err := checker.Check(ctx); err != nil {
			results[checker.Name()] = fmt.Sprintf("unhealthy: %v", err)
			allHealthy = false
		} else {
			results[checker.Name()] = "healthy"
		}
	}

	response := map[string]interface{}{
		"status": "ok",
		"checks": results,
	}

	if !allHealthy {
		response["status"] = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(response)
}

// metricsHandler provides basic metrics
func (s *Server) metricsHandler(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	checkers := s.checkers
	ready := s.ready
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	metrics := map[string]interface{}{
		"ready":          ready,
		"checkers_count": len(checkers),
		"uptime_seconds": time.Now().Unix(), // simplified
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(metrics)
}

// KubernetesHealthChecker checks Kubernetes API connectivity
type KubernetesHealthChecker struct {
	checkFunc func(context.Context) error
}

// NewKubernetesHealthChecker creates a new Kubernetes health checker
func NewKubernetesHealthChecker(checkFunc func(context.Context) error) *KubernetesHealthChecker {
	return &KubernetesHealthChecker{
		checkFunc: checkFunc,
	}
}

// Check performs the health check
func (k *KubernetesHealthChecker) Check(ctx context.Context) error {
	return k.checkFunc(ctx)
}

// Name returns the name of the checker
func (k *KubernetesHealthChecker) Name() string {
	return "kubernetes_api"
}

// TraefikHealthChecker checks Traefik API connectivity
type TraefikHealthChecker struct {
	backend interface {
		HealthCheck(context.Context) error
	}
}

// NewTraefikHealthChecker creates a new Traefik health checker
func NewTraefikHealthChecker(backend interface{ HealthCheck(context.Context) error }) *TraefikHealthChecker {
	return &TraefikHealthChecker{
		backend: backend,
	}
}

// Check performs the health check
func (t *TraefikHealthChecker) Check(ctx context.Context) error {
	return t.backend.HealthCheck(ctx)
}

// Name returns the name of the checker
func (t *TraefikHealthChecker) Name() string {
	return "traefik_api"
}
