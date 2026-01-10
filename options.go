package fyaml

import "fmt"

// Format specifies the output format for the packed document.
type Format string

const (
	// FormatYAML outputs YAML format (default).
	FormatYAML Format = "yaml"
	// FormatJSON outputs JSON format.
	FormatJSON Format = "json"
)

// Mode controls the output behavior of the packed document.
type Mode string

const (
	// ModeCanonical produces canonical output with sorted keys and no comments (default).
	ModeCanonical Mode = "canonical"
	// ModePreserve preserves authored key order and comments.
	ModePreserve Mode = "preserve"
)

// MergeStrategy controls how maps are merged when multiple files contribute to the same key.
type MergeStrategy string

const (
	// MergeShallow uses "last wins" behavior - later values completely replace earlier ones (default).
	MergeShallow MergeStrategy = "shallow"
	// MergeDeep recursively merges nested maps, only replacing values at the leaf level.
	MergeDeep MergeStrategy = "deep"
)

// PackOptions configures how a directory is packed into a single document.
type PackOptions struct {
	// Dir is the directory to pack (required).
	Dir string

	// Format specifies the output format. Defaults to FormatYAML if empty.
	Format Format

	// Mode controls output behavior. Defaults to ModeCanonical if empty.
	Mode Mode

	// MergeStrategy controls merge behavior. Defaults to MergeShallow if empty.
	MergeStrategy MergeStrategy

	// EnableIncludes processes !include, !include-text, and <<include()>> directives.
	EnableIncludes bool

	// ConvertBooleans converts unquoted YAML 1.1 booleans (on/off, yes/no) to YAML 1.2 (true/false).
	ConvertBooleans bool

	// Indent is the number of spaces for indentation. Defaults to 2 if zero.
	Indent int

	// Logger is an optional logger for verbose output. If nil, no logging is performed.
	Logger Logger
}

// ParseFormat parses a format string and returns the corresponding Format.
// Returns an error if the format is invalid.
func ParseFormat(s string) (Format, error) {
	switch s {
	case "yaml":
		return FormatYAML, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("%w: %s (must be 'yaml' or 'json')", ErrInvalidFormat, s)
	}
}

// ParseMode parses a mode string and returns the corresponding Mode.
// Returns an error if the mode is invalid.
func ParseMode(s string) (Mode, error) {
	switch s {
	case "canonical":
		return ModeCanonical, nil
	case "preserve":
		return ModePreserve, nil
	default:
		return "", fmt.Errorf("%w: %s (must be 'canonical' or 'preserve')", ErrInvalidMode, s)
	}
}

// ParseMergeStrategy parses a merge strategy string and returns the corresponding MergeStrategy.
// Returns an error if the strategy is invalid.
func ParseMergeStrategy(s string) (MergeStrategy, error) {
	switch s {
	case "shallow":
		return MergeShallow, nil
	case "deep":
		return MergeDeep, nil
	default:
		return "", fmt.Errorf("%w: %s (must be 'shallow' or 'deep')", ErrInvalidMergeStrategy, s)
	}
}

// CheckOptions configures how Check compares content.
// Zero value provides default behavior (exact byte comparison, YAML format).
type CheckOptions struct {
	// Format specifies the format used for normalization.
	// Defaults to FormatYAML if empty.
	// Used to normalize empty expected content to match format-specific empty output.
	Format Format

	// Future options can be added here without breaking changes.
	// For example: IgnoreWhitespace bool
}
