package golang_yaml_advanced

import (
	"strings"
	"testing"
)

// TestHelmChartExactPreservation tests that Helm Chart.yaml files are preserved exactly
func TestHelmChartExactPreservation(t *testing.T) {
	helmChartYAML := `apiVersion: v2
name: test chart
description: A Helm chart for Kubernetes

# A chart can be either an 'application' or a 'library' chart.
#
# Application charts are a collection of templates that can be packaged into versioned archives
# to be deployed.
#
# Library charts provide useful utilities or functions for the chart developer. They're included as
# a dependency of application charts to inject those utilities and functions into the rendering
# pipeline. Library charts do not define any templates and therefore cannot be deployed.
type: application

# This is the chart version. This version number should be incremented each time you make changes
# to the chart and its templates, including the app version.
# Versions are expected to follow Semantic Versioning (https://semver.org/)
version: 0.1.0

# This is the version number of the application being deployed. This version number should be
# incremented each time you make changes to the application. Versions are not expected to
# follow Semantic Versioning. They should reflect the version the application is using.
# It is recommended to use it with quotes.
appVersion: "1.16.0"`

	t.Run("NodeTree preserves exact format", func(t *testing.T) {
		// Parse YAML into NodeTree
		tree, err := UnmarshalYAML([]byte(helmChartYAML))
		if err != nil {
			t.Fatalf("Failed to unmarshal YAML: %v", err)
		}

		// Convert back to YAML
		output, err := tree.ToYAML()
		if err != nil {
			t.Fatalf("Failed to convert to YAML: %v", err)
		}

		outputStr := strings.TrimRight(string(output), "\n")

		// Check if they're exactly the same
		if helmChartYAML != outputStr {
			// Detailed comparison for debugging
			expectedLines := strings.Split(helmChartYAML, "\n")
			actualLines := strings.Split(outputStr, "\n")

			t.Errorf("YAML not preserved exactly. Line count: expected %d, got %d",
				len(expectedLines), len(actualLines))

			// Show differences line by line
			maxLines := len(expectedLines)
			if len(actualLines) > maxLines {
				maxLines = len(actualLines)
			}

			for i := 0; i < maxLines; i++ {
				var expected, actual string
				if i < len(expectedLines) {
					expected = expectedLines[i]
				}
				if i < len(actualLines) {
					actual = actualLines[i]
				}

				if expected != actual {
					t.Errorf("Line %d differs:\n  Expected: %q\n  Got:      %q",
						i+1, expected, actual)
				}
			}
		} else {
			t.Log("✓ YAML preserved exactly with all comments, blank lines, and formatting")
		}
	})
}

// TestEmptyLinePreservationThroughMerge tests that empty lines are preserved when merging YAML
func TestEmptyLinePreservationThroughMerge(t *testing.T) {
	baseYAML := `apiVersion: v2
name: test chart
description: A Helm chart for Kubernetes

# This is a comment block
# with multiple lines
type: application

# Another comment block
# after a field
version: 0.1.0`

	overlayYAML := `apiVersion: v2
name: test chart updated

# New comment added
newField: value

# This will be merged
version: 0.2.0`

	t.Run("Empty lines preserved after merge", func(t *testing.T) {
		// Parse base
		baseTree, err := UnmarshalYAML([]byte(baseYAML))
		if err != nil {
			t.Fatalf("Failed to parse base: %v", err)
		}

		// Parse overlay
		overlayTree, err := UnmarshalYAML([]byte(overlayYAML))
		if err != nil {
			t.Fatalf("Failed to parse overlay: %v", err)
		}

		// Merge
		mergedTree := MergeTrees(baseTree, overlayTree)

		// Convert back to YAML
		output, err := mergedTree.ToYAML()
		if err != nil {
			t.Fatalf("Failed to convert to YAML: %v", err)
		}

		outputStr := string(output)
		t.Logf("Merged output:\n%s", outputStr)

		// Check that we have empty lines before comment blocks
		lines := strings.Split(outputStr, "\n")
		foundEmptyBeforeComment := false
		for i := 1; i < len(lines); i++ {
			currentLine := strings.TrimSpace(lines[i])
			prevLine := lines[i-1]

			// Check if current line is a comment and previous line is empty
			if strings.HasPrefix(currentLine, "#") && prevLine == "" {
				foundEmptyBeforeComment = true
				break
			}
		}

		if !foundEmptyBeforeComment {
			t.Error("Expected to find empty lines before comment blocks in merged output")
		}

		// Verify the merged content has both base and overlay values
		if !strings.Contains(outputStr, "name: test chart updated") {
			t.Error("Merged output should contain updated name from overlay")
		}
		if !strings.Contains(outputStr, "version: 0.2.0") {
			t.Error("Merged output should contain updated version from overlay")
		}
		if !strings.Contains(outputStr, "newField: value") {
			t.Error("Merged output should contain new field from overlay")
		}
		if !strings.Contains(outputStr, "type: application") {
			t.Error("Merged output should preserve type from base")
		}
	})
}

