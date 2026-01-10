# API Reference

Complete reference for using fyaml as a Go library.

## Overview

Use the `fyaml` package to compile directory-structured YAML/JSON files into a single document from Go code. The API mirrors the CLI functionality: pass a directory path and options, get back the packed document as bytes.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jksmth/fyaml"
)

func main() {
    result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
        Dir: "./config",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(result))
}
```

## Functions

### `Pack`

```go
func Pack(ctx context.Context, opts PackOptions) ([]byte, error)
```

Compiles a directory of YAML/JSON files into a single document.

**Parameters:**

- `ctx` - Context for cancellation and timeout support
- `opts` - PackOptions configuring the packing operation

**Returns:**

- `[]byte` - The packed document as bytes
- `error` - Error if packing fails

**Behavior:**

- Safe for concurrent use by multiple goroutines (with separate PackOptions instances)
- Supports context cancellation
- Returns deterministic output based on directory structure
- Applies sensible defaults for all options

**Example:**

```go
result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
    Dir:    "./config",
    Format: fyaml.FormatYAML,
    Mode:   fyaml.ModeCanonical,
})
```

### `ParseFormat`

```go
func ParseFormat(s string) (Format, error)
```

Parses a format string and returns the corresponding Format type.

**Parameters:**

- `s` - Format string ("yaml" or "json")

**Returns:**

- `Format` - The parsed format constant
- `error` - Returns `ErrInvalidFormat` if the format is invalid

**Example:**

```go
format, err := fyaml.ParseFormat("json")
if err != nil {
    log.Fatal(err)
}
// format == fyaml.FormatJSON
```

### `ParseMode`

```go
func ParseMode(s string) (Mode, error)
```

Parses a mode string and returns the corresponding Mode type.

**Parameters:**

- `s` - Mode string ("canonical" or "preserve")

**Returns:**

- `Mode` - The parsed mode constant
- `error` - Returns `ErrInvalidMode` if the mode is invalid

**Example:**

```go
mode, err := fyaml.ParseMode("preserve")
if err != nil {
    log.Fatal(err)
}
// mode == fyaml.ModePreserve
```

### `ParseMergeStrategy`

```go
func ParseMergeStrategy(s string) (MergeStrategy, error)
```

Parses a merge strategy string and returns the corresponding MergeStrategy type.

**Parameters:**

- `s` - Strategy string ("shallow" or "deep")

**Returns:**

- `MergeStrategy` - The parsed strategy constant
- `error` - Returns `ErrInvalidMergeStrategy` if the strategy is invalid

**Example:**

```go
strategy, err := fyaml.ParseMergeStrategy("deep")
if err != nil {
    log.Fatal(err)
}
// strategy == fyaml.MergeDeep
```

### `Check`

```go
func Check(generated []byte, expected []byte, opts CheckOptions) error
```

Compares generated output with expected content. Returns `ErrCheckMismatch` if contents don't match.

**Parameters:**

- `generated` - The generated output bytes to check
- `expected` - The expected content bytes
- `opts` - Options for comparison behavior. Format defaults to FormatYAML if empty.

**Returns:**

- `error` - Returns `ErrCheckMismatch` if contents don't match, `nil` if they match

**Behavior:**

- Normalizes empty expected content based on opts.Format:
  - JSON format: empty expected is normalized to `"null\n"` (matches empty Pack() output)
  - YAML format: empty expected stays empty (matches empty Pack() output)
- Performs byte-by-byte comparison after normalization (whitespace differences are detected)
- Useful for programmatic validation in tests and CI/CD

**Example:**

```go
generated := []byte("key: value\n")
expected := []byte("key: value\n")

