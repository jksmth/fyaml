package fyaml

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to create PackOptions for tests.
func testOpts(dir string, format Format, enableIncludes, convertBooleans bool, mode Mode, mergeStrategy MergeStrategy) PackOptions {
	return PackOptions{
		Dir:             dir,
		Format:          format,
		EnableIncludes:  enableIncludes,
		ConvertBooleans: convertBooleans,
		Indent:          2,
		Mode:            mode,
		MergeStrategy:   mergeStrategy,
	}
}

// Helper to create a test directory with files.
func createTestDir(t *testing.T, files map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()
	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatalf("Failed to create directory %q: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create file %q: %v", fullPath, err)
		}
	}
	return tmpDir
}

func TestPack_Basic(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"entities/item1.yml": `entity:
  id: example1
  attributes:
    name: sample name`,
		"entities/item2.yml": `entity:
  id: example2
  attributes:
    name: another name`,
	})

	result, err := Pack(context.Background(), testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeShallow))
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("Pack() returned empty result")
	}

	// Verify basic structure
	resultStr := string(result)
	if !strings.Contains(resultStr, "entities:") {
		t.Error("result should contain 'entities:'")
	}
	if !strings.Contains(resultStr, "item1:") {
		t.Error("result should contain 'item1:'")
	}
	if !strings.Contains(resultStr, "item2:") {
		t.Error("result should contain 'item2:'")
	}
}

func TestPack_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	result, err := Pack(context.Background(), testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeShallow))
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Pack() for empty directory should return empty bytes, got %d bytes", len(result))
	}
}

func TestPack_EmptyDirectory_JSON(t *testing.T) {
	dir := t.TempDir()

	result, err := Pack(context.Background(), testOpts(dir, FormatJSON, false, false, ModeCanonical, MergeShallow))
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	expected := "null\n"
	if string(result) != expected {
		t.Errorf("Pack() for empty directory with JSON format = %q, want %q", string(result), expected)
	}
}

func TestPack_InvalidDirectory(t *testing.T) {
	_, err := Pack(context.Background(), testOpts("/nonexistent/path", FormatYAML, false, false, ModeCanonical, MergeShallow))
	if err == nil {
		t.Error("Pack() should return error for nonexistent directory")
	}
}

func TestPack_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	opts := testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeShallow)
	opts.Format = Format("invalid")

	_, err := Pack(context.Background(), opts)
	if err == nil {
		t.Error("Pack() should return error for invalid format")
	}
	if !errors.Is(err, ErrInvalidFormat) {
		t.Errorf("error should be ErrInvalidFormat, got: %v", err)
	}
}

func TestPack_InvalidMode(t *testing.T) {
	dir := t.TempDir()
	opts := testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeShallow)
	opts.Mode = Mode("invalid")

	_, err := Pack(context.Background(), opts)
	if err == nil {
		t.Error("Pack() should return error for invalid mode")
	}
	if !errors.Is(err, ErrInvalidMode) {
		t.Errorf("error should be ErrInvalidMode, got: %v", err)
	}
}

func TestPack_InvalidMergeStrategy(t *testing.T) {
	dir := t.TempDir()
	opts := testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeShallow)
	opts.MergeStrategy = MergeStrategy("invalid")

	_, err := Pack(context.Background(), opts)
	if err == nil {
		t.Error("Pack() should return error for invalid merge strategy")
	}
	if !errors.Is(err, ErrInvalidMergeStrategy) {
		t.Errorf("error should be ErrInvalidMergeStrategy, got: %v", err)
	}
}

func TestPack_InvalidIndent(t *testing.T) {
	dir := t.TempDir()
	opts := testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeShallow)
	opts.Indent = -1

	_, err := Pack(context.Background(), opts)
	if err == nil {
		t.Error("Pack() should return error for invalid indent")
	}
	if !errors.Is(err, ErrInvalidIndent) {
		t.Errorf("error should be ErrInvalidIndent, got: %v", err)
	}
}

func TestPack_EmptyDir(t *testing.T) {
	opts := PackOptions{}
	_, err := Pack(context.Background(), opts)
	if err == nil {
		t.Error("Pack() should return error for empty directory")
	}
	if !errors.Is(err, ErrDirectoryRequired) {
		t.Errorf("error should be ErrDirectoryRequired, got: %v", err)
	}
}

