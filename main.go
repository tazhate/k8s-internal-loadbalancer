package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tazhate/k8s-internal-loadbalancer/pkg/config"
	"github.com/tazhate/k8s-internal-loadbalancer/pkg/health"
	"github.com/tazhate/k8s-internal-loadbalancer/pkg/podwatcher"
	"github.com/tazhate/k8s-internal-loadbalancer/pkg/traefik"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Version information (set via ldflags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	VCSRef    = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Set version info
	cfg.Version = Version
	cfg.BuildTime = BuildTime
	cfg.VCSRef = VCSRef

	// Initialize logger
	initLogger(cfg)

	slog.Info("Starting K8s Internal Load Balancer",
		"version", Version,
		"build_time", BuildTime,
		"vcs_ref", VCSRef,
		"use_watch", cfg.UseWatch)

	// Validate configuration
	if validateErr := cfg.Validate(); validateErr != nil {
		slog.Error("Invalid configuration", "error", validateErr)
		os.Exit(1)
	}

	// Create Kubernetes client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		slog.Error("Failed to create Kubernetes config", "error", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		slog.Error("Failed to create Kubernetes clientset", "error", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create components
	traefikBackend := traefik.New(cfg)
	watcher := podwatcher.New(
		clientset,
		cfg.PodNamespace,
		cfg.PodLabels,
		cfg.BackendPort,
		cfg.UpdateInterval,
		cfg.UseWatch,
	)

	// Create health check server
	healthServer := health.NewServer(cfg.HealthCheckPort)

	// Add health checkers
	healthServer.AddChecker(health.NewKubernetesHealthChecker(func(ctx context.Context) error {
		_, err := clientset.CoreV1().Pods(cfg.PodNamespace).List(ctx, metav1.ListOptions{Limit: 1})
		return err
	}))
	healthServer.AddChecker(health.NewTraefikHealthChecker(traefikBackend))

	// Start health server
	go func() {
		if err := healthServer.Start(ctx); err != nil {
			slog.Error("Health server error", "error", err)
		}
	}()

	// Start watching pods
	backendsChan, errorsChan := watcher.Watch(ctx)

	// Mark as ready after initial setup
	healthServer.SetReady(true)
	slog.Info("Application ready",
		"namespace", cfg.PodNamespace,
		"labels", cfg.PodLabels,
		"health_port", cfg.HealthCheckPort)

	// Main event loop
	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down gracefully...")
			if err := watcher.Close(); err != nil {
				slog.Error("Error closing watcher", "error", err)
			}
			return

		case backends, ok := <-backendsChan:
			if !ok {
				slog.Info("Backends channel closed")
				return
			}
			if err := traefikBackend.UpdateBackends(ctx, backends); err != nil {
				slog.Error("Failed to update backends", "error", err,
					"circuit_breaker_state", traefikBackend.CircuitBreakerStats()["state"])
			}

		case err, ok := <-errorsChan:
			if !ok {
				slog.Info("Errors channel closed")
				return
			}
			slog.Error("Watcher error", "error", err)
		}
	}
}

// initLogger initializes the structured logger
func initLogger(cfg *config.Config) {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
