# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-01-XX

### Added
- Initial release of K8s Internal Load Balancer
- Dynamic pod discovery based on label selectors
- Automatic Traefik configuration via REST API
- Kubernetes operator with CRD support
- Helm chart for easy deployment
- Security-hardened configuration:
  - Non-root container execution
  - Read-only root filesystem
  - Dropped capabilities
  - Pod Security Standards compliance
- Prometheus metrics export from Traefik
- Comprehensive documentation:
  - README with quick start guide
  - CONTRIBUTING guidelines
  - CODE_OF_CONDUCT
  - SECURITY policy
  - Helm chart README
- CI/CD with GitHub Actions:
  - Automated linting and builds
  - Multi-arch Docker image builds (amd64, arm64)
  - Automated Docker Hub publishing on tags
- Development tooling:
  - Makefile with common tasks
  - golangci-lint configuration
  - GitHub issue and PR templates
- Examples and configuration templates
- MIT License

### Features
- **Pod Discovery**: Continuously monitors Kubernetes API for pods matching configured labels
- **Smart Updates**: Only updates Traefik when backend list changes
- **Load Balancing**: Uses least connections algorithm for optimal distribution
- **Sidecar Pattern**: Runs updater and Traefik in single pod
- **RBAC Support**: Full Kubernetes RBAC integration
- **Configurable**: Environment-based configuration for flexibility

### Technical Details
- Built with Go 1.22
- Uses Traefik v3.1.4
- Kubernetes client-go v0.31.1
- Alpine Linux 3.20.3 base image
- Multi-stage Docker builds for minimal image size

### Known Limitations
- Uses polling instead of Kubernetes watch API
- Traefik REST API requires insecure mode for provider updates
- No built-in TLS support (manual configuration required)
- Single namespace support per deployment

### Security
- All security contexts enabled by default
- Traefik dashboard disabled by default
- RBAC permissions minimized to required operations
- Container security best practices implemented

## Release Notes

### Migration Guide
This is the first release, no migration needed.

### Upgrade Instructions
N/A - Initial release

### Breaking Changes
N/A - Initial release

### Deprecations
None

---

## Future Plans

Planned for future releases:

### v0.2.0
- Replace polling with Kubernetes watch API for better performance
- Add health and readiness probe endpoints
- Implement comprehensive test suite
- Add custom Prometheus metrics

### v0.3.0
- Circuit breaker for Traefik API calls
- Support for multiple namespaces
- TLS support for Traefik API
- Configuration hot-reload

### v1.0.0
- Stable API
- Production-grade monitoring and alerting
- Advanced load balancing algorithms
- Multi-tenancy support

---

[Unreleased]: https://github.com/yourusername/k8s-internal-loadbalancer/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/yourusername/k8s-internal-loadbalancer/releases/tag/v0.1.0
