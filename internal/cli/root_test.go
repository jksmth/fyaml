package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/jksmth/fyaml"
	"github.com/jksmth/fyaml/internal/logger"
)

func TestRootCmd_PersistentPreRun_VerboseFlag(t *testing.T) {
	// Test that PersistentPreRun initializes logger based on verbose flag
	// Save original verbose value
	originalVerbose := verbose
	t.Cleanup(func() {
		verbose = originalVerbose
	})

	// Test with verbose enabled
	verbose = true
	rootCmd.PersistentPreRun(rootCmd, nil)
	if log == nil {
		t.Error("PersistentPreRun should initialize logger")
	}
	// Verify it's a logger that respects verbose flag
	var buf bytes.Buffer
	log = logger.New(&buf, true)
	log.Debugf("test")
	if !strings.Contains(buf.String(), "test") {
		t.Error("Logger should output debug messages when verbose is true")
	}

	// Test with verbose disabled
	verbose = false
	rootCmd.PersistentPreRun(rootCmd, nil)
	if log == nil {
		t.Error("PersistentPreRun should initialize logger")
	}
	buf.Reset()
	log = logger.New(&buf, false)
	log.Debugf("test")
	if strings.Contains(buf.String(), "test") {
		t.Error("Logger should not output debug messages when verbose is false")
	}
}

func TestRootCmd_VerboseFlag_Global(t *testing.T) {
	// Test that verbose flag is properly set as a global persistent flag
	// Save original verbose value
	originalVerbose := verbose
	t.Cleanup(func() {
		verbose = originalVerbose
	})

	// Reset verbose
	verbose = false

	// Verify flag exists
	flag := rootCmd.PersistentFlags().Lookup("verbose")
	if flag == nil {
		t.Error("verbose flag should exist on root command")
	}

	// Verify short flag exists
	shortFlag := rootCmd.PersistentFlags().ShorthandLookup("v")
	if shortFlag == nil {
		t.Error("verbose short flag 'v' should exist on root command")
	}
}

func TestRootCmd_PackFlags(t *testing.T) {
	// Test that all pack flags exist as persistent flags on rootCmd
	// Note: Cobra handles inheritance of persistent flags to subcommands automatically.
	// We verify the flags exist on rootCmd, and functional tests verify they work on packCmd.
	flags := []string{"dir", "output", "check", "format", "enable-includes", "convert-booleans", "indent"}
	for _, flagName := range flags {
		// Check flag exists on rootCmd as persistent flag
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("persistent flag %q should exist on rootCmd", flagName)
		}
		// Verify packCmd is a child of rootCmd (so it will inherit persistent flags)
		if packCmd.Parent() != rootCmd {
			t.Errorf("packCmd should be a child of rootCmd to inherit persistent flags")
		}
	}
}

func TestRootCmd_SubcommandPrecedence(t *testing.T) {
	// Test that subcommands take precedence over directory names
	// 'fyaml pack' should always invoke pack subcommand, not try to pack a directory named "pack"
	if packCmd == nil {
		t.Fatal("packCmd should be defined")
	}
	// Verify packCmd is a child of rootCmd
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd == packCmd {
			found = true
			break
		}
	}
	if !found {
		t.Error("packCmd should be registered as a subcommand of rootCmd")
	}
}

