//go:build mage
// +build mage

package main

import (
	"os"
	"path/filepath"

	// mage:import
	build "github.com/grafana/grafana-plugin-sdk-go/build"
)

// Default configures the default target for the mage build tool.
var Default = build.BuildAll

// BuildAll builds the backend plugin for all supported platforms.
func BuildAll() error {
	return build.BuildAll()
}

// Build builds the backend plugin for the current platform.
func Build() error {
	return build.Build()
}

// Test runs the Go test suite.
func Test() error {
	return build.Test()
}

// Clean removes build artifacts.
func Clean() error {
	targets := []string{"dist", "gpx_gaia_adql", "gpx_gaia_adql.exe"}
	for _, t := range targets {
		absPath := filepath.Join(".", t)
		if err := os.RemoveAll(absPath); err != nil {
			return err
		}
	}
	return nil
}

// Lint runs golangci-lint on the Go source.
func Lint() error {
	return build.Lint()
}