err := fyaml.Check(generated, expected, fyaml.CheckOptions{Format: fyaml.FormatYAML})
if err != nil {
    if errors.Is(err, fyaml.ErrCheckMismatch) {
        // Output doesn't match expected
    }
}
```

### `NewLogger`

```go
func NewLogger(w io.Writer, verbose bool) Logger
```

Creates a logger that writes to the specified io.Writer.

**Parameters:**

- `w` - Writer for log output (typically `os.Stderr`)
- `verbose` - If true, Debugf messages are shown

**Returns:**

- `Logger` - A logger instance

**Example:**

```go
logger := fyaml.NewLogger(os.Stderr, true)

result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
    Dir:    "./config",
    Logger: logger,
})
```

### `NopLogger`

```go
func NopLogger() Logger
```

Returns a no-op logger that discards all output.

**Returns:**

- `Logger` - A logger that performs no operations

**Example:**

```go
logger := fyaml.NopLogger()
// Use when you don't need logging but want to be explicit
```

## Types

### `PackOptions`

Configures how a directory is packed into a single document.

```go
type PackOptions struct {
    Dir             string        // Required: directory to pack
    Format          Format        // Output format (default: FormatYAML)
    Mode            Mode          // Output mode (default: ModeCanonical)
    MergeStrategy   MergeStrategy // Merge strategy (default: MergeShallow)
    EnableIncludes  bool          // Process include directives
    ConvertBooleans bool          // Convert YAML 1.1 booleans
    Indent          int           // Indentation spaces (default: 2)
    Logger          Logger        // Optional logger (default: no-op)
}
```

**Fields:**

- **Dir** (required) - The directory to pack. Must be a valid path to an existing directory.
- **Format** - Output format. Defaults to `FormatYAML` if empty.
- **Mode** - Output mode. Defaults to `ModeCanonical` if empty.
- **MergeStrategy** - Merge strategy. Defaults to `MergeShallow` if empty.
- **EnableIncludes** - If true, processes `!include`, `!include-text`, and `<<include()>>` directives.
- **ConvertBooleans** - If true, converts unquoted YAML 1.1 booleans (`on`/`off`, `yes`/`no`) to YAML 1.2 (`true`/`false`).
- **Indent** - Number of spaces for indentation. Defaults to 2 if zero. Must be at least 1.
- **Logger** - Optional logger for verbose output. If nil, no logging is performed.

**Example:**

```go
opts := fyaml.PackOptions{
    Dir:             "./config",
    Format:          fyaml.FormatJSON,
    Mode:            fyaml.ModePreserve,
    MergeStrategy:   fyaml.MergeDeep,
    EnableIncludes:  true,
    ConvertBooleans: true,
    Indent:          4,
    Logger:          fyaml.NewLogger(os.Stderr, true),
}
```

### `Format`

Specifies the output format for the packed document.

```go
type Format string
```

**Constants:**

- `FormatYAML` - YAML format output (default)
- `FormatJSON` - JSON format output

**Example:**

```go
opts := fyaml.PackOptions{
    Dir:    "./config",
    Format: fyaml.FormatJSON,
}
```

### `Mode`

Controls the output behavior of the packed document.

```go
type Mode string
```

**Constants:**

- `ModeCanonical` - Produces canonical output with sorted keys and no comments (default)
- `ModePreserve` - Preserves authored key order and comments

**Example:**

```go
opts := fyaml.PackOptions{
    Dir:  "./config",
    Mode: fyaml.ModePreserve,
}
```

### `MergeStrategy`

Controls how maps are merged when multiple files contribute to the same key.

```go
type MergeStrategy string
```

**Constants:**

- `MergeShallow` - Uses "last wins" behavior - later values completely replace earlier ones (default)
- `MergeDeep` - Recursively merges nested maps, only replacing values at the leaf level

**Example:**

```go
opts := fyaml.PackOptions{
    Dir:           "./config",
    MergeStrategy: fyaml.MergeDeep,
}
```

### `Logger`

Defines the logging interface for fyaml.

```go
type Logger interface {
    Debugf(format string, args ...interface{})
    Warnf(format string, args ...interface{})
}
```

**Methods:**

- **Debugf** - Logs verbose/debug information (shown when verbose enabled)
- **Warnf** - Logs warnings (always shown)

**Implementation:**

The package provides two implementations:

- `NewLogger()` - Creates a logger that writes to an io.Writer
- `NopLogger()` - Returns a no-op logger that discards all output

**Example:**

```go
type CustomLogger struct{}

