package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVersionSubcommand(t *testing.T) {
	// Test that 'fyaml version' subcommand prints version information
	oldStdout := os.Stdout
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	rootCmd.SetArgs([]string{"version"})
	executeErr := rootCmd.Execute()

	_ = w.Close()

	if executeErr != nil {
		t.Errorf("Execute() with version command error = %v", executeErr)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Error("version subcommand produced no output")
	}
}

func TestVersionFlag(t *testing.T) {
	// Test that 'fyaml --version' flag prints version information
	oldStdout := os.Stdout
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	rootCmd.SetArgs([]string{"--version"})
	executeErr := rootCmd.Execute()

	_ = w.Close()

	if executeErr != nil {
		t.Errorf("Execute() with --version flag error = %v", executeErr)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Error("--version flag produced no output")
	}
}
