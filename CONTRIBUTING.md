# Contributing to K8s Internal Load Balancer

Thank you for your interest in contributing to K8s Internal Load Balancer! We welcome contributions from the community.

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (code snippets, config files, etc.)
- **Describe the observed behavior** and **explain what you expected**
- **Include logs and error messages**
- **Specify your environment**: Kubernetes version, Go version, OS

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful**
- **List any similar features** in other projects if applicable

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following our coding standards
3. **Add tests** for new functionality
4. **Ensure all tests pass**: `make test`
5. **Run the linter**: `make lint`
6. **Update documentation** if needed
7. **Commit your changes** using clear commit messages
8. **Push to your fork** and submit a pull request

## Development Setup

### Prerequisites

- Go 1.22 or later
- Docker (for building images)
- Kubernetes cluster (for testing) - minikube, kind, or k3s recommended
- Helm 3.0+

### Local Development

1. Clone your fork:
```bash
git clone https://github.com/YOUR_USERNAME/k8s-internal-loadbalancer.git
cd k8s-internal-loadbalancer
```

2. Install dependencies:
```bash
go mod download
```

3. Run locally:
```bash
export POD_LABELS="app=test"
export TRAEFIK_API_URL="http://localhost:8080/api/providers/rest"
export POD_NAMESPACE="default"

go run main.go
```

4. Build:
```bash
make build
```

### Testing

Run all tests:
```bash
make test
```

Run tests with coverage:
```bash
go test -cover ./...
```

### Linting

We use `golangci-lint` for code quality:

```bash
make lint
```

Install golangci-lint if not present:
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

## Coding Standards

### Go Style Guide

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting (enforced by CI)
- Add comments to exported functions, types, and constants

### Code Structure

- Keep functions small and focused (< 50 lines when possible)
- Use meaningful variable and function names
- Avoid deep nesting (max 3-4 levels)
- Handle errors explicitly - never ignore errors

### Example

```go
// Good
func getPods(ctx context.Context, labels string) ([]Pod, error) {
    pods, err := client.List(ctx, labels)
    if err != nil {
        return nil, fmt.Errorf("failed to list pods: %w", err)
    }
    return pods, nil
}

// Bad
func get_pods(l string) []Pod {
    pods, _ := client.List(context.Background(), l)
    return pods
}
```

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that don't affect code meaning (formatting, etc.)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `chore`: Changes to build process or auxiliary tools

Examples:
```
feat: add support for custom backend ports

fix: prevent race condition in pod discovery

docs: update installation instructions for Helm 3

refactor: extract HTTP client creation to separate function
```

## Pull Request Process

1. **Update documentation** if you're changing behavior
2. **Add tests** for new functionality
3. **Ensure CI passes** - all tests and linting must pass
4. **Request review** from maintainers
5. **Address feedback** in a timely manner
6. **Squash commits** if requested before merging

### PR Title

Use the same format as commit messages:
```
feat: add health check endpoints
```

### PR Description Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe how you tested your changes

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-reviewed my code
- [ ] Commented complex code sections
- [ ] Updated documentation
- [ ] Added tests for new functionality
- [ ] All tests pass locally
- [ ] No new warnings generated
```

## Project Structure

```
.
├── main.go              # Main application (pod watcher)
├── operator.go          # Kubernetes operator (CRD controller)
├── Dockerfile           # Container image for main app
├── Dockerfile-traefik   # Container image for Traefik
├── traefik.yml          # Traefik configuration
├── chart/               # Helm chart
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
├── examples/            # Example configurations
├── .github/             # GitHub Actions workflows
│   └── workflows/
└── docs/                # Additional documentation
```

## Release Process

Releases are managed by maintainers:

1. Update version in `chart/Chart.yaml`
2. Update `CHANGELOG.md`
3. Create and push a version tag: `git tag -a v0.1.0 -m "Release v0.1.0"`
4. GitHub Actions automatically builds and pushes Docker images
5. Create GitHub Release with changelog

## Getting Help

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Pull Request Comments**: For code review discussions

## Recognition

Contributors will be recognized in:
- README.md contributors section
- Release notes
- Project documentation

Thank you for contributing to K8s Internal Load Balancer!
