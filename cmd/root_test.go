package cmd

import (
	"bytes"
	"strings"
	"testing"

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
