package main

import (
	"os"
	"testing"
)

// TestMainFileExists verifies that the main.go file exists
func TestMainFileExists(t *testing.T) {
	_, err := os.Stat("main.go")
	if err != nil {
		t.Fatalf("main.go file not found: %v", err)
	}
}

// TestMainPackageImport is a simple test to ensure the main package can be loaded
func TestMainPackageImport(t *testing.T) {
	// This test passes simply by being compiled, confirming the package is valid
}
