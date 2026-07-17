# Contributing to dhook

Thank you for your interest in contributing to dhook! This document provides guidelines for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/RivnZero/dhook.git`
3. Create a feature branch: `git checkout -b feature/my-feature`
4. Make your changes
5. Run tests and lint
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Building

```bash
go build ./...
```

### Running Tests

```bash
go test -v ./...
```

### Running Tests with Race Detector

```bash
go test -v -race ./...
```

### Linting

```bash
go vet ./...
```

## Code Style

- Follow standard Go conventions and idioms
- Run `go vet ./...` before committing
- No inline comments in Go source files; code must be self-documenting
- Use meaningful variable and function names
- Keep functions focused and small

## Pull Request Process

1. Ensure all tests pass
2. Ensure `go vet ./...` reports no issues
3. Update documentation if your change affects the public API
4. Keep pull requests focused on a single change
5. Write a clear description of what the PR does and why

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include your Go version (`go version`) and OS
- Provide a minimal reproduction when reporting bugs
- For security vulnerabilities, see [SECURITY.md](SECURITY.md)

## License

By contributing, you agree that your contributions will be licensed under the
MIT License with Attribution Requirement as defined in [LICENSE](LICENSE).