func TestPack_Defaults(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"test.yml": `key: value`,
	})

	// Test with zero values - should use defaults
	opts := PackOptions{
		Dir: dir,
	}

	result, err := Pack(context.Background(), opts)
	if err != nil {
		t.Fatalf("Pack() with defaults error = %v", err)
	}

	if len(result) == 0 {
		t.Error("Pack() should return non-empty result")
	}

	// Verify it's YAML (default format)
	resultStr := string(result)
	if !strings.Contains(resultStr, "key:") || !strings.Contains(resultStr, "value") {
		t.Errorf("result should be YAML, got: %q", resultStr)
	}
}

func TestPack_JSONFormat(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"test.yml": `key: value`,
	})

	result, err := Pack(context.Background(), testOpts(dir, FormatJSON, false, false, ModeCanonical, MergeShallow))
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	resultStr := string(result)
	// JSON should start with { or contain "key"
	if !strings.Contains(resultStr, `"key"`) {
		t.Errorf("result should be JSON, got: %q", resultStr)
	}
}

func TestPack_WithLogger(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"test.yml": `key: value`,
	})

	var logOutput strings.Builder
	log := NewLogger(&logOutput, true)

	opts := testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeShallow)
	opts.Logger = log

	_, err := Pack(context.Background(), opts)
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	// Logger should have been called
	if logOutput.Len() == 0 {
		t.Error("Logger should have been called")
	}
}

func TestPack_NilLogger(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"test.yml": `key: value`,
	})

	opts := testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeShallow)
	opts.Logger = nil

	// Should not panic
	result, err := Pack(context.Background(), opts)
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("Pack() should return non-empty result")
	}
}

func TestPack_ModePreserve(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"test.yml": `# Comment
zebra: value-z
alpha: value-a`,
	})

	result, err := Pack(context.Background(), testOpts(dir, FormatYAML, false, false, ModePreserve, MergeShallow))
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	resultStr := string(result)
	// Preserve mode should maintain order (zebra before alpha)
	zebraPos := strings.Index(resultStr, "zebra")
	alphaPos := strings.Index(resultStr, "alpha")
	if zebraPos == -1 || alphaPos == -1 {
		t.Fatalf("result should contain both keys, got: %q", resultStr)
	}
	if zebraPos > alphaPos {
		t.Error("Preserve mode should maintain key order (zebra before alpha)")
	}
}

func TestPack_MergeDeep(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"@base.yml": `config:
  setting1: value1
  nested:
    a: 1
    b: 2`,
		"@override.yml": `config:
  setting2: value2
  nested:
    c: 3`,
	})

	result, err := Pack(context.Background(), testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeDeep))
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	resultStr := string(result)
	// Deep merge should combine nested maps
	if !strings.Contains(resultStr, "setting1") {
		t.Error("Deep merge should preserve setting1")
	}
	if !strings.Contains(resultStr, "setting2") {
		t.Error("Deep merge should include setting2")
	}
	if !strings.Contains(resultStr, "a:") || !strings.Contains(resultStr, "b:") {
		t.Error("Deep merge should preserve nested values a and b")
	}
	if !strings.Contains(resultStr, "c:") {
		t.Error("Deep merge should include nested value c")
	}
}

func TestPack_EnableIncludes(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"shared/defaults.yml": `timeout: 30
retries: 3`,
		"entities/item1.yml": `entity:
  id: example1
  config: !include ../shared/defaults.yml`,
	})

	result, err := Pack(context.Background(), testOpts(dir, FormatYAML, true, false, ModeCanonical, MergeShallow))
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	resultStr := string(result)
	// Include should be processed
	if strings.Contains(resultStr, "!include") {
		t.Error("Include directive should be processed, not left as literal")
	}
	if !strings.Contains(resultStr, "timeout:") || !strings.Contains(resultStr, "retries:") {
		t.Error("Included content should be present in output")
	}
}

func TestPack_ConvertBooleans(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"config.yml": `enabled: on
active: yes`,
	})

	result, err := Pack(context.Background(), testOpts(dir, FormatYAML, false, true, ModeCanonical, MergeShallow))
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	resultStr := string(result)
	// Boolean conversion should convert on/yes to true
	if strings.Contains(resultStr, "enabled: on") {
		t.Error("Boolean conversion should convert 'on' to 'true'")
	}
	if strings.Contains(resultStr, "active: yes") {
		t.Error("Boolean conversion should convert 'yes' to 'true'")
	}
	if !strings.Contains(resultStr, "enabled: true") {
		t.Error("Result should contain 'enabled: true'")
	}
	if !strings.Contains(resultStr, "active: true") {
		t.Error("Result should contain 'active: true'")
	}
}

