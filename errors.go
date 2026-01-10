package fyaml

import "errors"

// Sentinel errors for programmatic error handling.
// Use errors.Is() to check for specific errors:
//
//	result, err := fyaml.Pack(ctx, opts)
//	if err != nil {
//		if errors.Is(err, fyaml.ErrInvalidFormat) {
//			// Handle invalid format error
//		}
//	}
var (
	// ErrDirectoryRequired is returned when Dir is empty or not provided.
	ErrDirectoryRequired = errors.New("directory is required")

	// ErrInvalidFormat is returned when Format is not FormatYAML or FormatJSON.
	ErrInvalidFormat = errors.New("invalid format")

	// ErrInvalidMode is returned when Mode is not ModeCanonical or ModePreserve.
	ErrInvalidMode = errors.New("invalid mode")

	// ErrInvalidMergeStrategy is returned when MergeStrategy is not MergeShallow or MergeDeep.
	ErrInvalidMergeStrategy = errors.New("invalid merge strategy")

	// ErrInvalidIndent is returned when Indent is less than 1.
	ErrInvalidIndent = errors.New("invalid indent")

	// ErrCheckMismatch is returned when Check() finds differences between
	// generated output and expected content.
	ErrCheckMismatch = errors.New("output mismatch")
)
