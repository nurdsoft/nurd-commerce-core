//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/filter_coverage.go <input_coverage_file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	tempFile := inputFile + ".tmp"

	patternsMap := map[string]struct{}{
		"main.go":               {},
		"cmd/.*":                {},
		"endpoints.go":          {},
		"module.go":             {},
		"mock_":                 {},
		"scripts/.*":            {},
		"server.go":             {},
		"config.go":             {},
		"interceptors/.*":       {},
		"shared/cfg/.*":         {},
		"shared/health/.*":      {},
		"shared/transport/.*":   {},
		"shared/meta/.*":        {},
		"shared/module/.*":      {},
		"internal/transport/.*": {},
		"shared/log/.*":         {},
		"sql_repository.go":     {},
	}

	// Compile regex patterns
	patterns := make([]string, 0, len(patternsMap))
	for pattern := range patternsMap {
		patterns = append(patterns, pattern)
	}
	regex := regexp.MustCompile(strings.Join(patterns, "|"))

	// Open input file
	inFile, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		os.Exit(1)
	}
	defer inFile.Close()

	// Create temporary output file
	outFile, err := os.Create(tempFile)
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		os.Exit(1)
	}
	defer outFile.Close()

	// Filter content
	scanner := bufio.NewScanner(inFile)
	writer := bufio.NewWriter(outFile)

	for scanner.Scan() {
		line := scanner.Text()
		if !regex.MatchString(line) {
			if _, err := writer.WriteString(line + "\n"); err != nil {
				fmt.Println("Error writing to temporary file:", err)
				os.Exit(1)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input file:", err)
		os.Exit(1)
	}

	writer.Flush()

	// Replace original file with the filtered version
	if err := os.Rename(tempFile, inputFile); err != nil {
		fmt.Println("Error renaming temporary file to original file:", err)
		os.Exit(1)
	}
}
