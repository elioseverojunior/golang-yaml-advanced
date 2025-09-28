# Changelog

All notable changes to the Golang YAML Advanced library will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-09-28

### Added
- **Flexible Merge Strategy**: New `MergeFlexible()` function supporting mixed input types
  - Handles NodeTree and interface{} inputs in any combination
  - Automatic type conversion with comment preservation
  - Smart merge logic based on input types
- **Enhanced Conversion Utilities**:
  - `ConvertToNodeTree()` - Universal conversion function
  - `MergeFlexibleToNodeTree()` - Always returns NodeTree result
  - `MergeFlexibleToYAML()` - Direct YAML output from flexible merge
  - `MergeInterfaces()` - Interface{} merging with Go type system
- **Production-Ready Testing**: Comprehensive test suite with 580+ test cases
  - Real-world scenarios (Kubernetes, Helm configurations)
  - Edge case coverage and error handling validation
  - Type conversion and interface compatibility tests

### Enhanced
- **API Documentation**: Complete documentation for all new flexible merge functions
- **README Examples**: Added practical examples for configuration management use cases
- **Type Safety**: Robust type assertion and conversion with proper error handling

### Use Cases
- Configuration management with mixed data sources
- Kubernetes manifest merging with environment overrides
- Helm chart values composition from multiple sources
- CI/CD pipeline configuration assembly

## [1.0.1] - 2025-09-28

### Added
- Comprehensive unit tests for all public functions (441+ test cases)
- Complete API documentation in `API.md`
- Examples documentation in `examples/README.md`
- Test coverage reporting (achieved 91% coverage)
- `yaml_complete_test.go` with exhaustive unit tests
- Testing section in README with coverage details

### Enhanced
- Test coverage increased from ~80% to 91%
- All edge cases now covered with dedicated tests
- Nil handling validated across all functions
- Error scenarios comprehensively tested

### Documentation
- Added detailed API reference with examples
- Created examples README with usage patterns
- Updated main README with testing guidelines
- Added troubleshooting guide for common issues

### Quality Improvements
- Every public function now has dedicated unit tests
- All node operations thoroughly tested
- Merge and diff operations validated with multiple scenarios
- Comment preservation verified across all operations

## [1.0.0] - 2025-09-27 [RETRACTED]

### Initial Release
- Core YAML parsing with comment preservation
- Node-based tree structure for YAML manipulation
- Merge functionality with comment preservation
- Diff operations for comparing YAML structures
- Large integer preservation (no scientific notation)
- Multi-document YAML support
- Anchor and alias handling
- Style preservation (literal, folded, quoted)
- YAML 1.2.2 specification compliance

### Core Features
- `UnmarshalYAML()` - Parse YAML while preserving all metadata
- `MergeTrees()` - Intelligent merging with override support
- `DiffTrees()` - Comprehensive comparison of YAML structures
- `Node.Walk()` - Tree traversal with visitor pattern
- `Node.Find()` - Query nodes with predicates
- `Node.Clone()` - Deep copying with metadata preservation

### Known Issues (Fixed in 1.0.1)
- Limited test coverage for edge cases
- Missing documentation for some utility functions

## [Unreleased]

### Planned
- Streaming parser for large files (>100MB)
- Transformation DSL for complex operations
- XPath-like query language
- Schema validation with JSON Schema support
- Performance optimizations for deep nesting
- Concurrent processing support
- YAML 1.3 specification support when released

## Migration Guide

### From v1.0.0 to v1.0.1
No breaking changes. This is a quality improvement release with better test coverage.

### From yaml.v3 to golang-yaml-advanced

1. Replace imports:
```go
// Before
import "gopkg.in/yaml.v3"

// After
import "github.com/elioetibr/golang-yaml-advanced"
```

2. Use enhanced parsing for comment preservation:
```go
// Before
var node yaml.Node
yaml.Unmarshal(data, &node)

// After
tree, err := golang_yaml_advanced.UnmarshalYAML(data)
```

3. Leverage new features:
- Use `MergeTrees()` for configuration merging
- Use `DiffTrees()` for change detection
- Use `Node.Walk()` for tree traversal
- Comments are automatically preserved

## Support

For issues and feature requests, please use the [GitHub issue tracker](https://github.com/elioetibr/golang-yaml-advanced/issues).

## Contributors

- Elio Severo Junior (@elioseverojunior) - Initial implementation and maintenance

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.