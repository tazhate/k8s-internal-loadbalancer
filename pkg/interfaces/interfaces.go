package interfaces

import (
	"context"
)

// PodWatcher watches for pod changes in Kubernetes
type PodWatcher interface {
	// Watch starts watching for pod changes and sends backend addresses to the channel
	Watch(ctx context.Context) (<-chan []string, <-chan error)
	// Close stops the watcher
	Close() error
}

// LoadBalancerBackend manages load balancer backend configuration
type LoadBalancerBackend interface {
	// UpdateBackends updates the backend servers
	UpdateBackends(ctx context.Context, backends []string) error
	// HealthCheck checks if the backend is healthy
	HealthCheck(ctx context.Context) error
}

// HealthChecker provides health check functionality
type HealthChecker interface {
	// Check returns true if the component is healthy
	Check(ctx context.Context) error
	// Name returns the name of the component being checked
	Name() string
}

// CircuitBreaker provides circuit breaker functionality
type CircuitBreaker interface {
	// Execute runs the given function with circuit breaker protection
	Execute(func() error) error
	// State returns the current state of the circuit breaker
	State() string
}
