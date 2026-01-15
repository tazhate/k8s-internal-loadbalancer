# Kubernetes Internal Load Balancer

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)

A Kubernetes-native internal load balancer that dynamically discovers pods and configures Traefik to route TCP traffic. This project bridges Kubernetes service discovery with Traefik's flexible routing capabilities to provide automated, real-time load balancing.

## Overview

The K8s Internal Load Balancer automatically:
- Discovers pods based on label selectors
- Updates Traefik configuration via REST API
- Maintains an up-to-date list of backend servers
- Provides dynamic TCP load balancing without manual intervention

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Kubernetes Cluster                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────────────────┐                          │
│  │  Pod with 2 Containers       │                          │
│  ├──────────────────────────────┤                          │
│  │ ┌──────────────────────────┐ │                          │
│  │ │ Updater Container        │ │ ─┐                       │
│  │ │ - Polls K8s API          │ │  │                       │
│  │ │ - Discovers pods         │ │  │ Sidecar Pattern      │
│  │ │ - Updates Traefik        │ │  │                       │
│  │ └──────────────────────────┘ │  │                       │
│  │ ┌──────────────────────────┐ │  │                       │
│  │ │ Traefik Container        │ │ ─┘                       │
│  │ │ - Routes TCP traffic     │ │                          │
│  │ │ - Listens on :3333       │ │                          │
│  │ │ - Exposes metrics        │ │                          │
│  │ └──────────────────────────┘ │                          │
│  └──────────────┬───────────────┘                          │
│                 │                                           │
│  ┌──────────────▼──────────────────┐                       │
│  │  Application Pods (discovered)  │                       │
│  │  - Pod 1: 10.0.0.1:3333         │                       │
│  │  - Pod 2: 10.0.0.2:3333         │                       │
│  │  - Pod N: 10.0.0.N:3333         │                       │
│  └─────────────────────────────────┘                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Features

- **Dynamic Pod Discovery**: Continuously monitors Kubernetes API for pods matching label selectors
- **Automatic Configuration**: Updates Traefik backends automatically when pods are added/removed
- **Smart Change Detection**: Only updates Traefik when backend list actually changes
- **Least Connections Load Balancing**: Uses `leastconn` algorithm for optimal distribution
- **Kubernetes-Native**: Deployed via Helm chart with full RBAC support
- **Prometheus Metrics**: Built-in metrics export from Traefik
- **Production-Ready Security**: Configurable security contexts and non-root execution

## When to Use This

This load balancer is designed for a specific scenario that standard Kubernetes Ingress controllers don't handle well:

**You need TCP load balancing for long-lived connections without modifying your existing Ingress setup.**

### The Problem

Standard Kubernetes Ingress is built for HTTP/HTTPS traffic at Layer 7. But many applications rely on raw TCP connections that:

- Stay open for hours, days, or even weeks (database connections, message queues, game servers)
- Use custom protocols that aren't HTTP-based (MQTT, custom binary protocols, database wire protocols)
- Require connection-level load balancing, not request-level

When you have such a service and want to add load balancing:
- **Option A**: Modify your Ingress controller to handle TCP streams — complex, requires config changes, may affect other services
- **Option B**: Use a dedicated TCP load balancer — this project

### The Solution

This project provides a **self-contained, single-service TCP load balancer** that:

1. **Doesn't touch your Ingress** — runs as a separate deployment
2. **Handles long-lived TCP connections** — uses Traefik's TCP routing with least-connections algorithm
3. **Dynamically discovers backends** — watches Kubernetes pods in real-time via Watch API
4. **Operates at Layer 4** — raw TCP, no protocol assumptions

### Ideal Use Cases

| Scenario | Why This Load Balancer |
|----------|------------------------|
| **Database connection pooling** | PostgreSQL/MySQL connections that stay open for connection pools |
| **Message broker clusters** | RabbitMQ, Kafka, NATS with persistent consumer connections |
| **Real-time services** | WebSocket backends, game servers, chat systems |
| **IoT gateways** | MQTT brokers with thousands of long-lived device connections |
| **Custom TCP protocols** | Proprietary protocols that Ingress can't parse |

### When NOT to Use This

- For HTTP/HTTPS APIs — use standard Ingress
- When you need TLS termination with SNI routing — use Ingress with TLS
- For services that already work with your existing load balancing setup

## Requirements

- Kubernetes 1.20+
- Helm 3.0+
- Go 1.22+ (for building from source)

## Quick Start

### Installation via Helm

1. Clone the repository:
```bash
git clone https://github.com/yourusername/k8s-internal-loadbalancer.git
cd k8s-internal-loadbalancer
```

2. Install the Helm chart:
```bash
helm install my-loadbalancer ./chart \
  --set env.relay=app=my-app \
  --set env.updateinterval=5s
```

3. Verify the deployment:
```bash
kubectl get pods
kubectl logs -f <pod-name> -c ilb
```

