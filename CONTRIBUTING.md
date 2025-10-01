# Contributing to Seata Go Client

Thank you for your interest in contributing to the Seata Go Client! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make (optional, for using Makefile)

### Development Setup

1. **Fork the repository**
   ```bash
   # Fork the repository on GitHub, then clone your fork
   git clone https://github.com/your-username/seata-go-client.git
   cd seata-go-client
   ```

2. **Set up the development environment**
   ```bash
   # Install dependencies
   go mod tidy
   
   # Install development tools (optional)
   make install-tools
   ```

3. **Run tests to ensure everything works**
   ```bash
   make test
   # or
   go test ./...
   ```

## 📝 Development Guidelines

### Code Style

- Follow Go naming conventions
- Use `gofmt` to format code
- Use meaningful variable and function names
- Add comments for public APIs
- Keep functions focused and small

### Testing

- Write tests for all new functionality
- Maintain or improve test coverage
- Use table-driven tests where appropriate
- Test error conditions and edge cases

### Documentation

- Update README.md for new features
- Add examples for new functionality
- Update API documentation
- Include usage examples in code comments

## 🔧 Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Your Changes

- Write your code following the guidelines above
- Add tests for your changes
- Update documentation as needed

### 3. Run Quality Checks

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Run all checks
make check
```

### 4. Commit Your Changes

```bash
git add .
git commit -m "feat: add new feature description"
# or
git commit -m "fix: fix bug description"
```

Use conventional commit messages:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test changes
- `refactor:` for code refactoring
- `perf:` for performance improvements

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## 🧪 Testing Guidelines

### Unit Tests

- Test individual functions and methods
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test both success and error cases

### Integration Tests

- Test complete workflows
- Test with real Seata server (when possible)
- Test error handling and recovery

### Example Tests

```go
func TestNewClient(t *testing.T) {
    config := DefaultConfig()
    client := NewClient(config)
    
    assert.NotNil(t, client)
    assert.NotNil(t, client.httpClient)
    assert.Equal(t, config, client.config)
}

func TestSagaWorkflowValidation(t *testing.T) {
    // Test empty workflow
    workflow := CreateSagaWorkflow([]SagaStep{})
    err := workflow.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "at least one step")
}
```

## 📚 Documentation Guidelines

### README Updates

- Update feature list for new features
- Add usage examples
- Update installation instructions if needed
- Update configuration options

### Code Documentation

```go
// Client represents a Seata client for distributed transaction management
type Client struct {
    httpClient *resty.Client
    grpcClient *GrpcClient
    config     *Config
}

// NewClient creates a new Seata client with the given configuration
func NewClient(config *Config) *Client {
    // Implementation...
}
```

### Example Documentation

```go
// Example of basic client usage
func ExampleClient_BasicUsage() {
    client := seata.NewClientWithDefaults()
    defer client.Close()
    
    ctx := context.Background()
    payload := []byte(`{"order_id": "12345"}`)
    
    tx, err := client.StartTransaction(ctx, seata.ModeSaga, payload)
    if err != nil {
        log.Fatal(err)
    }
    
    // Add branches and submit...
}
```

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Description**: Clear description of the bug
2. **Steps to Reproduce**: Detailed steps to reproduce the issue
3. **Expected Behavior**: What should happen
4. **Actual Behavior**: What actually happens
5. **Environment**: Go version, OS, etc.
6. **Code Sample**: Minimal code that reproduces the issue

## 💡 Feature Requests

When requesting features, please include:

1. **Description**: Clear description of the feature
2. **Use Case**: Why this feature is needed
3. **Proposed Solution**: How you think it should work
4. **Alternatives**: Other solutions you've considered
5. **Additional Context**: Any other relevant information

## 🔍 Code Review Process

### For Contributors

- Ensure all tests pass
- Ensure code follows style guidelines
- Update documentation as needed
- Respond to review feedback promptly

### For Reviewers

- Check code quality and style
- Verify tests are adequate
- Ensure documentation is updated
- Test the changes if possible

## 📋 Pull Request Template

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] No breaking changes (or documented)
```

## 🏷️ Release Process

1. Update version in `go.mod`
2. Update `CHANGELOG.md`
3. Create release tag
4. Update documentation

## 📞 Getting Help

- **Issues**: [GitHub Issues](https://github.com/seata-team/seata-go-client/issues)
- **Discussions**: [GitHub Discussions](https://github.com/seata-team/seata-go-client/discussions)
- **Documentation**: [README.md](README.md)

## 🙏 Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

Thank you for contributing to Seata Go Client! 🎉
