//go:build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if _, err := exec.LookPath("mockgen"); err != nil {
		fmt.Println("Error: 'mockgen' command not found. Please install it before running this script.")
		os.Exit(1)
	}

	scriptName := filepath.Base(os.Args[0]) // Get the script name to avoid modifying itself

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing file: %v\n", err)
			return nil
		}

		// Skip directories and this script itself
		if info.IsDir() || filepath.Base(path) == scriptName {
			return nil
		}

		// Process files that start with "mock_"
		if strings.HasPrefix(filepath.Base(path), "mock_") {
			updateMock(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through files: %v\n", err)
		os.Exit(1)
	}
}

func updateMock(mockFilePath string) {
	// Open the mock file to find its package name
	file, err := os.Open(mockFilePath)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", mockFilePath, err)
		return
	}
	defer file.Close()

	// Scan for the package name
	scanner := bufio.NewScanner(file)
	packageRegex := regexp.MustCompile(`^package\s+(\S+)`)
	var packageName string

	for scanner.Scan() {
		line := scanner.Text()
		if matches := packageRegex.FindStringSubmatch(line); matches != nil {
			packageName = matches[1]
			break
		}
	}

	if packageName == "" {
		fmt.Printf("Package name not found in %s. Skipping.\n", mockFilePath)
		return
	}

	// Correctly derive the source file path
	sourceFilePath := getSourceFilePath(mockFilePath)

	// Ensure the source file exists before running mockgen
	if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
		fmt.Printf("Source file %s does not exist. Skipping.\n", sourceFilePath)
		return
	}

	fmt.Printf("Updating mock: %s\n", mockFilePath)

	// Run mockgen
	cmd := exec.Command("mockgen", "-source="+sourceFilePath, "-destination="+mockFilePath, "-package="+packageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running mockgen for %s: %v\n", mockFilePath, err)
	}
}

// getSourceFilePath correctly replaces only the first occurrence of "mock_"
func getSourceFilePath(mockFilePath string) string {
	dir, file := filepath.Split(mockFilePath)
	return filepath.Join(dir, strings.Replace(file, "mock_", "", 1))
}
