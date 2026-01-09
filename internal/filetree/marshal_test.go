package filetree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jksmth/fyaml/internal/logger"
	"go.yaml.in/yaml/v4"
)

// marshal_test.go contains tests for shared marshaling code (marshal.go).

func TestFormatYAMLError_TypeError(t *testing.T) {
	testFile := "/test/path/file.yml"

	tests := []struct {
		name     string
		typeErr  *yaml.TypeError
		wantLine bool
	}{
		{
			name: "with line/column info",
			typeErr: &yaml.TypeError{
				Errors: []*yaml.UnmarshalError{
					{Line: 2, Column: 5, Err: fmt.Errorf("cannot decode !!str as !!int")},
				},
			},
			wantLine: true,
		},
		{
			name: "without line/column info",
			typeErr: &yaml.TypeError{
				Errors: []*yaml.UnmarshalError{
					{Line: 0, Column: 0, Err: fmt.Errorf("type conversion error")},
				},
			},
			wantLine: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formattedErr := formatYAMLError(tt.typeErr, testFile)
			errStr := formattedErr.Error()

			assertErrorContains(t, formattedErr, "YAML/JSON type errors in")
			assertErrorContains(t, formattedErr, testFile)

			lines := strings.Split(errStr, "\n")
			if len(lines) < 2 || lines[1] == "" {
				t.Fatalf("Expected error detail line, got: %s", errStr)
			}
			detailLine := lines[1]
			hasLineFormat := strings.Contains(detailLine, "line ")
			if tt.wantLine != hasLineFormat {
				t.Errorf("wantLine=%v but hasLineFormat=%v, detail: %q", tt.wantLine, hasLineFormat, detailLine)
			}
			if !strings.HasPrefix(detailLine, "  ") {
				t.Errorf("Expected error line to start with '  ', got: %q", detailLine)
			}
		})
	}
}

func TestNormalizeYAML11Booleans(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantBool bool
	}{
		{"unquoted on", "value: on", "true", true},
		{"unquoted off", "value: off", "false", true},
		{"unquoted yes", "value: yes", "true", true},
		{"unquoted no", "value: no", "false", true},
		{"unquoted y", "value: y", "true", true},
		{"unquoted n", "value: n", "false", true},
		{"unquoted Y", "value: Y", "true", true},
		{"unquoted N", "value: N", "false", true},
		{"unquoted Yes", "value: Yes", "true", true},
		{"unquoted No", "value: No", "false", true},
		{"unquoted ON", "value: ON", "true", true},
		{"unquoted OFF", "value: OFF", "false", true},
		{"unquoted On", "value: On", "true", true},
		{"unquoted Off", "value: Off", "false", true},
		{"unquoted YES", "value: YES", "true", true},
		{"unquoted NO", "value: NO", "false", true},
		{"double-quoted on", `value: "on"`, "on", false},
		{"double-quoted yes", `value: "yes"`, "yes", false},
		{"single-quoted on", "value: 'on'", "on", false},
		{"single-quoted yes", "value: 'yes'", "yes", false},
		{"unquoted true", "value: true", "true", true},
		{"unquoted false", "value: false", "false", true},
		{"unquoted norway", "value: norway", "norway", false},
		{"unquoted yesss", "value: yesss", "yesss", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tt.input), &node); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			normalizeYAML11Booleans(&node)

			var result map[string]interface{}
			if err := node.Decode(&result); err != nil {
				t.Fatalf("Failed to decode node: %v", err)
			}

			val := result["value"]
			if tt.wantBool {
				boolVal, ok := val.(bool)
				if !ok {
					t.Errorf("Expected bool, got %T: %v", val, val)
					return
				}
				expectedBool := tt.expected == "true"
				if boolVal != expectedBool {
					t.Errorf("Got %v, want %v", boolVal, expectedBool)
				}
			} else {
				strVal, ok := val.(string)
				if !ok {
					t.Errorf("Expected string, got %T: %v", val, val)
					return
				}
				if strVal != tt.expected {
					t.Errorf("Got %q, want %q", strVal, tt.expected)
				}
			}
		})
	}
}

func TestNormalizeYAML11Booleans_Nested(t *testing.T) {
	input := `
config:
  enabled: on
  debug: off
  settings:
    verbose: yes
    quiet: "no"
  items:
    - active: on
    - active: "off"
`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(input), &node); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	normalizeYAML11Booleans(&node)

	var result map[string]interface{}
	if err := node.Decode(&result); err != nil {
		t.Fatalf("Failed to decode node: %v", err)
	}

	config := result["config"].(map[string]interface{})

	if enabled, ok := config["enabled"].(bool); !ok || !enabled {
		t.Errorf("enabled should be bool true, got %T: %v", config["enabled"], config["enabled"])
	}
	if debug, ok := config["debug"].(bool); !ok || debug {
		t.Errorf("debug should be bool false, got %T: %v", config["debug"], config["debug"])
	}

	settings := config["settings"].(map[string]interface{})
	if verbose, ok := settings["verbose"].(bool); !ok || !verbose {
		t.Errorf("verbose should be bool true, got %T: %v", settings["verbose"], settings["verbose"])
	}
	if quiet, ok := settings["quiet"].(string); !ok || quiet != "no" {
		t.Errorf("quiet should be string 'no', got %T: %v", settings["quiet"], settings["quiet"])
	}

	items := config["items"].([]interface{})
	item0 := items[0].(map[string]interface{})
	if active, ok := item0["active"].(bool); !ok || !active {
		t.Errorf("items[0].active should be bool true, got %T: %v", item0["active"], item0["active"])
	}
	item1 := items[1].(map[string]interface{})
	if active, ok := item1["active"].(string); !ok || active != "off" {
		t.Errorf("items[1].active should be string 'off', got %T: %v", item1["active"], item1["active"])
	}
}

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
			PackRoot:        "/tmp",
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
			PackRoot:        "/tmp",
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
		PackRoot: tmpDir,
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
				PackRoot: tmpDir,
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
				PackRoot: tmpDir,
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
				PackRoot:       absDir,
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
				_, err2 = testNode.marshalLeafPreserve(opts)
			} else {
				_, err2 = testNode.marshalLeaf(opts)
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
				PackRoot: tmpDir,
				Mode:     tt.mode,
				Logger:   logger.Nop(),
			}

			var err2 error
			if tt.mode == ModePreserve {
				_, err2 = testNode.marshalLeafPreserve(opts)
			} else {
				_, err2 = testNode.marshalLeaf(opts)
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
				PackRoot: tmpDir,
				Mode:     tt.mode,
				Logger:   logger.Nop(),
			}

			_, err = dirNode.Marshal(opts)
			assertErrorContains(t, err, "expected a map")
			assertErrorContains(t, err, "scalar.yml")
		})
	}
}
