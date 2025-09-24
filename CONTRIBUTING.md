# Contributing to kubectl-reboot

Thank you for your interest in contributing to kubectl-reboot! This document provides guidelines and information for contributors.

## Code of Conduct

This project follows the [Kubernetes Community Code of Conduct](https://github.com/kubernetes/community/blob/master/code-of-conduct.md). By participating, you are expected to uphold this code.

## Quick Start for Contributors

1. **Fork** the repository on GitHub
2. **Clone** your fork locally
3. **Create** a feature branch from `main`
4. **Make** your changes and add tests
5. **Run** tests and linting: `make check`
6. **Commit** with conventional commit messages
7. **Push** to your fork and **submit** a Pull Request

## How to Contribute

### 🐛 Reporting Bugs

1. **Check existing issues** first to avoid duplicates
2. **Use the bug report template** when creating new issues
3. **Provide detailed information**:
   - kubectl-reboot version (`kubectl reboot --version`)
   - Kubernetes version (`kubectl version`)
   - Operating system and architecture
   - SSH configuration and target node details
   - Steps to reproduce the issue
   - Expected vs actual behavior
   - Complete error messages and logs

### 💡 Suggesting Features

1. **Check existing feature requests** first
2. **Open an issue** with the feature request template
3. **Describe the use case** and why the feature would be valuable
4. **Provide examples** of how the feature would be used
5. **Consider backwards compatibility** implications

### 🔧 Contributing Code

#### Prerequisites

- **Go 1.23 or later**
- **kubectl** installed and configured
- **Access to a Kubernetes cluster** for testing (kind, minikube, or cloud cluster)
- **SSH access** to test nodes (for integration testing)
- **Git** configured with your GitHub account

#### Development Environment Setup

```bash
# Clone your fork
git clone https://github.com/YOUR-USERNAME/kubectl-reboot.git
cd kubectl-reboot

# Add upstream remote
git remote add upstream https://github.com/ayetkin/kubectl-reboot.git

# Install dependencies
go mod download

# Run initial build and tests
make build
make test
make check
```

#### Development Workflow

1. **Sync with upstream**:
   ```bash
   git checkout main
   git pull upstream main
   git push origin main
   ```

2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**:
   - Follow existing code patterns
   - Add comprehensive tests
   - Update documentation
   - Test manually with real clusters

4. **Validate your changes**:
   ```bash
   make check      # Run all linting and formatting checks
   make test       # Run unit tests
   make build      # Ensure it builds successfully
   ./bin/kubectl-reboot --help  # Test the binary
   ```

5. **Commit with conventional commit format**:
   ```bash
   git add .
   git commit -m "feat: add support for custom drain timeout"
   ```

   **Commit message format**:
   - `feat:` - New features
   - `fix:` - Bug fixes
   - `docs:` - Documentation changes
   - `test:` - Adding or updating tests
   - `refactor:` - Code refactoring without functional changes
   - `perf:` - Performance improvements
   - `ci:` - CI/CD related changes
   - `chore:` - Maintenance and tooling

6. **Push and create Pull Request**:
   ```bash
   git push origin feature/your-feature-name
   ```

#### Code Guidelines

**Go Style**:
- Follow standard Go conventions (`go fmt`, `go vet`)
- Use meaningful variable and function names
- Keep functions focused and small (< 50 lines when possible)
- Handle all errors appropriately
- Use the existing logging patterns with structured logging

**Architecture Principles**:
- **Separation of Concerns**: Keep Kubernetes operations, SSH operations, and business logic separate
- **Testability**: Write code that can be easily unit tested
- **Error Handling**: Provide clear, actionable error messages
- **Configuration**: Use the existing configuration patterns
- **Logging**: Use structured logging with appropriate levels

**Testing**:
- **Unit Tests**: Test individual functions and components
- **Integration Tests**: Test end-to-end workflows where possible
- **Error Scenarios**: Test error conditions and edge cases
- **Multiple Platforms**: Consider platform-specific behavior

#### Adding New Features

**Major Features** (new commands, significant workflow changes):
1. **Open an issue first** to discuss the design
2. **Write a design document** for complex features
3. **Break down into smaller PRs** when possible
4. **Update help text** and documentation
5. **Add comprehensive tests**

**Minor Features** (flags, small enhancements):
1. **Follow existing patterns** for similar functionality
2. **Add appropriate validation**
3. **Update help text**
4. **Add tests**

#### File Organization

```
├── cmd/k8s-restart/       # Main application entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── kube/            # Kubernetes client operations
│   └── ssh/             # SSH client operations  
├── scripts/             # Build and release scripts
├── .github/workflows/   # CI/CD pipelines
└── docs/               # Additional documentation
```

### 📝 Documentation

- **README updates**: Keep installation and usage instructions current
- **Help text**: Update command-line help for new options
- **Examples**: Add real-world usage examples
- **CHANGELOG**: Follow Keep a Changelog format
- **Code comments**: Document complex logic and public APIs

### 🧪 Testing

#### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run specific tests
go test ./internal/config -v

# Run integration tests (requires cluster access)
go test -tags=integration ./test/...
```

#### Test Categories

- **Unit Tests**: Fast, isolated tests for individual components
- **Integration Tests**: Tests that require Kubernetes cluster access
- **End-to-End Tests**: Full workflow tests with real nodes

#### Writing Tests

- **Use table-driven tests** for multiple scenarios
- **Mock external dependencies** (Kubernetes API, SSH connections)
- **Test error conditions** as well as success paths
- **Use descriptive test names** that explain what's being tested

### 🚀 Pull Request Process

#### Before Submitting

- [ ] **Tests pass**: `make check` succeeds
- [ ] **Documentation updated**: README, help text, etc.
- [ ] **Self-review completed**: Check your own code for issues
- [ ] **Commits are clean**: Squash fixup commits, use clear messages
- [ ] **Branch is current**: Rebased on latest main

#### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review performed
- [ ] Documentation updated
- [ ] Tests added for new functionality
- [ ] All tests pass
```

#### Review Process

1. **Automated checks**: CI must pass
2. **Code review**: At least one maintainer review
3. **Testing**: Reviewers may test manually
4. **Approval**: Maintainer approval required for merge
5. **Merge**: Squash and merge preferred

## 🏷️ Release Process

Releases are automated but follow this process:

### For Maintainers

1. **Update version** in relevant files
2. **Update CHANGELOG.md** with new version
3. **Create and push tag**: `./scripts/release.sh v1.x.x`
4. **GitHub Actions** handles the rest automatically
5. **Update Krew index** after release is published

### Release Script Usage

```bash
# Create a new release
./scripts/release.sh v1.2.3

# Preview what would happen
./scripts/release.sh v1.2.3 --dry-run

# Create draft release for testing
./scripts/release.sh v1.2.3 --draft
```

## 🆘 Getting Help

- **Issues**: Search existing issues or create new ones
- **Discussions**: Use GitHub Discussions for questions
- **Documentation**: Check README and code comments
- **Examples**: Look at existing tests and usage patterns

## 🎖️ Recognition

Contributors are recognized through:
- **Contributors section** in README
- **Release notes** for significant contributions  
- **GitHub's contributor statistics**
- **Special thanks** in major release announcements

## 📄 License

By contributing to kubectl-reboot, you agree that your contributions will be licensed under the project's MIT License.
