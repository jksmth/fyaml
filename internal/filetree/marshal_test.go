package filetree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jksmth/fyaml/internal/logger"
	"go.yaml.in/yaml/v4"
)

// marshal_test.go contains tests for shared marshaling code (marshal.go).

func TestOptions_Log(t *testing.T) {
	t.Run("nil options", func(t *testing.T) {
		log := (*Options)(nil).log()
		if log == nil {
			t.Error("Options.log() should return a logger, not nil")
		}
		log.Debugf("test")
		log.Warnf("test")
	})

	t.Run("nil logger", func(t *testing.T) {
		opts := &Options{
			EnableIncludes:  false,
			Chroot:        "/tmp",
			ConvertBooleans: false,
			Logger:          nil,
		}
		log := opts.log()
		if log == nil {
			t.Error("Options.log() should return a logger, not nil")
		}
		log.Debugf("test")
		log.Warnf("test")
	})

	t.Run("with logger", func(t *testing.T) {
		var buf strings.Builder
		testLogger := logger.New(&buf, true)
		opts := &Options{
			EnableIncludes:  false,
			Chroot:        "/tmp",
			ConvertBooleans: false,
			Logger:          testLogger,
		}
		log := opts.log()
		if log != testLogger {
			t.Error("Options.log() should return the configured logger")
		}
		log.Debugf("test message")
		if !strings.Contains(buf.String(), "test message") {
			t.Error("Options.log() should return the configured logger that actually logs")
		}
	})
}

func TestMarshal_DefaultsToCanonical(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `key: value`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		Chroot: tmpDir,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	_, isNode := result.(*yaml.Node)
	if isNode {
		t.Error("Default mode should be canonical (interface{}), not preserve (*yaml.Node)")
	}
}

func TestMarshal_NilOptsDefaultsToCanonical(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `key: value`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(nil)
	assertNoError(t, err)

	_, isNode := result.(*yaml.Node)
	if isNode {
		t.Error("nil opts should default to canonical mode")
	}
}

// TestMarshal_InvalidYAML tests error handling for invalid YAML in both modes.
func TestMarshal_InvalidYAML(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
	}{
		{"canonical", ModeCanonical},
		{"preserve", ModePreserve},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTestDir(t, map[string]string{
				"another_dir/another_dir_file.yml": "1some: in: valid: yaml",
			}, nil)
			anotherDirFile := filepath.Join(tmpDir, "another_dir", "another_dir_file.yml")

			tree, err := NewTree(tmpDir)
			assertNoError(t, err)

			opts := &Options{
				Chroot: tmpDir,
				Mode:     tt.mode,
				Logger:   logger.Nop(),
			}

			_, err = tree.Marshal(opts)
			assertErrorContains(t, err, anotherDirFile)
			errStr := strings.ToLower(err.Error())
			if !strings.Contains(errStr, "yaml") && !strings.Contains(errStr, "json") {
				t.Errorf("yaml.Marshal() error = %v, expected YAML/JSON parsing error", err)
			}
		})
	}
}

// TestMarshal_FormatYAMLError_ParserError tests error formatting for parser errors in both modes.
func TestMarshal_FormatYAMLError_ParserError(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
	}{
		{"canonical", ModeCanonical},
		{"preserve", ModePreserve},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTestDir(t, map[string]string{
				"test.yml": "key: [unclosed",
			}, nil)
			testFile := filepath.Join(tmpDir, "test.yml")

			tree, err := NewTree(tmpDir)
			assertNoError(t, err)

			opts := &Options{
				Chroot: tmpDir,
				Mode:     tt.mode,
				Logger:   logger.Nop(),
			}

			_, err = tree.Marshal(opts)
			assertErrorContains(t, err, "YAML/JSON syntax error in")
			if !strings.Contains(err.Error(), testFile) {
				t.Errorf("Expected file path in error message, got: %s", err.Error())
			}
			if !strings.Contains(err.Error(), ":") {
				t.Errorf("Expected position information (line:column) in error message, got: %s", err.Error())
			}
		})
	}
}

// TestMarshal_WithIncludes_ErrorCases tests include error handling in both modes.
func TestMarshal_WithIncludes_ErrorCases(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
	}{
		{"canonical", ModeCanonical},
		{"preserve", ModePreserve},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTestDir(t, map[string]string{
				"test.yml": `command: <<include(nonexistent.sh)>>`,
			}, nil)

			absDir, err := filepath.Abs(tmpDir)
			assertNoError(t, err)

			opts := &Options{
				EnableIncludes: true,
				Chroot:       absDir,
				Mode:           tt.mode,
				Logger:         logger.Nop(),
			}

			tree, err := NewTree(tmpDir)
			assertNoError(t, err)

			testNode := findNodeByName(t, tree, "test.yml")
			if testNode == nil {
				t.Fatal("Could not find test.yml node")
			}

			var err2 error
			if tt.mode == ModePreserve {
				_, err2 = testNode.marshalLeafNode(opts)
			} else {
				_, err2 = testNode.marshalLeafMap(opts)
			}
			assertErrorContains(t, err2, "could not open")
		})
	}
}

// TestMarshal_FileReadError tests file read error handling in both modes.
func TestMarshal_FileReadError(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
	}{
		{"canonical", ModeCanonical},
		{"preserve", ModePreserve},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTestDir(t, map[string]string{
				"test.yml": "key: value",
			}, nil)
			yamlFile := filepath.Join(tmpDir, "test.yml")

			tree, err := NewTree(tmpDir)
			assertNoError(t, err)

			testNode := findNodeByName(t, tree, "test.yml")
			if testNode == nil {
				t.Fatal("Could not find test.yml node")
			}

			if err := os.Chmod(yamlFile, 0000); err != nil {
				t.Fatalf("Failed to chmod file: %v", err)
			}
			t.Cleanup(func() {
				_ = os.Chmod(yamlFile, 0600)
			})

			opts := &Options{
				Chroot: tmpDir,
				Mode:     tt.mode,
				Logger:   logger.Nop(),
			}

			var err2 error
			if tt.mode == ModePreserve {
				_, err2 = testNode.marshalLeafNode(opts)
			} else {
				_, err2 = testNode.marshalLeafMap(opts)
			}
			assertErrorContains(t, err2, "failed to read file")
		})
	}
}

// TestMarshal_NonMapTypeError tests non-map type error handling in both modes.
func TestMarshal_NonMapTypeError(t *testing.T) {
	tests := []struct {
		name string
		mode Mode
	}{
		{"canonical", ModeCanonical},
		{"preserve", ModePreserve},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := createTestDir(t, map[string]string{
				"dir/scalar.yml": "just a string",
				"dir/map.yml":    "key: value",
			}, nil)

			tree, err := NewTree(tmpDir)
			assertNoError(t, err)

			dirNode := findNodeByName(t, tree, "dir")
			if dirNode == nil {
				t.Fatal("Could not find dir node")
			}

			opts := &Options{
				Chroot: tmpDir,
				Mode:     tt.mode,
				Logger:   logger.Nop(),
			}

			_, err = dirNode.Marshal(opts)
			assertErrorContains(t, err, "expected map")
			assertErrorContains(t, err, "scalar.yml")
		})
	}
}
