# Contributing to NVIDIA FabricManager Go Package

Thank you for your interest in contributing to the NVIDIA FabricManager Go Package! This document provides guidelines for contributing to this project.

## Development Setup

### Prerequisites

1. **Go 1.22 or later**
   ```bash
   # Check your Go version
   go version
   ```

2. **CGO enabled** (required for C library bindings)
   ```bash
   # Verify CGO is available
   go env CGO_ENABLED
   ```

3. **NVIDIA FabricManager Development Package**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install nvidia-fabricmanager-dev-<version>
   
   # RHEL/CentOS
   sudo yum install nvidia-fabricmanager-devel-<version>
   ```

4. **NVIDIA FabricManager Library**
   ```bash
   # Ubuntu/Debian
   sudo apt-get install nvidia-fabricmanager-<version>
   
   # RHEL/CentOS
   sudo yum install nvidia-fabricmanager-<version>
   ```

### Building

1. **Clone the repository**
   ```bash
   git clone https://github.com/NVIDIA/go-fabricmanager.git
   cd go-fabricmanager
   ```

2. **Build the project**
   ```bash
   # Use the build script
   ./build.sh
   
   # Or use make
   make build
   
   # Or build manually
   CGO_ENABLED=1 go build -o fmpm cmd/fmpm/main.go
   ```

## Development Workflow

### Code Style

- Follow Go conventions and the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code
- Use `golint` to check for style issues
- Use `go vet` to check for common mistakes

### Testing

1. **Run unit tests**
   ```bash
   CGO_ENABLED=1 go test -v ./...
   ```

2. **Run integration tests** (requires running FabricManager)
   ```bash
   # Start FabricManager service
   sudo systemctl start nvidia-fabricmanager
   
   # Run integration tests
   CGO_ENABLED=1 go test -v -tags=integration ./...
   ```

3. **Test the CLI tool**
   ```bash
   ./fmpm --help
   ./fmpm list
   ```

### Adding New Features

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Implement your changes**
   - Add new functions to `fabricmanager.go`
   - Add corresponding CLI commands in `cmd/fmpm/main.go`
   - Add tests in `fabricmanager_test.go`
   - Update documentation

3. **Test your changes**
   ```bash
   make test
   make build
   ./fmpm --help
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "Add feature: brief description"
   ```

### Error Handling

When adding new functionality, ensure proper error handling:

1. **Use the existing error types**
   ```go
   return &FMError{Code: FM_ST_BADPARAM, Message: "Invalid parameter"}
   ```

2. **Add error classification functions** if needed
   ```go
   func IsNewErrorType(err error) bool {
       if fmErr, ok := err.(*FMError); ok {
           return fmErr.Code == FM_ST_NEW_ERROR_TYPE
       }
       return false
   }
   ```

3. **Update tests** to cover error cases

### Documentation

1. **Update README.md** for new features
2. **Add examples** in the `examples/` directory
3. **Update API documentation** in code comments
4. **Update CLI help** text

## Submitting Changes

### Pull Request Process

1. **Fork the repository** on GitHub
2. **Push your changes** to your fork
3. **Create a pull request** with a clear description
4. **Ensure all tests pass**
5. **Request review** from maintainers

### Pull Request Guidelines

- **Title**: Clear, concise description
- **Description**: 
  - What the change does
  - Why it's needed
  - How it was tested
  - Any breaking changes
- **Tests**: Ensure all tests pass
- **Documentation**: Update relevant docs

### Commit Message Format

Use conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Examples:
```
feat(cli): add support for custom timeout values
fix(core): handle nil pointer in partition activation
docs(readme): update installation instructions
test(api): add tests for error handling
```

## Code Review

All contributions require review before merging. Reviewers will check:

- **Functionality**: Does the code work as intended?
- **Testing**: Are there adequate tests?
- **Documentation**: Is the code well-documented?
- **Style**: Does the code follow Go conventions?
- **Error handling**: Are errors handled appropriately?
- **Security**: Are there any security concerns?

## Getting Help

- **Issues**: Use GitHub issues for bug reports and feature requests
- **Discussions**: Use GitHub discussions for questions and general discussion
- **Documentation**: Check the README and examples

## License

By contributing to this project, you agree that your contributions will be licensed under the same terms as the project (see LICENSE file).

## Code of Conduct

Please be respectful and inclusive in all interactions. We follow the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/0/code_of_conduct/). 