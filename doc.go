// Package fyaml provides a programmatic API for compiling directory-structured
// YAML/JSON files into a single document.
//
// fyaml compiles a directory tree of YAML or JSON files into a single
// deterministic document. Directory names become map keys, file names
// (without extension) become nested keys, and files starting with @ merge
// their contents into the parent directory.
//
// Example:
//
//	package main
//
//	import (
//		"context"
//		"fmt"
//		"log"
//
//		"github.com/jksmth/fyaml"
//	)
//
//	func main() {
//		result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
//			Dir: "./config",
//		})
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println(string(result))
//	}
//
// Error Handling:
//
// The package defines sentinel errors for programmatic error handling:
//   - ErrDirectoryRequired
//   - ErrInvalidFormat
//   - ErrInvalidMode
//   - ErrInvalidMergeStrategy
//   - ErrInvalidIndent
//   - ErrCheckMismatch
//
// Use errors.Is() to check for specific errors:
//
//	result, err := fyaml.Pack(ctx, opts)
//	if err != nil {
//		if errors.Is(err, fyaml.ErrInvalidFormat) {
//			// Handle invalid format
//		}
//	}
//
//	err := fyaml.Check(generated, expected, fyaml.CheckOptions{Format: fyaml.FormatYAML})
//	if err != nil {
//		if errors.Is(err, fyaml.ErrCheckMismatch) {
//			// Handle output mismatch
//		}
//	}
//
// For more examples, see the examples in the test files.
package fyaml