### Building from Source

```bash
# Build the binary
make build

# Build Docker image
make docker-build

# Run locally (requires kubeconfig)
./traefik-updater
```

## Configuration

The load balancer is configured via environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `POD_LABELS` | Label selector for pods to discover | - | Yes |
| `TRAEFIK_API_URL` | Traefik REST API endpoint | `http://localhost:8080/api/providers/rest` | Yes |
| `POD_NAMESPACE` | Kubernetes namespace to watch | Current namespace | Yes |
| `UPDATE_INTERVAL` | Poll interval for pod discovery | `1s` | No |

### Helm Values

See `chart/values.yaml` for all available configuration options. Key settings:

```yaml
env:
  # Label selector for target pods
  relay: "app=my-app"

  # Update check interval
  updateinterval: 5s

# Security context (enabled by default)
securityContext:
  runAsNonRoot: true
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

## Usage Examples

### Basic Deployment

Deploy a load balancer for pods with label `app=redis`:

```bash
helm install redis-lb ./chart \
  --set env.relay=app=redis \
  --set env.updateinterval=10s
```

### Multiple Labels

Use multiple labels for pod selection:

```bash
helm install my-lb ./chart \
  --set env.relay="app=myapp,tier=backend"
```

### Custom Service Port

Change the exposed service port:

```yaml
# custom-values.yaml
service:
  port: 8080

env:
  relay: "app=my-app"
```

```bash
helm install my-lb ./chart -f custom-values.yaml
```

## How It Works

1. **Discovery Phase**: The updater container polls the Kubernetes API at regular intervals (default 1s)
2. **Selection Phase**: Filters pods matching the configured label selector in the specified namespace
3. **Comparison Phase**: Compares discovered pods with the previous state
4. **Update Phase**: If changes detected, sends new configuration to Traefik REST API
5. **Routing Phase**: Traefik applies the new backend configuration and routes traffic using least connections algorithm

## Monitoring

### Prometheus Metrics

Traefik exposes Prometheus metrics on port `9090`:

```bash
# Port-forward to access metrics
kubectl port-forward <pod-name> 9090:9090

# Access metrics
curl http://localhost:9090/metrics
```

### Logs

View structured JSON logs:

```bash
# Updater logs
kubectl logs -f <pod-name> -c ilb

# Traefik logs
kubectl logs -f <pod-name> -c traefik
```

## Troubleshooting

### Pods Not Being Discovered

1. Check label selector:
```bash
kubectl get pods -l "app=my-app"
```

2. Verify RBAC permissions:
```bash
kubectl auth can-i list pods --as=system:serviceaccount:default:my-loadbalancer
```

3. Check updater logs:
```bash
kubectl logs <pod-name> -c ilb | grep "Found pods"
```

### Traefik Not Updating

1. Verify Traefik API is accessible:
```bash
kubectl exec -it <pod-name> -c ilb -- wget -O- http://localhost:8080/api/providers/rest
```

2. Check for API errors in logs:
```bash
kubectl logs <pod-name> -c ilb | grep "Failed to update"
```

### Connection Issues

1. Test direct connectivity to backend pods:
```bash
kubectl exec -it <pod-name> -- telnet <backend-pod-ip> 3333
```

2. Check Traefik routing configuration:
```bash
kubectl exec -it <pod-name> -c traefik -- cat /etc/traefik/traefik.yml
```

## Development

### Running Locally

```bash
# Set up environment variables
export POD_LABELS="app=test"
export TRAEFIK_API_URL="http://localhost:8080/api/providers/rest"
export POD_NAMESPACE="default"
export UPDATE_INTERVAL="5s"

# Run the application
go run main.go
```

### Running Tests

```bash
make test
```

### Linting

```bash
make lint
```

## Security Considerations

### Production Deployment

For production environments, consider:

1. **Enable TLS** for Traefik API (modify `traefik.yml`)
2. **Implement authentication** for Traefik dashboard
3. **Use NetworkPolicies** to restrict access
4. **Review RBAC permissions** and apply principle of least privilege
5. **Enable Pod Security Standards**

### Security Context

The Helm chart includes secure defaults:
- Runs as non-root user (UID 65534)
- Read-only root filesystem
- All capabilities dropped
- Privilege escalation disabled

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/k8s-internal-loadbalancer/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/k8s-internal-loadbalancer/discussions)

## Roadmap

- [ ] Replace polling with Kubernetes watch API
- [ ] Add comprehensive unit and integration tests
- [ ] Implement circuit breaker pattern for Traefik API calls
- [ ] Support for multiple Traefik instances
- [ ] Health and readiness probes
- [ ] Custom metrics export
- [ ] Helm chart repository publishing

## Acknowledgments

- Built with [Traefik](https://traefik.io/) v3.1.4
- Uses [client-go](https://github.com/kubernetes/client-go) for Kubernetes integration
- Inspired by Kubernetes service discovery patterns
