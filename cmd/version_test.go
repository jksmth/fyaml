package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	// Test that version command prints version information
	// Capture stdout
	oldStdout := os.Stdout
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	versionCmd.Run(versionCmd, []string{})

	_ = w.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Error("version command produced no output")
	}
}

func TestExecute_Version(t *testing.T) {
	// Test Execute function with version command via rootCmd
	// This exercises the Execute() -> rootCmd.Execute() path
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
		t.Error("Execute() with version command produced no output")
	}
}
