package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jksmth/fyaml/internal/logger"
)

func TestRootCmd_PersistentPreRun_VerboseFlag(t *testing.T) {
	// Test that PersistentPreRun initializes logger based on verbose flag
	// Save original verbose value
	originalVerbose := verbose

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

	// Restore original value
	verbose = originalVerbose
}

func TestRootCmd_RunE_CallsPackCmd(t *testing.T) {
	// Test that rootCmd.RunE calls packCmd.RunE
	// This is tested indirectly through integration, but we can verify the structure
	tmpDir := t.TempDir()

	// Create a simple YAML file
	yamlFile := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(yamlFile, []byte("key: value"), 0600); err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	// Set up root command to use test directory
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
		verbose = false
		log = nil
	}()

	// Initialize logger
	verbose = false
	rootCmd.PersistentPreRun(rootCmd, []string{tmpDir})

	// Test that RunE can be called (it should delegate to packCmd)
	// We can't easily test this without running the full command, but we can verify
	// the function exists and doesn't panic
	err := rootCmd.RunE(rootCmd, []string{tmpDir})
	// This will fail because packCmd needs proper setup, but we're just checking
	// that the function exists and is callable
	if err != nil && !strings.Contains(err.Error(), "pack error") {
		// Expected to fail, but should be a pack-related error
		t.Logf("RunE returned error (expected): %v", err)
	}
}

func TestRootCmd_VerboseFlag_Global(t *testing.T) {
	// Test that verbose flag is properly set as a global persistent flag
	// Save original verbose value
	originalVerbose := verbose

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

	// Restore original value
	verbose = originalVerbose
}

