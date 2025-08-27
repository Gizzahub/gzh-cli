// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	apprunner "github.com/Gizzahub/gzh-cli/internal/apprunner"
)

var version = "dev"

func main() {
	// Create and run the application
	runner := apprunner.NewRunner(version)

	if err := runner.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
