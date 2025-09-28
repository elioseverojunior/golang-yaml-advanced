package golang_yaml_advanced

// EmptyLinePolicy defines how to handle empty lines in YAML output
type EmptyLinePolicy int

const (
	// EmptyLinesKeepAsIs preserves the empty line formatting using heuristics
	EmptyLinesKeepAsIs EmptyLinePolicy = iota

	// EmptyLinesNormalize applies a consistent number of empty lines
	EmptyLinesNormalize

	// EmptyLinesRemove removes all empty lines (not yet implemented)
	EmptyLinesRemove
)

// EmptyLineConfig configures how empty lines are handled in YAML output
type EmptyLineConfig struct {
	// Policy determines the empty line handling strategy
	Policy EmptyLinePolicy

	// NormalizedCount is the number of empty lines to use when Policy is EmptyLinesNormalize
	// For example, 1 means single empty line between sections
	NormalizedCount int

	// PreserveBeforeComments when true, ensures empty lines before comment blocks
	PreserveBeforeComments bool

	// PreserveAfterComments when true, ensures empty lines after comment blocks
	PreserveAfterComments bool
}

// DefaultEmptyLineConfig returns the default configuration (keep as is)
func DefaultEmptyLineConfig() EmptyLineConfig {
	return EmptyLineConfig{
		Policy:                 EmptyLinesKeepAsIs,
		NormalizedCount:        1,
		PreserveBeforeComments: true,
		PreserveAfterComments:  false,
	}
}

// NormalizedEmptyLineConfig returns a config that normalizes to a specific count
func NormalizedEmptyLineConfig(count int) EmptyLineConfig {
	return EmptyLineConfig{
		Policy:                 EmptyLinesNormalize,
		NormalizedCount:        count,
		PreserveBeforeComments: true,
		PreserveAfterComments:  false,
	}
}

// NoEmptyLinesConfig returns a config that removes empty lines
func NoEmptyLinesConfig() EmptyLineConfig {
	return EmptyLineConfig{
		Policy:                 EmptyLinesRemove,
		NormalizedCount:        0,
		PreserveBeforeComments: false,
		PreserveAfterComments:  false,
	}
}