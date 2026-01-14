# Examples

This directory contains example configurations for deploying K8s Internal Load Balancer.

## Files

### Kubernetes Manifests

- **basic-deployment.yaml**: Example application deployment that you want to load balance

### Helm Values

- **basic-values.yaml**: Minimal configuration for getting started
- **production-values.yaml**: Production-ready configuration with resource limits and anti-affinity
- **ha-values.yaml**: High availability configuration with autoscaling

## Usage

### Basic Deployment

1. Deploy your application:
```bash
kubectl apply -f examples/basic-deployment.yaml
```

2. Install the load balancer:
```bash
helm install my-lb ./chart --values examples/basic-values.yaml
```

3. Verify deployment:
```bash
kubectl get pods
kubectl logs -f <pod-name> -c ilb
```

### Production Deployment

```bash
helm install prod-lb ./chart \
  --values examples/production-values.yaml \
  --namespace production \
  --create-namespace
```

### High Availability Deployment

```bash
helm install ha-lb ./chart \
  --values examples/ha-values.yaml \
  --namespace critical \
  --create-namespace
```

## Customization

### Label Selectors

The `env.relay` field accepts Kubernetes label selectors:

```yaml
# Single label
env:
  relay: "app=my-app"

# Multiple labels
env:
  relay: "app=my-app,tier=backend,version=v1"
```

### Update Interval

Configure how often to check for pod changes:

```yaml
env:
  # Check every second (high responsiveness, more API calls)
  updateinterval: "1s"

  # Check every 10 seconds (balanced)
  updateinterval: "10s"

  # Check every minute (low overhead)
  updateinterval: "60s"
```

### Service Type

```yaml
# ClusterIP (default, internal only)
service:
  type: ClusterIP
  port: 3333

# LoadBalancer (expose externally)
service:
  type: LoadBalancer
  port: 3333

# NodePort (access via node IP)
service:
  type: NodePort
  port: 3333
  nodePort: 30333
```

## Testing

Test the load balancer is working:

```bash
# Get the service endpoint
kubectl get svc -l app=k8s-internal-loadbalancer

# Test connectivity (from within cluster)
kubectl run -it --rm test --image=busybox --restart=Never -- \
  telnet <service-name> 3333

# Check logs
kubectl logs -l app=k8s-internal-loadbalancer -c ilb --tail=100

# Check Traefik logs
kubectl logs -l app=k8s-internal-loadbalancer -c traefik --tail=100
```

## Troubleshooting

### Pods not discovered

Check label selector matches your pods:
```bash
# What the load balancer is looking for
kubectl get svc my-lb -o jsonpath='{.spec.selector}'

# What pods exist with those labels
kubectl get pods -l "app=my-app"
```

### No traffic routing

Check Traefik configuration:
```bash
kubectl exec -it <pod-name> -c traefik -- cat /etc/traefik/traefik.yml
```

### Permission errors

Check RBAC:
```bash
kubectl auth can-i list pods \
  --as=system:serviceaccount:default:my-lb-k8s-internal-loadbalancer
```

## Next Steps

- See [main README](../README.md) for detailed documentation
- Review [Helm chart README](../chart/README.md) for all available options
- Check [CONTRIBUTING](../CONTRIBUTING.md) for development guidelines