func TestPack_CustomIndent(t *testing.T) {
	dir := createTestDir(t, map[string]string{
		"test.yml": `key:
  nested: value`,
	})

	opts := testOpts(dir, FormatYAML, false, false, ModeCanonical, MergeShallow)
	opts.Indent = 4

	result, err := Pack(context.Background(), opts)
	if err != nil {
		t.Fatalf("Pack() error = %v", err)
	}

	// Verify indentation (4 spaces instead of 2)
	resultStr := string(result)
	lines := strings.Split(resultStr, "\n")
	found4SpaceIndent := false
	for _, line := range lines {
		if strings.HasPrefix(line, "    nested:") {
			found4SpaceIndent = true
			break
		}
	}
	if !found4SpaceIndent {
		t.Error("Custom indent of 4 should be used")
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Format
		wantErr bool
		errType error
	}{
		{"yaml", "yaml", FormatYAML, false, nil},
		{"json", "json", FormatJSON, false, nil},
		{"invalid", "invalid", "", true, ErrInvalidFormat},
		{"empty", "", "", true, ErrInvalidFormat},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat() = %v, want %v", got, tt.want)
			}
			if tt.wantErr && !errors.Is(err, tt.errType) {
				t.Errorf("ParseFormat() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Mode
		wantErr bool
		errType error
	}{
		{"canonical", "canonical", ModeCanonical, false, nil},
		{"preserve", "preserve", ModePreserve, false, nil},
		{"invalid", "invalid", "", true, ErrInvalidMode},
		{"empty", "", "", true, ErrInvalidMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseMode() = %v, want %v", got, tt.want)
			}
			if tt.wantErr && !errors.Is(err, tt.errType) {
				t.Errorf("ParseMode() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestParseMergeStrategy(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    MergeStrategy
		wantErr bool
		errType error
	}{
		{"shallow", "shallow", MergeShallow, false, nil},
		{"deep", "deep", MergeDeep, false, nil},
		{"invalid", "invalid", "", true, ErrInvalidMergeStrategy},
		{"empty", "", "", true, ErrInvalidMergeStrategy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMergeStrategy(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMergeStrategy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseMergeStrategy() = %v, want %v", got, tt.want)
			}
			if tt.wantErr && !errors.Is(err, tt.errType) {
				t.Errorf("ParseMergeStrategy() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestCheck_Matching(t *testing.T) {
	generated := []byte("key: value\n")
	expected := []byte("key: value\n")

	err := Check(generated, expected, CheckOptions{Format: FormatYAML})
	if err != nil {
		t.Errorf("Check() with matching content should not return error, got: %v", err)
	}
}

func TestCheck_Mismatch(t *testing.T) {
	generated := []byte("key: value\n")
	expected := []byte("key: different\n")

	err := Check(generated, expected, CheckOptions{Format: FormatYAML})
	if err == nil {
		t.Error("Check() with mismatched content should return error")
	}
	if !errors.Is(err, ErrCheckMismatch) {
		t.Errorf("Check() should return ErrCheckMismatch, got: %v", err)
	}
}

func TestCheck_EmptyYAML(t *testing.T) {
	// Empty YAML output is empty bytes
	generated := []byte{}
	expected := []byte{}

	err := Check(generated, expected, CheckOptions{Format: FormatYAML})
	if err != nil {
		t.Errorf("Check() with empty YAML should not return error, got: %v", err)
	}
}

func TestCheck_EmptyJSON(t *testing.T) {
	// Empty JSON output is normalized to "null\n"
	generated := []byte("null\n")
	expected := []byte{} // Empty

	err := Check(generated, expected, CheckOptions{Format: FormatJSON})
	if err != nil {
		t.Errorf("Check() with empty JSON (normalized) should not return error, got: %v", err)
	}
}

func TestCheck_EmptyJSONMismatch(t *testing.T) {
	// If generated is not "null\n" but expected is empty, should mismatch
	generated := []byte("key: value\n")
	expected := []byte{}

	err := Check(generated, expected, CheckOptions{Format: FormatJSON})
	if err == nil {
		t.Error("Check() with non-empty generated and empty expected should return error")
	}
	if !errors.Is(err, ErrCheckMismatch) {
		t.Errorf("Check() should return ErrCheckMismatch, got: %v", err)
	}
}

func TestCheck_JSONFormat(t *testing.T) {
	generated := []byte(`{"key": "value"}`)
	expected := []byte(`{"key": "value"}`)

	err := Check(generated, expected, CheckOptions{Format: FormatJSON})
	if err != nil {
		t.Errorf("Check() with matching JSON should not return error, got: %v", err)
	}
}
