# Security Guide for golang-yaml-advanced

## Overview

golang-yaml-advanced implements comprehensive security measures to protect against common YAML parsing vulnerabilities, including the billion laughs attack, resource exhaustion, and malformed document attacks.

## Built-in Protection

### 1. Resource Limits

golang-yaml-advanced enforces configurable resource limits to prevent resource exhaustion attacks:

- **Max Depth**: Limits nesting depth of collections (default: 1000)
- **Max Anchors**: Limits number of anchors per document (default: 10,000)
- **Max Document Size**: Limits total document size in bytes (default: 100MB)
- **Max String Length**: Limits individual string length (default: 10MB)
- **Max Alias Depth**: Limits alias expansion depth (default: 100)
- **Max Collection Size**: Limits items in a single collection (default: 1,000,000)
- **Max Complexity Score**: Limits overall document complexity (default: 1,000,000)

### 2. Protection Against Billion Laughs Attack

The billion laughs attack (exponential entity expansion) is prevented through:

- Alias expansion depth tracking
- Cyclic reference detection
- Complexity scoring for nested structures
- Collection size limits

Example of prevented attack:

```yaml

# This would expand exponentially without protection
a: &a ["lol", "lol", "lol", "lol", "lol", "lol", "lol", "lol", "lol"]
b: &b [*a, *a, *a, *a, *a, *a, *a, *a, *a]
c: &c [*b, *b, *b, *b, *b, *b, *b, *b, *b]

# ... would expand to 9^n items
```

### 3. Cyclic Reference Detection

golang-yaml-advanced detects and prevents cyclic alias references:

```yaml

# This cyclic reference is detected and rejected
a: &a
  b: *b
b: &b
  a: *a
```

## Configuration Presets

### Strict Mode (Untrusted Input)

```go
import "github.com/elioetibr/golang-yaml-advanced"

// Parse with strict limits for untrusted input
tree, err := yaml.UnmarshalYAMLWithLimits(data, yaml.StrictLimits())
if err != nil {
    log.Printf("Failed to parse YAML: %v", err)
}
```

Strict limits:

- Max depth: 50
- Max anchors: 100
- Max document size: 1MB
- Max string length: 64KB
- Max alias depth: 5
- Max collection size: 10,000
- Timeout: 5 seconds

### Secure Mode

```go
// Use secure defaults for production
tree, err := yaml.UnmarshalYAMLWithLimits(data, yaml.SecureLimits())
```

Balanced security for production use with reasonable limits.

### Permissive Mode (Trusted Input)

```go
// Higher limits for trusted sources
tree, err := yaml.UnmarshalYAMLWithLimits(data, yaml.PermissiveLimits())
```

Higher limits for trusted sources:

- Max depth: 10,000
- Max anchors: 100,000
- Max document size: 1GB
- Max string length: 100MB

### Custom Limits

```go
import (
    "time"
    "github.com/elioetibr/golang-yaml-advanced"
)

limits := &yaml.Limits{
    MaxDepth:        100,
    MaxAnchors:      500,
    MaxDocumentSize: 5 * 1024 * 1024, // 5MB
    Timeout:         10 * time.Second,
}

tree, err := yaml.UnmarshalYAMLWithLimits(data, limits)
```

## Best Practices

### 1. Always Use Limits for Untrusted Input

Never parse untrusted YAML without resource limits:

```go
// BAD - No protection
tree, err := yaml.UnmarshalYAML(untrustedInput) // Vulnerable!

// GOOD - Protected parsing
tree, err := yaml.UnmarshalYAMLWithLimits(
    untrustedInput,
    yaml.SecureLimits(),
) // Protected
```

### 2. Choose Appropriate Limits

Select limits based on your use case:

- **Configuration files**: Use `strict()` limits
- **User-generated content**: Use `strict()` with custom timeout
- **Internal data**: Use `default()` or `permissive()`
- **Large datasets**: Use custom limits with appropriate sizes

### 3. Handle Errors Gracefully

Resource limit errors should be handled appropriately:

```go
tree, err := yaml.UnmarshalYAMLWithLimits(input, yaml.StrictLimits())
if err != nil {
    if yaml.IsResourceLimitError(err) {
        // Log security event
        log.Printf("Resource limit exceeded: %v", err)
        // Return safe error to user
        return nil, fmt.Errorf("document too complex")
    }
    // Handle other parsing errors
    return nil, fmt.Errorf("invalid YAML: %w", err)
}
```

### 4. Monitor Resource Usage

Use metrics to monitor actual usage:

```go
stats, err := yaml.ParseWithStats(data)
if err == nil {
    log.Printf("Max depth: %d, Anchors: %d, Nodes: %d",
        stats.MaxDepth, stats.AnchorCount, stats.NodeCount)
}
```

### 5. Validate Content After Parsing

Even with security limits, validate parsed content:

```go
tree, err := yaml.UnmarshalYAMLWithLimits(input, yaml.SecureLimits())
if err != nil {
    return nil, err
}

// Validate expected structure
if err := yaml.ValidateSchema(tree, schema); err != nil {
    return nil, fmt.Errorf("invalid document structure: %w", err)
}

// Sanitize values if needed
sanitized := yaml.SanitizeValues(tree)
```

## Common Attack Vectors

### 1. Exponential Expansion

- **Attack**: Nested aliases causing exponential growth
- **Protection**: Alias depth limits, complexity scoring

### 2. Deep Nesting

- **Attack**: Deeply nested structures causing stack overflow
- **Protection**: Max depth limits

### 3. Large Collections

- **Attack**: Huge arrays/maps consuming memory
- **Protection**: Collection size limits

### 4. Long Strings

- **Attack**: Multi-gigabyte strings
- **Protection**: String length limits

### 5. Anchor Bombs

- **Attack**: Thousands of anchors slowing parsing
- **Protection**: Anchor count limits

## Testing Security

Run security tests to verify protection:

```bash
go test -run TestSecurity ./...
```

Tests include:

- Billion laughs attack prevention
- Cyclic reference detection
- Resource limit enforcement
- Timeout handling

## Reporting Security Issues

If you discover a security vulnerability:

1. **Do not** create a public issue
2. Email security concerns to security@elio.eti.br
3. Include:
    - Description of the vulnerability
    - Steps to reproduce
    - Potential impact
    - Suggested fix (if any)

## Updates and Patches

- Security updates are released as patch versions
- Critical vulnerabilities trigger immediate releases
- Subscribe to security advisories for notifications

## Additional Resources

- [YAML Security Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/YAML_Security_Cheat_Sheet.html)
- [CVE Database for YAML](https://cve.mitre.org/cgi-bin/cvekey.cgi?keyword=yaml)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

---

*Last updated: September 2025*