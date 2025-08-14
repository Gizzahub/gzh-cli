// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/Gizzahub/gzh-cli/internal/app"
)

var version = "dev"

func main() {
	// Create and run the application
	runner := app.NewRunner(version)

	if err := runner.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
