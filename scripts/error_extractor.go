//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// APIError represents the extracted error format
type APIError struct {
	ErrorCode  string `json:"error_code"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

var (
	// Pattern to detect module-specific error maps
	errorMapPattern = regexp.MustCompile(`"(?P<errorCode>[A-Z0-9_]+)"\s*:\s*\{\s*StatusCode:\s*http\.(?P<statusCode>\w+),\s*Message:\s*"(?P<message>[^"]+)"\s*\}`)
)

// Mapping HTTP status codes to integers
var httpStatusCodes = map[string]int{
	"StatusOK":                  200,
	"StatusCreated":             201,
	"StatusAccepted":            202,
	"StatusNoContent":           204,
	"StatusNotModified":         304,
	"StatusBadRequest":          400,
	"StatusUnauthorized":        401,
	"StatusForbidden":           403,
	"StatusNotFound":            404,
	"StatusConflict":            409,
	"StatusInternalServerError": 500,
	"StatusServiceUnavailable":  503,
}

func main() {
	projectDir := "./" // Set root directory to scan
	errors := []APIError{}

	// Walk through project directory but only process "errors.go" files
	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Process only files named "errors.go"
		if !info.IsDir() && strings.HasSuffix(path, "errors.go") {
			fileErrors, err := extractErrorsFromFile(path)
			if err != nil {
				fmt.Printf("Error processing %s: %v\n", path, err)
				return nil
			}
			errors = append(errors, fileErrors...)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error scanning project:", err)
		return
	}

	// Save extracted errors to a JSON file
	outputFile := "errors.json"
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating JSON file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print JSON
	if err := encoder.Encode(errors); err != nil {
		fmt.Println("Error writing JSON:", err)
		return
	}

	fmt.Printf("Extracted %d errors from `errors.go` files and saved to %s\n", len(errors), outputFile)
}

// extractErrorsFromFile reads a Go file and extracts errors based on the regex pattern
func extractErrorsFromFile(filename string) ([]APIError, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	matches := errorMapPattern.FindAllStringSubmatch(string(content), -1)
	var extractedErrors []APIError

	for _, match := range matches {
		statusCode, exists := httpStatusCodes[match[2]]
		if !exists {
			fmt.Printf("Warning: Unknown HTTP status %s in %s\n", match[2], filename)
			continue
		}

		extractedErrors = append(extractedErrors, APIError{
			ErrorCode:  match[1],
			StatusCode: statusCode,
			Message:    match[3],
		})
	}

	return extractedErrors, nil
}
