# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Currently supported versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

We take the security of K8s Internal Load Balancer seriously. If you believe you have found a security vulnerability, please report it to us responsibly.

### Please Do NOT

- Open a public GitHub issue for security vulnerabilities
- Disclose the vulnerability publicly before it has been addressed

### Please DO

1. **Email us**: Send details to the project maintainers via GitHub (create a security advisory)
2. **Provide details**: Include as much information as possible:
   - Type of vulnerability
   - Full paths of source files related to the vulnerability
   - Location of the affected source code (tag/branch/commit or direct URL)
   - Step-by-step instructions to reproduce the issue
   - Proof-of-concept or exploit code (if possible)
   - Impact of the vulnerability

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 72 hours
- **Assessment**: We will assess the vulnerability and determine its impact and severity
- **Fix**: We will work on a fix and prepare a security advisory
- **Disclosure**: Once a fix is available, we will:
  - Release a patched version
  - Publish a security advisory
  - Credit you for the discovery (unless you prefer to remain anonymous)

### Timeline

- **Initial response**: Within 72 hours
- **Status update**: Within 7 days
- **Fix timeline**: Depends on severity, typically:
  - Critical: 7 days
  - High: 14 days
  - Medium: 30 days
  - Low: 90 days

## Security Best Practices

### For Users

When deploying K8s Internal Load Balancer, follow these security best practices:

1. **Network Security**
   - Use NetworkPolicies to restrict traffic
   - Keep Traefik API accessible only from localhost or trusted networks
   - Implement TLS for production deployments

2. **Authentication & Authorization**
   - Review and minimize RBAC permissions
   - Use dedicated service accounts
   - Enable Pod Security Standards

3. **Container Security**
   - Use the provided security contexts (enabled by default)
   - Run as non-root user (enforced in Helm chart)
   - Keep images up to date

4. **Configuration**
   - Protect sensitive configuration with Kubernetes Secrets
   - Regularly review and audit configuration
   - Use read-only root filesystem (enabled by default)

5. **Monitoring**
   - Monitor logs for suspicious activity
   - Set up alerts for unusual behavior
   - Regular security audits

### Known Security Considerations

1. **Traefik REST API**: By default, the Traefik REST API is configured with `insecure: true` for the REST provider. This is necessary for the load balancer to function but should only be accessible from localhost or a secured network.

2. **Dashboard Access**: The Traefik dashboard is disabled by default in the latest configuration. If you enable it, ensure proper authentication is configured.

3. **Pod Discovery**: The updater requires RBAC permissions to list and watch pods in the namespace. Review these permissions to ensure they align with your security requirements.

## Security Updates

Security updates will be announced via:
- GitHub Security Advisories
- Release notes
- README.md

Subscribe to repository releases to stay informed about security updates.

## Compliance

This project aims to follow:
- OWASP Top 10 security practices
- Kubernetes security best practices
- CIS Kubernetes Benchmark recommendations

## Contact

For security concerns, please use GitHub's security advisory feature or contact the maintainers through GitHub issues (for non-sensitive security discussions).

## Acknowledgments

We appreciate the security research community's efforts in responsibly disclosing vulnerabilities. Contributors who report valid security issues will be acknowledged in our security advisories and release notes (unless they prefer to remain anonymous).