func (l *CustomLogger) Debugf(format string, args ...interface{}) {
    // Custom debug logging
}

func (l *CustomLogger) Warnf(format string, args ...interface{}) {
    // Custom warning logging
}

logger := &CustomLogger{}
```

### `CheckOptions`

Configures how `Check` compares content.

```go
type CheckOptions struct {
    Format Format  // Format used for normalization (defaults to FormatYAML if empty)
    // Future options can be added here without breaking changes.
}
```

**Fields:**

- **Format** - Format used for normalization of empty expected content. Defaults to `FormatYAML` if empty.

**Future Extensibility:**

The struct is designed to allow adding options in the future without breaking changes. For example, options like `IgnoreWhitespace bool` could be added later.

**Example:**

```go
// With explicit format
err := fyaml.Check(generated, expected, fyaml.CheckOptions{
    Format: fyaml.FormatYAML,
})

// Default format (YAML)
err := fyaml.Check(generated, expected, fyaml.CheckOptions{})

// Future usage (when options are added):
// err := fyaml.Check(generated, expected, fyaml.CheckOptions{
//     Format:           fyaml.FormatYAML,
//     IgnoreWhitespace: true,
// })
```

## Errors

fyaml defines sentinel errors for programmatic error handling. Use `errors.Is()` to check for specific errors.

### Error Types

```go
var (
    ErrDirectoryRequired    = errors.New("directory is required")
    ErrInvalidFormat        = errors.New("invalid format")
    ErrInvalidMode          = errors.New("invalid mode")
    ErrInvalidMergeStrategy = errors.New("invalid merge strategy")
    ErrInvalidIndent        = errors.New("invalid indent")
    ErrCheckMismatch        = errors.New("output mismatch")
)
```

### Error Handling

```go
result, err := fyaml.Pack(ctx, opts)
if err != nil {
    if errors.Is(err, fyaml.ErrDirectoryRequired) {
        // Handle missing directory
    } else if errors.Is(err, fyaml.ErrInvalidFormat) {
        // Handle invalid format
    } else if errors.Is(err, fyaml.ErrInvalidMode) {
        // Handle invalid mode
    } else if errors.Is(err, fyaml.ErrInvalidMergeStrategy) {
        // Handle invalid merge strategy
    } else if errors.Is(err, fyaml.ErrInvalidIndent) {
        // Handle invalid indent
    } else {
        // Handle other errors (I/O, parsing, etc.)
    }
}

err = fyaml.Check(generated, expected, fyaml.CheckOptions{Format: fyaml.FormatYAML})
if err != nil {
    if errors.Is(err, fyaml.ErrCheckMismatch) {
        // Handle output mismatch
    }
}
```

### Error Details

- **ErrDirectoryRequired** - Returned when `Dir` is empty or not provided
- **ErrInvalidFormat** - Returned when `Format` is not `FormatYAML` or `FormatJSON`
- **ErrInvalidMode** - Returned when `Mode` is not `ModeCanonical` or `ModePreserve`
- **ErrInvalidMergeStrategy** - Returned when `MergeStrategy` is not `MergeShallow` or `MergeDeep`
- **ErrInvalidIndent** - Returned when `Indent` is less than 1
- **ErrCheckMismatch** - Returned when `Check()` finds differences between generated and expected content

## Examples

### Basic Usage

```go
result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
    Dir: "./config",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(result))
