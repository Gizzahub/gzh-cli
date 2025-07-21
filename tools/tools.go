//go:build tools

// Package tools manages build tool dependencies for the project.
package tools

// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

//go:generate go install github.com/golangci/golangci-lint/cmd/golangci-lint
//go:generate go install mvdan.cc/gofumpt
//go:generate go install github.com/daixiang0/gci
//go:generate go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt
//go:generate go install golang.org/x/tools/cmd/goimports
//go:generate go install github.com/go-critic/go-critic/cmd/gocritic

// nolint
import (
	// gci
	_ "github.com/daixiang0/gci"
	// gocritic
	_ "github.com/go-critic/go-critic/cmd/gocritic"
	// golangci-lint
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	// gotestfmt
	_ "github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt"
	// goimports
	_ "golang.org/x/tools/cmd/goimports"
	// gofumpt
	_ "mvdan.cc/gofumpt"
)
