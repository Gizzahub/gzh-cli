// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
)

// Example of how the refactored functions would look using ToolChecker

// Before: 25+ lines of duplicated code
// func checkGolangciLint(ctx context.Context) DevEnvResult {
//     result := DevEnvResult{Tool: "golangci-lint", Status: "missing"}
//     lintPath, err := exec.LookPath("golangci-lint")
//     if err != nil {
//         result.Suggestion = "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
//         return result
//     }
//     result.Path = lintPath
//     result.Status = "found"
//     cmd := exec.CommandContext(ctx, "golangci-lint", "version")
//     output, err := cmd.Output()
//     if err != nil {
//         result.Status = "error"
//         result.Details = map[string]interface{}{"error": err.Error()}
//         return result
//     }
//     result.Version = strings.TrimSpace(string(output))
//     result.Status = "ok"
//     return result
// }

// After: 3 lines using ToolChecker
func checkGolangciLintRefactored(ctx context.Context) DevEnvResult {
	checker := CreateToolChecker("golangci-lint")
	return checker.Check(ctx)
}

func checkGofumptRefactored(ctx context.Context) DevEnvResult {
	checker := CreateToolChecker("gofumpt")
	return checker.Check(ctx)
}

func checkGciRefactored(ctx context.Context) DevEnvResult {
	checker := CreateToolChecker("gci")
	return checker.Check(ctx)
}

func checkDeadcodeRefactored(ctx context.Context) DevEnvResult {
	checker := CreateToolChecker("deadcode")
	return checker.Check(ctx)
}

func checkDuplRefactored(ctx context.Context) DevEnvResult {
	checker := CreateToolChecker("dupl")
	return checker.Check(ctx)
}

// Even better: Check multiple tools at once
func checkAllGoToolsRefactored(ctx context.Context) []DevEnvResult {
	return CheckMultipleTools(ctx, []string{
		"golangci-lint", "gofumpt", "gci", "deadcode", "dupl",
	})
}

// Or check all common tools
func checkAllCommonToolsRefactored(ctx context.Context) []DevEnvResult {
	return CheckAllCommonTools(ctx)
}