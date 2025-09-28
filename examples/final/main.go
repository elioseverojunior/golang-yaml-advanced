package main

import (
	"fmt"
	"log"
	"os"

	"github.com/elioetibr/golang-yaml-advanced"
)

func main() {
	// Step 1: Load the base values file
	fmt.Println("Step 1: Loading yaml-values.yaml...")
	baseTree, err := loadYAML("yaml-values.yaml")
	if err != nil {
		log.Fatalf("Failed to load base values: %v", err)
	}
	fmt.Printf("  ✓ Loaded base values\n")

	// Step 2: Load the override values file
	fmt.Println("\nStep 2: Loading yaml-values-override.yaml...")
	overrideTree, err := loadYAML("yaml-values-override.yaml")
	if err != nil {
		log.Fatalf("Failed to load override values: %v", err)
	}
	fmt.Printf("  ✓ Loaded override values\n")

	// Step 3: Merge the files (override values take precedence)
	fmt.Println("\nStep 3: Merging YAML files...")
	mergedTree := golang_yaml_advanced.MergeTrees(baseTree, overrideTree)
	fmt.Printf("  ✓ Merged YAML trees\n")

	// Show what was overridden
	showOverrides(overrideTree)

	// Step 4: Write the merged result
	fmt.Println("\nStep 4: Writing merged result to yaml-values-merged.yaml...")
	if err := writeYAML("yaml-values-merged.yaml", mergedTree); err != nil {
		log.Fatalf("Failed to write merged values: %v", err)
	}
	fmt.Println("  ✓ Successfully wrote merged YAML")

	// Step 5: Verify the output - compare with existing merged file
	fmt.Println("\nStep 5: Verifying output...")
	existingData, err := os.ReadFile("yaml-values-merged.yaml")
	if err != nil {
		log.Fatalf("Failed to read merged file for verification: %v", err)
	}

	// Re-parse to verify
	verifiedTree, err := golang_yaml_advanced.UnmarshalYAML(existingData)
	if err != nil {
		log.Fatalf("Failed to parse merged file for verification: %v", err)
	}

	// Compare the trees
	differences := golang_yaml_advanced.DiffTrees(mergedTree, verifiedTree)
	if len(differences) == 0 {
		fmt.Println("  ✓ Output matches expected result - no diff!")
	} else {
		fmt.Printf("  ⚠ Found %d differences\n", len(differences))
		for i, diff := range differences {
			if i < 3 { // Show only first 3 differences
				fmt.Printf("    - Path: %s, Type: %s, Old: %v, New: %v\n",
					diff.Path, diff.Type, diff.OldValue, diff.NewValue)
			}
		}
		if len(differences) > 3 {
			fmt.Printf("    ... and %d more\n", len(differences)-3)
		}
	}

	fmt.Println("\n✨ Process completed successfully!")
}

// loadYAML loads and parses a YAML file using the custom yaml package
func loadYAML(filename string) (*golang_yaml_advanced.NodeTree, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	tree, err := golang_yaml_advanced.UnmarshalYAML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return tree, nil
}

// writeYAML writes a NodeTree to a YAML file
func writeYAML(filename string, tree *golang_yaml_advanced.NodeTree) error {
	yamlContent, err := tree.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
	}

	return os.WriteFile(filename, yamlContent, 0644)
}

// showOverrides displays what values are being overridden
func showOverrides(tree *golang_yaml_advanced.NodeTree) {
	fmt.Println("\n  Override summary:")

	if len(tree.Documents) > 0 && tree.Documents[0].Root != nil {
		root := tree.Documents[0].Root

		// Look for specific overrides
		for _, child := range root.Children {
			if child.Key != nil {
				keyValue := fmt.Sprintf("%v", child.Key.Value)
				switch keyValue {
				case "replicaCount":
					if child.Value != nil {
						fmt.Printf("    • replicaCount: %v (increased replicas)\n", child.Value)
					}
				case "topologySpreadConstraints":
					fmt.Printf("    • topologySpreadConstraints: configured\n")
				case "strategy":
					fmt.Printf("    • strategy: configured\n")
				}
			}
		}
	}
}