```

### With All Options

```go
result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
    Dir:             "./config",
    Format:          fyaml.FormatJSON,
    Mode:            fyaml.ModePreserve,
    MergeStrategy:   fyaml.MergeDeep,
    EnableIncludes:  true,
    ConvertBooleans: true,
    Indent:          4,
    Logger:          fyaml.NewLogger(os.Stderr, true),
})
```

### With Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := fyaml.Pack(ctx, fyaml.PackOptions{
    Dir: "./config",
})
```

### With Error Handling

```go
result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
    Dir: "./config",
})
if err != nil {
    if errors.Is(err, fyaml.ErrDirectoryRequired) {
        log.Fatal("Directory is required")
    } else if errors.Is(err, fyaml.ErrInvalidFormat) {
        log.Fatal("Invalid format specified")
    }
    log.Fatalf("Pack failed: %v", err)
}
```

### With Logging

```go
logger := fyaml.NewLogger(os.Stderr, true)

result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
    Dir:    "./config",
    Logger: logger,
})
```

### Parsing User Input

```go
formatStr := "json"
format, err := fyaml.ParseFormat(formatStr)
if err != nil {
    return fmt.Errorf("invalid format %q: %w", formatStr, err)
}

modeStr := "preserve"
mode, err := fyaml.ParseMode(modeStr)
if err != nil {
    return fmt.Errorf("invalid mode %q: %w", modeStr, err)
}

result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
    Dir:    "./config",
    Format: format,
    Mode:   mode,
})
```

### Validating Output

```go
// Pack and validate against expected output
result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
    Dir: "./config",
})
if err != nil {
    log.Fatal(err)
}

expected, err := os.ReadFile("expected.yml")
if err != nil {
    log.Fatal(err)
}

err = fyaml.Check(result, expected, fyaml.CheckOptions{Format: fyaml.FormatYAML})
if err != nil {
    if errors.Is(err, fyaml.ErrCheckMismatch) {
        log.Fatal("Generated output doesn't match expected")
    }
    log.Fatal(err)
}
```

## Concurrency

`Pack()` is safe for concurrent use by multiple goroutines, provided that:

- Each call uses a separate `PackOptions` instance
- If multiple goroutines share the same `Logger` instance, log output may interleave (this does not affect correctness, only log formatting)

**Example:**

```go
var wg sync.WaitGroup
for _, dir := range []string{"./config1", "./config2", "./config3"} {
    wg.Add(1)
    go func(d string) {
        defer wg.Done()
        result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
            Dir: d,
        })
        // Handle result...
    }(dir)
}
wg.Wait()
```

## Context Support

The `Pack()` function accepts a `context.Context` for cancellation and timeout support.

**Cancellation:**

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Cancel after some condition
go func() {
    time.Sleep(5 * time.Second)
    cancel()
}()

result, err := fyaml.Pack(ctx, opts)
if err != nil {
    if errors.Is(err, context.Canceled) {
        // Operation was canceled
    }
}
```

**Timeout:**

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := fyaml.Pack(ctx, opts)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Operation timed out
    }
}
```

## Defaults

When options are not specified, `Pack()` applies sensible defaults:

- **Format** - `FormatYAML`
- **Mode** - `ModeCanonical`
- **MergeStrategy** - `MergeShallow`
- **Indent** - `2`
- **Logger** - No-op logger (no output)

**Example:**

```go
// All of these are equivalent:
opts1 := fyaml.PackOptions{Dir: "./config"}
opts2 := fyaml.PackOptions{
    Dir:           "./config",
    Format:        fyaml.FormatYAML,
    Mode:          fyaml.ModeCanonical,
    MergeStrategy: fyaml.MergeShallow,
    Indent:        2,
}
```

## See Also

- [pkg.go.dev](https://pkg.go.dev/github.com/jksmth/fyaml) - Complete godoc documentation
- [Usage Guide](usage.md) - How fyaml works and common patterns
- [Examples](examples.md) - Detailed examples with outputs
- [Command Reference](reference.md) - CLI command and flag reference
