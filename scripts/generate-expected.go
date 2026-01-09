package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var allFlag = flag.Bool("all", false, "Generate for all testdata directories")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: go run scripts/generate-expected.go [--all] <mode> [testdata-dir]\n")
		fmt.Fprintf(os.Stderr, "  mode: canonical or preserve\n")
		fmt.Fprintf(os.Stderr, "  testdata-dir: path to testdata directory (e.g., testdata/simple)\n")
		os.Exit(1)
	}

	mode := args[0]
	if mode != "canonical" && mode != "preserve" {
		fmt.Fprintf(os.Stderr, "Error: mode must be 'canonical' or 'preserve'\n")
		os.Exit(1)
	}

	// Find fyaml binary
	fyamlPath := "./fyaml"
	if _, err := os.Stat(fyamlPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: fyaml binary not found. Run 'make build' first.\n")
		os.Exit(1)
	}

	if *allFlag {
		// Generate for all testdata directories
		testdataDirs := []string{
			"testdata/simple",
			"testdata/nested",
			"testdata/at-root",
			"testdata/at-files",
			"testdata/ordering",
			"testdata/anchors",
			"testdata/includes",
			"testdata/at-directories",
			"testdata/json-input",
		}

		for _, dir := range testdataDirs {
			if err := generateExpected(dir, mode, fyamlPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating %s for %s: %v\n", mode, dir, err)
				os.Exit(1)
			}
		}
		fmt.Println("Generated expected files for all testdata directories")
	} else {
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Error: testdata directory required when not using --all\n")
			os.Exit(1)
		}
		testdataDir := args[1]
		if err := generateExpected(testdataDir, mode, fyamlPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated expected-%s.yml for %s\n", mode, testdataDir)
	}
}

func generateExpected(testdataDir, mode, fyamlPath string) error {
	inputDir := filepath.Join(testdataDir, "input")
	if _, err := os.Stat(inputDir); os.IsNotExist(err) {
		return fmt.Errorf("input directory not found: %s", inputDir)
	}

	expectedFile := filepath.Join(testdataDir, fmt.Sprintf("expected-%s.yml", mode))

	// Build command
	cmd := exec.Command(fyamlPath, inputDir, "--mode", mode, "-o", expectedFile)

	// Handle includes flag for includes testdata
	if strings.Contains(testdataDir, "includes") {
		cmd.Args = append(cmd.Args, "--enable-includes")
	}

	// Run command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run fyaml: %v\nOutput: %s", err, string(output))
	}

	return nil
}