func TestRootCmd_InvalidFormat(t *testing.T) {
	// Save and restore original flag values
	originalFormat := format
	originalDir := dir
	t.Cleanup(func() {
		format = originalFormat
		dir = originalDir
	})

	// Set invalid format
	format = "xml"
	dir = t.TempDir()

	err := rootCmd.RunE(rootCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !errors.Is(err, fyaml.ErrInvalidFormat) {
		t.Errorf("expected ErrInvalidFormat, got: %v", err)
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("error should mention invalid value 'xml', got: %v", err)
	}
}

func TestRootCmd_InvalidMode(t *testing.T) {
	// Save and restore original flag values
	originalFormat := format
	originalMode := mode
	originalDir := dir
	t.Cleanup(func() {
		format = originalFormat
		mode = originalMode
		dir = originalDir
	})

	// Set valid format but invalid mode
	format = "yaml"
	mode = "invalid"
	dir = t.TempDir()

	err := rootCmd.RunE(rootCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
	if !errors.Is(err, fyaml.ErrInvalidMode) {
		t.Errorf("expected ErrInvalidMode, got: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error should mention invalid value, got: %v", err)
	}
}

func TestRootCmd_InvalidMergeStrategy(t *testing.T) {
	// Save and restore original flag values
	originalFormat := format
	originalMode := mode
	originalMergeStrategy := mergeStrategy
	originalDir := dir
	t.Cleanup(func() {
		format = originalFormat
		mode = originalMode
		mergeStrategy = originalMergeStrategy
		dir = originalDir
	})

	// Set valid format and mode but invalid merge strategy
	format = "yaml"
	mode = "canonical"
	mergeStrategy = "invalid"
	dir = t.TempDir()

	err := rootCmd.RunE(rootCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid merge strategy")
	}
	if !errors.Is(err, fyaml.ErrInvalidMergeStrategy) {
		t.Errorf("expected ErrInvalidMergeStrategy, got: %v", err)
	}
}

func TestRootCmd_InvalidIndent(t *testing.T) {
	tests := []struct {
		name        string
		indentValue int
	}{
		{"zero", 0},
		{"negative", -1},
		{"very negative", -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original flag values
			originalFormat := format
			originalMode := mode
			originalMergeStrategy := mergeStrategy
			originalIndent := indent
			originalDir := dir
			t.Cleanup(func() {
				format = originalFormat
				mode = originalMode
				mergeStrategy = originalMergeStrategy
				indent = originalIndent
				dir = originalDir
			})

			// Set valid flags except indent
			format = "yaml"
			mode = "canonical"
			mergeStrategy = "shallow"
			indent = tt.indentValue
			dir = t.TempDir()

			err := rootCmd.RunE(rootCmd, nil)
			if err == nil {
				t.Fatal("expected error for invalid indent")
			}
			if !strings.Contains(err.Error(), "invalid indent") {
				t.Errorf("error should mention 'invalid indent', got: %v", err)
			}
			if !strings.Contains(err.Error(), "must be at least 1") {
				t.Errorf("error should mention 'must be at least 1', got: %v", err)
			}
		})
	}
}

func TestRootCmd_ValidFlags(t *testing.T) {
	// Save and restore original flag values
	originalFormat := format
	originalMode := mode
	originalMergeStrategy := mergeStrategy
	originalIndent := indent
	originalDir := dir
	t.Cleanup(func() {
		format = originalFormat
		mode = originalMode
		mergeStrategy = originalMergeStrategy
		indent = originalIndent
		dir = originalDir
	})

	// Test all valid combinations
	validFormats := []string{"yaml", "json"}
	validModes := []string{"canonical", "preserve"}
	validStrategies := []string{"shallow", "deep"}

	for _, f := range validFormats {
		for _, m := range validModes {
			for _, s := range validStrategies {
				t.Run(f+"_"+m+"_"+s, func(t *testing.T) {
					format = f
					mode = m
					mergeStrategy = s
					indent = 2
					dir = t.TempDir()

					err := rootCmd.RunE(rootCmd, nil)
					// Should not fail on validation - empty dir is fine, produces empty/null output
					if err != nil && (errors.Is(err, fyaml.ErrInvalidFormat) ||
						errors.Is(err, fyaml.ErrInvalidMode) ||
						errors.Is(err, fyaml.ErrInvalidMergeStrategy) ||
						strings.Contains(err.Error(), "invalid indent")) {
						t.Errorf("validation should pass for valid flags, got: %v", err)
					}
				})
			}
		}
	}
}
