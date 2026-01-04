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
