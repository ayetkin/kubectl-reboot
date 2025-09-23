# Development Guide

This guide covers development setup, testing, and release processes for kubectl-reboot.

## Quick Start

1. **Clone and setup**:
   ```bash
   git clone https://github.com/ayetkin/kubectl-reboot.git
   cd kubectl-reboot
   go mod download
   ```

2. **Build and test**:
   ```bash
   make build
   make test
   ```

3. **Install locally**:
   ```bash
   make install-local
   ```

## Development Workflow

### Project Structure

```
kubectl-reboot/
├── cmd/k8s-restart/           # Main application entry point
│   ├── main.go                # CLI logic and workflow orchestration
│   └── main_test.go           # Unit tests for main package
├── internal/                  # Internal packages (not importable)
│   ├── config/                # Configuration parsing and validation
│   │   └── config.go
│   ├── kube/                  # Kubernetes client wrapper
│   │   └── client.go
│   └── ssh/                   # SSH operations
│       └── ssh.go
├── scripts/                   # Build and deployment scripts
│   └── submit-to-krew.sh      # Krew submission helper
├── .github/workflows/         # CI/CD pipelines
│   ├── ci.yml                 # Continuous integration
│   └── release.yml            # Release automation
├── kubectl-reboot.yaml        # Krew plugin manifest
├── Makefile                   # Build automation
└── README.md                  # User documentation
```

### Building

```bash
# Build for current platform
make build

# Build for all supported platforms
make release

# Package binaries for distribution
make package

# Clean build artifacts
make clean
```

### Testing

```bash
# Run unit tests
make test

# Run all checks (vet, fmt, test)
make check

# Run specific test
go test -v ./cmd/k8s-restart -run TestBuildSSHHost
```

### Code Quality

```bash
# Format code
make fmt

# Vet code
make vet

# Run linter (requires golangci-lint)
golangci-lint run
```

## Testing Strategy

### Unit Tests

- **Location**: `*_test.go` files next to the code they test
- **Coverage**: Focus on business logic and edge cases
- **Mocking**: Use interfaces for external dependencies

### Integration Tests

- **Environment**: Requires access to a Kubernetes cluster
- **Scope**: Test full workflows with real Kubernetes API
- **Safety**: Use dry-run mode to avoid actual node operations

### Manual Testing

1. **Local testing**:
   ```bash
   # Build and test locally
   make build
   ./bin/kubectl-reboot --dry-run --all
   ```

2. **Cluster testing**:
   ```bash
   # Test with dry-run first
   kubectl reboot --dry-run node1
   
   # Test actual restart (use with caution!)
   kubectl reboot node1
   ```

## Release Process

### Versioning

- Follow [Semantic Versioning](https://semver.org/)
- Tag format: `v1.2.3`
- Pre-release format: `v1.2.3-rc.1`

### Creating a Release

1. **Update version references**:
   ```bash
   # Update version in kubectl-reboot.yaml
   sed -i 's/version: v.*/version: v1.2.3/' kubectl-reboot.yaml
   ```

2. **Create and push tag**:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

3. **GitHub Actions will**:
   - Run tests
   - Build binaries for all platforms
   - Create GitHub release
   - Upload release assets

4. **Update Krew manifest**:
   ```bash
   # Download release assets and update checksums
   make krew-manifest
   
   # Submit to krew-index
   ./scripts/submit-to-krew.sh v1.2.3
   ```

### Krew Submission

1. **Automated via script**:
   ```bash
   ./scripts/submit-to-krew.sh v1.2.3
   ```

2. **Manual process**:
   - Fork [krew-index](https://github.com/kubernetes-sigs/krew-index)
   - Update `plugins/restart.yaml` with new version and checksums
   - Submit pull request
   - Wait for review and merge

## Debugging

### Enable Debug Logging

```bash
# Set log level to debug
kubectl restart --dry-run node1 --log-level debug
```

### Common Development Issues

1. **Import path issues**:
   ```bash
   # Ensure go.mod has correct module name
   go mod tidy
   ```

2. **Cross-compilation issues**:
   ```bash
   # Disable CGO for static binaries
   CGO_ENABLED=0 go build
   ```

3. **SSH connectivity testing**:
   ```bash
   # Test SSH connection manually
   ssh -o StrictHostKeyChecking=no -o BatchMode=yes user@node
   ```

## Performance Considerations

- **Concurrent operations**: Currently processes nodes sequentially for safety
- **Memory usage**: Minimal, mostly network I/O bound
- **Network**: SSH connections and Kubernetes API calls

## Security

### Code Security

- **Input validation**: Validate all user inputs
- **SSH security**: Use key-based auth, avoid password auth
- **Secrets**: Never log sensitive information

### Dependencies

```bash
# Audit dependencies
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed contribution guidelines.

### Code Review Checklist

- [ ] Tests pass and cover new functionality
- [ ] Code follows Go conventions and project style
- [ ] Documentation updated for user-facing changes
- [ ] Error handling is appropriate
- [ ] No sensitive information in logs
- [ ] Performance impact considered
- [ ] Security implications reviewed

## Troubleshooting Development Issues

### Go Module Issues

```bash
# Clean module cache
go clean -modcache
go mod download
```

### Build Issues

```bash
# Verbose build output
go build -v ./cmd/k8s-restart

# Check build constraints
go list -f '{{.GoFiles}}' ./cmd/k8s-restart
```

### Test Issues

```bash
# Verbose test output
go test -v -race ./...

# Run specific test with timeout
go test -timeout 30s -run TestBuildSSHHost ./cmd/k8s-restart
```
