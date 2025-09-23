# Contributing to kubectl-reboot

Thank you for your interest in contributing to kubectl-reboot! This document provides guidelines and information for contributors.

## Code of Conduct

This project follows the [Kubernetes Community Code of Conduct](https://github.com/kubernetes/community/blob/master/code-of-conduct.md). By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

1. **Check existing issues** first to avoid duplicates
2. **Use the bug report template** when creating new issues
3. **Provide detailed information**:
   - kubectl-reboot version
   - Kubernetes version
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs or output

### Suggesting Features

1. **Check existing feature requests** first
2. **Open an issue** with the feature request template
3. **Describe the use case** and why the feature would be valuable
4. **Provide examples** of how the feature would be used

### Contributing Code

#### Before You Start

1. **Fork the repository** and create a new branch from `main`
2. **Check existing issues** - look for issues labeled `good first issue` or `help wanted`
3. **Discuss major changes** by opening an issue first

#### Development Setup

1. **Prerequisites**:
   - Go 1.23 or later
   - kubectl installed and configured
   - Access to a Kubernetes cluster for testing

2. **Clone and setup**:
   ```bash
   git clone https://github.com/your-username/kubectl-reboot.git
   cd kubectl-reboot
   go mod download
   ```

3. **Build and test**:
   ```bash
   make build
   make test
   make check
   ```

#### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed
   - Ensure all tests pass

3. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

   Use conventional commit messages:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `test:` for adding tests
   - `refactor:` for code refactoring
   - `chore:` for maintenance tasks

4. **Push and create a pull request**:
   ```bash
   git push origin feature/your-feature-name
   ```

#### Code Style Guidelines

- **Follow Go conventions**: Use `go fmt`, `go vet`, and `golangci-lint`
- **Write clear, descriptive variable and function names**
- **Add comments for complex logic**
- **Keep functions small and focused**
- **Handle errors appropriately**
- **Use structured logging with the existing logger**

#### Testing

- **Write unit tests** for new functionality
- **Test with different Kubernetes versions** when possible
- **Test on different platforms** (Linux, macOS, Windows)
- **Include integration tests** for significant features
- **Verify that `make check` passes**

#### Documentation

- **Update README.md** if adding new features or changing behavior
- **Add or update command-line help text**
- **Include examples** for new functionality
- **Update the Krew manifest** if needed

### Pull Request Process

1. **Ensure your PR has a clear title and description**
2. **Reference related issues** using "Fixes #123" or "Relates to #123"
3. **Include a checklist** of what you've tested
4. **Respond to feedback** promptly and constructively
5. **Keep your branch up to date** with the main branch

#### PR Checklist

- [ ] Code follows the project's style guidelines
- [ ] Self-review of the code has been performed
- [ ] Code is commented, particularly in hard-to-understand areas
- [ ] Corresponding changes to documentation have been made
- [ ] Changes generate no new warnings
- [ ] Unit tests pass locally
- [ ] Integration tests pass (if applicable)
- [ ] Any dependent changes have been merged and published

## Release Process

Releases are automated through GitHub Actions when tags are pushed:

1. **Version bump**: Update version in relevant files
2. **Tag release**: `git tag v1.x.x && git push origin v1.x.x`
3. **GitHub Actions**: Automatically builds and publishes release
4. **Krew manifest**: Update checksums in `kubectl-reboot.yaml`

## Getting Help

- **Documentation**: Check the README.md and inline help
- **Issues**: Search existing issues or create a new one
- **Discussions**: Use GitHub Discussions for questions and ideas

## Recognition

Contributors will be recognized in:
- Release notes for significant contributions
- Contributors section of the README
- GitHub's automatic contributor recognition

## License

By contributing to kubectl-reboot, you agree that your contributions will be licensed under the MIT License.
