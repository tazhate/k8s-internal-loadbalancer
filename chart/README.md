# K8s Internal LoadBalancer Helm Chart

This Helm chart deploys the Kubernetes Internal Load Balancer with Traefik for dynamic TCP load balancing.

## TL;DR

```bash
helm install my-loadbalancer ./chart \
  --set env.relay="app=my-app" \
  --set env.updateinterval="5s"
```

## Introduction

This chart bootstraps a Kubernetes Internal Load Balancer deployment on a Kubernetes cluster using the Helm package manager.

## Prerequisites

- Kubernetes 1.20+
- Helm 3.0+

## Installing the Chart

To install the chart with the release name `my-loadbalancer`:

```bash
helm install my-loadbalancer ./chart
```

The command deploys the load balancer on the Kubernetes cluster with default configuration. The [Parameters](#parameters) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-loadbalancer` deployment:

```bash
helm delete my-loadbalancer
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

### Global Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `nameOverride` | String to partially override name | `""` |
| `fullnameOverride` | String to fully override name | `""` |

### Image Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Image repository | `tazhate/k8s-internal-loadbalancer` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.tag` | Image tag for updater container | `traefik-wrap6` |
| `image.tag_traefik` | Image tag for Traefik container | `traefik11` |

### Service Account Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `serviceAccount.create` | Specifies whether a service account should be created | `true` |
| `serviceAccount.automount` | Automatically mount SA token | `true` |
| `serviceAccount.annotations` | Annotations for service account | `{}` |
| `serviceAccount.name` | The name of the service account to use | `""` (auto-generated) |

### Security Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podSecurityContext.runAsNonRoot` | Run pod as non-root user | `true` |
| `podSecurityContext.runAsUser` | User ID to run the pod as | `65534` |
| `podSecurityContext.fsGroup` | Group ID for filesystem | `65534` |
| `podSecurityContext.seccompProfile.type` | Seccomp profile type | `RuntimeDefault` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.capabilities.drop` | Capabilities to drop | `[ALL]` |
| `securityContext.readOnlyRootFilesystem` | Read-only root filesystem | `true` |
| `securityContext.runAsNonRoot` | Run container as non-root | `true` |
| `securityContext.runAsUser` | User ID to run container as | `65534` |

### Service Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `3333` |

### Environment Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `env.relay` | Label selector for target pods | `stratum-relay-demo-bestpool-1` |
| `env.portname` | Port name (legacy, unused) | `relay-3333` |
| `env.updateinterval` | Update check interval | `5s` |

### Resource Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | Not set |
| `resources.limits.memory` | Memory limit | Not set |
| `resources.requests.cpu` | CPU request | Not set |
| `resources.requests.memory` | Memory request | Not set |

### Autoscaling Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `autoscaling.enabled` | Enable horizontal pod autoscaler | `false` |
| `autoscaling.minReplicas` | Minimum number of replicas | `1` |
| `autoscaling.maxReplicas` | Maximum number of replicas | `100` |
| `autoscaling.targetCPUUtilizationPercentage` | Target CPU utilization | `80` |

### Metrics Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `metrics.enabled` | Enable Prometheus metrics | `"true"` |

### Additional Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podAnnotations` | Annotations for pods | `{}` |
| `podLabels` | Labels for pods | `{}` |
| `nodeSelector` | Node selector for pod assignment | `{}` |
| `tolerations` | Tolerations for pod assignment | `[]` |
| `affinity` | Affinity rules for pod assignment | `{}` |
| `volumes` | Additional volumes | `[]` |
| `volumeMounts` | Additional volume mounts | `[]` |

## Configuration Examples

### Basic Configuration

```yaml
# values.yaml
env:
  relay: "app=my-app"
  updateinterval: "10s"

service:
  port: 3333

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Production Configuration

```yaml
# production-values.yaml
replicaCount: 2

env:
  relay: "app=backend,tier=production"
  updateinterval: "5s"

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app
            operator: In
            values:
            - k8s-internal-loadbalancer
        topologyKey: kubernetes.io/hostname

metrics:
  enabled: "true"
```

### High Availability Configuration

```yaml
# ha-values.yaml
replicaCount: 3

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app
          operator: In
          values:
          - k8s-internal-loadbalancer
      topologyKey: kubernetes.io/hostname

tolerations:
- key: "dedicated"
  operator: "Equal"
  value: "loadbalancer"
  effect: "NoSchedule"
```

## Upgrading

### To 0.1.0

This is the initial release.

## Support

For issues and questions:
- GitHub Issues: https://github.com/yourusername/k8s-internal-loadbalancer/issues
- Documentation: https://github.com/yourusername/k8s-internal-loadbalancer

## License

MIT License - see [LICENSE](../LICENSE) for details