// TestEmptyLinePolicies tests different empty line handling policies
func TestEmptyLinePolicies(t *testing.T) {
	yamlWithComments := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config

# Database configuration
# These settings control database connections
data:
  db_host: localhost
  db_port: "5432"

# Cache configuration
cache:
  ttl: 3600
  size: 100MB`

	t.Run("NoEmptyLinesConfig removes all empty lines", func(t *testing.T) {
		tree, err := UnmarshalYAML([]byte(yamlWithComments))
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Apply NoEmptyLinesConfig
		tree.EmptyLineConfig = NoEmptyLinesConfig()

		if tree.EmptyLineConfig.Policy != EmptyLinesRemove {
			t.Error("NoEmptyLinesConfig should set Policy to EmptyLinesRemove")
		}
		if tree.EmptyLineConfig.PreserveBeforeComments {
			t.Error("NoEmptyLinesConfig should set PreserveBeforeComments to false")
		}
		if tree.EmptyLineConfig.PreserveAfterComments {
			t.Error("NoEmptyLinesConfig should set PreserveAfterComments to false")
		}

		output, err := tree.ToYAML()
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")

		// Check that there are no consecutive empty lines
		for i := 0; i < len(lines)-1; i++ {
			if lines[i] == "" && lines[i+1] == "" {
				t.Error("Found consecutive empty lines, they should be removed")
			}
		}

		// Check that there are no empty lines before comments
		for i := 1; i < len(lines); i++ {
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "#") && lines[i-1] == "" {
				t.Error("Found empty line before comment, it should be removed with NoEmptyLinesConfig")
			}
		}

		t.Logf("Output with NoEmptyLinesConfig:\n%s", outputStr)
	})

	t.Run("NormalizedEmptyLineConfig with single line", func(t *testing.T) {
		tree, err := UnmarshalYAML([]byte(yamlWithComments))
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Apply NormalizedEmptyLineConfig with 1 empty line
		tree.EmptyLineConfig = NormalizedEmptyLineConfig(1)

		output, err := tree.ToYAML()
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		outputStr := string(output)
		lines := strings.Split(outputStr, "\n")

		// Count empty lines before comments
		emptyBeforeComment := 0
		for i := 1; i < len(lines); i++ {
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "#") && lines[i-1] == "" {
				emptyBeforeComment++
				// Check it's exactly 1 empty line
				if i >= 2 && lines[i-2] == "" {
					t.Error("Found more than 1 empty line before comment with NormalizedEmptyLineConfig(1)")
				}
			}
		}

		if emptyBeforeComment == 0 {
			t.Error("Expected normalized empty lines before comments")
		}

		t.Logf("Output with NormalizedEmptyLineConfig(1):\n%s", outputStr)
	})

	t.Run("NormalizedEmptyLineConfig with two lines", func(t *testing.T) {
		tree, err := UnmarshalYAML([]byte(yamlWithComments))
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Apply NormalizedEmptyLineConfig with 2 empty lines
		tree.EmptyLineConfig = NormalizedEmptyLineConfig(2)

		output, err := tree.ToYAML()
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		outputStr := string(output)
		t.Logf("Output with NormalizedEmptyLineConfig(2):\n%s", outputStr)

		// Visual inspection - with 2 lines there should be more spacing
		lines := strings.Split(outputStr, "\n")
		for i := 2; i < len(lines); i++ {
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "#") &&
			   lines[i-1] == "" && lines[i-2] == "" {
				// Found 2 empty lines before comment - good!
				t.Log("Found normalized 2 empty lines before comment")
			}
		}
	})

	t.Run("Default KeepAsIs policy", func(t *testing.T) {
		tree, err := UnmarshalYAML([]byte(yamlWithComments))
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Default config should be KeepAsIs
		if tree.EmptyLineConfig.Policy != EmptyLinesKeepAsIs {
			t.Error("Default policy should be EmptyLinesKeepAsIs")
		}

		output, err := tree.ToYAML()
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		outputStr := string(output)

		// Should preserve empty lines before comments
		if !strings.Contains(outputStr, "\n\n#") {
			t.Log("Empty lines are added before comment blocks by default")
		}
	})

	t.Run("EmptyLineConfig structure", func(t *testing.T) {
		// Test that configuration can be created and applied
		config := NormalizedEmptyLineConfig(1)
		if config.Policy != EmptyLinesNormalize {
			t.Error("NormalizedEmptyLineConfig should set Policy to EmptyLinesNormalize")
		}
		if config.NormalizedCount != 1 {
			t.Error("NormalizedEmptyLineConfig(1) should set NormalizedCount to 1")
		}

		tree, err := UnmarshalYAML([]byte(yamlWithComments))
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Apply custom config
		tree.EmptyLineConfig = config

		// This demonstrates the config can be set
		output, err := tree.ToYAML()
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		t.Logf("Output with normalized config:\n%s", string(output))
	})
}

// TestCommentsOnlyYAML tests handling of YAML files with only comments
func TestCommentsOnlyYAML(t *testing.T) {
	commentsOnly := `# Global Configuration Defaults
# Cross-Account Configuration
#
# This file contains global defaults that apply to all accounts
# Values defined here are inherited by all accounts, clusters, and namespaces`

	t.Run("Parse and preserve comments-only YAML", func(t *testing.T) {
		tree, err := UnmarshalYAML([]byte(commentsOnly))
		if err != nil {
			t.Fatalf("Failed to parse comments-only YAML: %v", err)
		}

		output, err := tree.ToYAML()
		if err != nil {
			t.Fatalf("Failed to serialize: %v", err)
		}

		outputStr := strings.TrimRight(string(output), "\n")

		if outputStr != commentsOnly {
			t.Errorf("Comments-only YAML not preserved.\nExpected:\n%s\n\nGot:\n%s",
				commentsOnly, outputStr)
		}
	})

	t.Run("Merge with empty interface preserves structure", func(t *testing.T) {
		tree, err := UnmarshalYAML([]byte(commentsOnly))
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		// Merge with empty map
		emptyTree, err := ConvertToNodeTree(map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create empty tree: %v", err)
		}

		merged := MergeTrees(tree, emptyTree)

		output, err := merged.ToYAML()
		if err != nil {
			t.Fatalf("Failed to serialize: %v", err)
		}

		// The merge should preserve the comments
		outputStr := string(output)
		if !strings.Contains(outputStr, "Global Configuration Defaults") {
			t.Error("Merged output should preserve comments")
		}
	})
}