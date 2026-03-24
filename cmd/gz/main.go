// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	apprunner "github.com/gizzahub/gzh-cli/internal/apprunner"
	"github.com/gizzahub/gzh-cli/internal/version"
)

func main() {
	// Create and run the application
	runner := apprunner.NewRunner(version.Version)

	if err := runner.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
