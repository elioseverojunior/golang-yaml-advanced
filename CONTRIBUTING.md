# Contributing to golang-yaml-advanced

Thank you for your interest in contributing to golang-yaml-advanced! We welcome contributions from the community and are pleased to have you aboard.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Security](#security)
- [Community](#community)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

### Prerequisites

- **Go**: Version 1.20 or higher
- **Git**: For version control
- **Make**: For running build tasks

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

   ```bash
   git clone https://github.com/elioetibr/golang-yaml-advanced.git
   cd golang-yaml-advanced
   ```

3. Add the upstream repository:

   ```bash
   git remote add upstream https://github.com/elioetibr/golang-yaml-advanced.git
   ```

## Development Setup

### Initial Setup

```bash
# Install dependencies
go mod download

# Run initial tests
go test ./...

# Check formatting
gofmt -l .

# Run linting
go vet ./...
```

### Development Environment Setup

Run the setup script to configure git hooks and development tools:

```bash
./scripts/setup-hooks.sh
```

This will:

- Configure git hooks for commit message validation
- Set up conventional commit template
- Install commitlint dependencies (if Node.js available)
- Verify Go toolchain and tools
- Run initial project checks
- Configure helpful git aliases

### Commit Message Format

We use [conventional commits](https://www.conventionalcommits.org/) for consistent commit messages:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**

- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build tasks, dependency updates, etc.
- `perf`: Performance improvements
- `ci`: CI/CD changes
- `build`: Build system changes
- `revert`: Reverting previous commits

**Scopes (optional):**
`parser`, `scanner`, `emitter`, `composer`, `constructor`, `lib`, `cli`, `docs`, `tests`, `benches`, `examples`, `ci`, `deps`, `release`

**Examples:**

```bash
feat(parser): add support for complex mapping keys
fix(emitter): resolve string quoting for version numbers
docs: update installation instructions
test(scanner): add edge case tests for deeply nested structures
chore: bump version to 1.1.0 [skip ci]
```

**Special markers:**

- `+semver: major|minor|patch|none` - Control version bumping
- `[skip ci]` or `[ci skip]` - Skip CI builds (for version bumps, docs-only changes)
- `BREAKING CHANGE: description` - Indicate breaking changes

### Development Tools

We recommend installing these additional tools:

```bash
# For test coverage visualization
go install github.com/mattn/goveralls@latest

# For benchmarking
go install golang.org/x/perf/cmd/benchstat@latest

# For security scanning
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

### Project Structure

```
golang-yaml-advanced/
├── src/                    # Source code
│   ├── scanner/           # Token scanning
│   ├── parser/            # Event parsing
│   ├── composer.rs        # Node composition
│   ├── emitter.rs         # YAML output
│   └── lib.rs             # Main library
├── tests/                 # Integration tests
├── benches/              # Performance benchmarks
├── examples/             # Usage examples
└── docs/                 # Documentation
```

## Making Changes

### Branch Naming

Use descriptive branch names:

- `feature/add-streaming-support`
- `fix/merge-keys-bug`
- `docs/update-readme`
- `perf/optimize-scanner`

### Commit Messages

Follow conventional commits format:

```
type(scope): description

[optional body]

[optional footer]
```

Examples:

- `feat(parser): add support for complex keys`
- `fix(scanner): resolve quote style detection bug`
- `docs(readme): update installation instructions`
- `perf(emitter): optimize string allocation`

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

### Code Guidelines

#### General Principles

1. **Safety First**: No `unsafe` code without exceptional justification
2. **Performance**: Zero-copy parsing where possible
3. **Error Handling**: Comprehensive error context with positions
4. **Documentation**: All public APIs must be documented
5. **Testing**: All features must have tests

#### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` or `gofumpt` for formatting
- Address all `clippy` warnings
- Prefer explicit error types over `anyhow` in library code
- Use `IndexMap` for preserving key order in mappings

#### YAML Implementation

- Full YAML 1.2 specification compliance
- Round-trip preservation (parse → serialize → parse)
- Secure by default (no code execution)
- Clear error messages with position information

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run specific test suite
go test -run TestMerge ./...
go test -run TestDiff ./...

# Run with coverage
make coverage

# Generate coverage HTML report
make coverage-html

# Run benchmarks
make bench
```

### Writing Tests

#### Unit Tests

Place unit tests in separate `_test.go` files:

```go
package yaml

import (
    "testing"
)

func TestFeatureName(t *testing.T) {
    // Test implementation
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

#### Integration Tests

Place integration tests with `_integration_test.go` suffix:

```go
package yaml_test

import (
    "testing"
    "github.com/elioetibr/golang-yaml-advanced"
)

func TestRealWorldScenario(t *testing.T) {
    tree, err := yaml.UnmarshalYAML([]byte(yamlContent))
    // Test real-world usage
}
```

#### Test Categories

- **Basic Functionality**: Core parsing and generation
- **YAML 1.2 Compliance**: Specification conformance
- **Advanced Features**: Anchors, merge keys, complex structures
- **Error Handling**: Invalid input and edge cases
- **Performance**: Benchmarks and stress tests
- **Round-trip**: Parse → serialize → parse consistency

### Test Data

For YAML test files, use the `tests/data/` directory:

- Valid YAML files: `tests/data/valid/`
- Invalid YAML files: `tests/data/invalid/`
- Edge cases: `tests/data/edge_cases/`

## Code Style

### Formatting

```bash
# Format code
go fmt ./...
# or with gofumpt for stricter formatting
gofumpt -l -w .

# Check formatting
gofmt -l .
```

### Linting

```bash
# Run go vet
go vet ./...

# Run golangci-lint (comprehensive)
golangci-lint run

# Run specific linters
golangci-lint run --enable gofumpt,gosec,goconst
```

### Documentation

```bash
# View documentation
go doc -all

# Serve documentation locally
godoc -http=:6060
# Then visit http://localhost:6060/pkg/github.com/elioetibr/golang-yaml-advanced/
```

#### Documentation Guidelines

- Document all public APIs with examples
- Include error conditions in documentation
- Provide usage examples for complex features
- Link to relevant YAML specification sections

## Submitting Changes

### Before Submitting

1. **Rebase** your branch on the latest `main`:

   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run the full test suite**:

   ```bash
   make test
   make lint
   gofmt -l .
   ```

3. **Update documentation** if needed
4. **Add tests** for new features
5. **Update CHANGELOG.md** for significant changes

### Pull Request Process

1. **Create a descriptive title**:
    - `Add support for streaming YAML parsing`
    - `Fix memory leak in large document processing`

2. **Write a detailed description**:

   ```markdown
   ## Summary
   Brief description of changes

   ## Changes
   - Specific change 1
   - Specific change 2

   ## Testing
   - [ ] All existing tests pass
   - [ ] Added new tests for feature X
   - [ ] Manual testing completed

   ## Performance Impact
   Description of any performance changes

   ## Breaking Changes
   List any breaking changes
   ```

3. **Link related issues**: Use `Fixes #123` or `Closes #123`

### Review Process

- At least one maintainer review required
- All CI checks must pass
- Address all review feedback
- Keep the PR focused and atomic

## Security

For security vulnerabilities, please see our [Security Policy](SECURITY.md). Do not report security issues through public GitHub issues.

## Community

### Getting Help

- **GitHub Issues**: For bugs and feature requests
- **Discussions**: For questions and general discussion
- **Documentation**: Check the docs first

### Ways to Contribute

#### Code Contributions

- Bug fixes
- New features
- Performance improvements
- Code refactoring

#### Documentation

- API documentation improvements
- Usage examples
- Tutorials and guides
- README enhancements

#### Testing

- Test case improvements
- Performance benchmarks
- YAML specification compliance tests
- Edge case testing

#### Community

- Help answer questions
- Review pull requests
- Participate in discussions
- Share usage examples

## Release Process

(For maintainers)

1. Update version in `go.mod`
2. Update `CHANGELOG.md`
3. Create release PR
4. After merge, tag release: `git tag v1.2.3`
5. Push tags: `git push --tags`
6. GitHub Actions will handle publishing

## Development Tips

### Performance Testing

```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark with memory allocation stats
go test -bench=BenchmarkMerge -benchmem

# Compare benchmark results
benchstat old.txt new.txt
```

### Memory Testing

```bash
# Profile memory usage
go test -memprofile mem.prof -bench=.
go tool pprof mem.prof

# Check for race conditions
go test -race ./...
```

### Fuzzing

```bash
# Create fuzz test (Go 1.18+)
go test -fuzz=FuzzUnmarshal -fuzztime=10s

# Run existing fuzz tests
go test -fuzz=Fuzz ./...
```

## Architecture Notes

### Key Components

1. **Parser** (`yaml.go`): Core YAML parsing with yaml.v3
2. **NodeTree** (`yaml.go`): Tree structure for YAML documents
3. **Merger** (`merge.go`): Deep merging with comment preservation
4. **Differ** (`diff.go`): Structural comparison engine
5. **Transformer** (`transform.go`): DSL for YAML transformations
6. **Query Engine** (`query.go`): XPath-like querying

### Design Principles

- **Comment Preservation**: Never lose comments during operations
- **Format Preservation**: Maintain style, quotes, and empty lines
- **Extensibility**: Plugin-based architecture for features
- **Performance**: Efficient handling of large YAML files
- **Compatibility**: Work seamlessly with yaml.v3

## License

By contributing, you agree that your contributions will be dual licensed under either:

- Apache License, Version 2.0 ([LICENSE-APACHE-2.0](LICENSE-APACHE-2.0))
- MIT license ([LICENSE-MIT](LICENSE-MIT))

at your option.

---

Thank you for contributing to golang-yaml-advanced! 🎉