// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// TestCategory represents a category of tests
type TestCategory struct {
	Name        string
	Path        string
	Description string
}

var testCategories = []TestCategory{
	{
		Name:        "Main Application",
		Path:        ".",
		Description: "Main application setup and integration tests",
	},
	{
		Name:        "Models",
		Path:        "./backend/model",
		Description: "Data models validation and business logic",
	},
	{
		Name:        "Repositories",
		Path:        "./backend/repository",
		Description: "Database operations and queries",
	},
	{
		Name:        "Handlers",
		Path:        "./backend/handler",
		Description: "HTTP request handlers and API endpoints",
	},
	{
		Name:        "Middleware",
		Path:        "./backend/middleware",
		Description: "Authentication and authorization middleware",
	},
	{
		Name:        "Services",
		Path:        "./backend/service",
		Description: "Business logic and external integrations",
	},
	{
		Name:        "Utils",
		Path:        "./backend/utils",
		Description: "Utility functions and helpers",
	},
}

func main() {
	fmt.Println("🚀 PeopleFlow Comprehensive Test Suite")
	fmt.Println("=====================================")
	fmt.Println()

	startTime := time.Now()
	totalPassed := 0
	totalFailed := 0
	failedCategories := []string{}

	// Run tests for each category
	for _, category := range testCategories {
		fmt.Printf("📦 Testing %s\n", category.Name)
		fmt.Printf("   %s\n", category.Description)
		fmt.Printf("   Path: %s\n", category.Path)
		
		// Check if path exists
		if _, err := os.Stat(strings.TrimPrefix(category.Path, "./")); os.IsNotExist(err) {
			fmt.Printf("   ⚠️  Path does not exist, skipping...\n\n")
			continue
		}
		
		// Run tests with coverage
		cmd := exec.Command("go", "test", "-v", "-cover", category.Path)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			fmt.Printf("   ❌ FAILED\n")
			failedCategories = append(failedCategories, category.Name)
			totalFailed++
			
			// Print error output
			fmt.Printf("   Error output:\n")
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "FAIL") || strings.Contains(line, "Error") {
					fmt.Printf("     %s\n", line)
				}
			}
		} else {
			fmt.Printf("   ✅ PASSED\n")
			totalPassed++
			
			// Extract coverage percentage
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "coverage:") {
					fmt.Printf("   %s\n", line)
					break
				}
			}
		}
		
		fmt.Println()
	}

	// Summary
	duration := time.Since(startTime)
	fmt.Println("=====================================")
	fmt.Println("📊 Test Summary")
	fmt.Printf("   Total Categories: %d\n", len(testCategories))
	fmt.Printf("   ✅ Passed: %d\n", totalPassed)
	fmt.Printf("   ❌ Failed: %d\n", totalFailed)
	fmt.Printf("   ⏱️  Duration: %s\n", duration.Round(time.Second))
	
	if len(failedCategories) > 0 {
		fmt.Println("\n❌ Failed Categories:")
		for _, cat := range failedCategories {
			fmt.Printf("   - %s\n", cat)
		}
	}

	// Run overall coverage report
	fmt.Println("\n📈 Generating Overall Coverage Report...")
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "./...")
	if err := cmd.Run(); err == nil {
		// Generate HTML report
		cmd = exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
		if err := cmd.Run(); err == nil {
			fmt.Println("   ✅ Coverage report generated: coverage.html")
		}
		
		// Show coverage summary
		cmd = exec.Command("go", "tool", "cover", "-func=coverage.out")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				lastLine := lines[len(lines)-2] // Last line before empty line
				if strings.Contains(lastLine, "total:") {
					fmt.Printf("   📊 %s\n", lastLine)
				}
			}
		}
	}

	// Exit with appropriate code
	if totalFailed > 0 {
		fmt.Println("\n❌ Some tests failed!")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ All tests passed!")
		os.Exit(0)
	}
}