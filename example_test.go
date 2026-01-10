package fyaml_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jksmth/fyaml"
)

func ExamplePack() {
	// Pack a directory with default options (YAML output, canonical mode)
	result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
		Dir: "./config",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result))
}

func ExamplePack_withOptions() {
	// Pack with all options specified
	result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
		Dir:             "./config",
		Format:          fyaml.FormatJSON,
		Mode:            fyaml.ModePreserve,
		MergeStrategy:   fyaml.MergeDeep,
		EnableIncludes:  true,
		ConvertBooleans: true,
		Indent:          4,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result))
}

func ExamplePack_withLogger() {
	// Pack with a logger for verbose output
	logger := fyaml.NewLogger(os.Stderr, true)

	result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
		Dir:    "./config",
		Logger: logger,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result))
}

func ExamplePack_minimal() {
	// Minimal usage - only directory required, all other options use defaults
	result, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
		Dir: "./config",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(result))
}

func ExampleFormat() {
	// Use FormatYAML for YAML output (default)
	_ = fyaml.FormatYAML

	// Use FormatJSON for JSON output
	_ = fyaml.FormatJSON
}

func ExampleMode() {
	// Use ModeCanonical for sorted keys, no comments (default)
	_ = fyaml.ModeCanonical

	// Use ModePreserve for authored order and comments
	_ = fyaml.ModePreserve
}

func ExampleMergeStrategy() {
	// Use MergeShallow for "last wins" behavior (default)
	_ = fyaml.MergeShallow

	// Use MergeDeep for recursive nested map merging
	_ = fyaml.MergeDeep
}

func ExamplePack_errorHandling() {
	_, err := fyaml.Pack(context.Background(), fyaml.PackOptions{
		Dir: "",
	})
	if err != nil {
		if errors.Is(err, fyaml.ErrDirectoryRequired) {
			// Handle directory required error
			fmt.Println("Directory is required")
		} else if errors.Is(err, fyaml.ErrInvalidFormat) {
			// Handle invalid format error
			fmt.Println("Invalid format specified")
		}
	}
}

func ExampleParseFormat() {
	format, err := fyaml.ParseFormat("yaml")
	if err != nil {
		log.Fatal(err)
	}
	_ = format // Use format
}

func ExampleParseMode() {
	mode, err := fyaml.ParseMode("preserve")
	if err != nil {
		log.Fatal(err)
	}
	_ = mode // Use mode
}

func ExampleParseMergeStrategy() {
	strategy, err := fyaml.ParseMergeStrategy("deep")
	if err != nil {
		log.Fatal(err)
	}
	_ = strategy // Use strategy
}

func ExampleCheck() {
	generated := []byte("key: value\n")
	expected := []byte("key: value\n")

	err := fyaml.Check(generated, expected, fyaml.CheckOptions{Format: fyaml.FormatYAML})
	if err != nil {
		log.Fatal(err)
	}
	// Output matches
}

func ExampleCheck_mismatch() {
	generated := []byte("key: value\n")
	expected := []byte("key: different\n")

	err := fyaml.Check(generated, expected, fyaml.CheckOptions{Format: fyaml.FormatYAML})
	if errors.Is(err, fyaml.ErrCheckMismatch) {
		fmt.Println("Output mismatch detected")
	}
	// Output: Output mismatch detected
}

func ExampleCheck_emptyJSON() {
	// Empty JSON output is normalized to "null\n"
	generated := []byte("null\n")
	expected := []byte{} // Empty

	err := fyaml.Check(generated, expected, fyaml.CheckOptions{Format: fyaml.FormatJSON})
	if err != nil {
		log.Fatal(err)
	}
	// Empty expected is normalized, so it matches
}
