// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// OutputFormatter handles different output formats for git-synclone operations.
// It integrates with the existing synclone output formatting to maintain consistency.
type OutputFormatter struct {
	format string
	writer io.Writer
}

// NewOutputFormatter creates a new output formatter.
func NewOutputFormatter(format string, writer io.Writer) *OutputFormatter {
	return &OutputFormatter{
		format: format,
		writer: writer,
	}
}

// PrintCloneResult prints the result of a clone operation.
func (of *OutputFormatter) PrintCloneResult(result *CloneResult) error {
	switch of.format {
	case "json":
		return of.printCloneResultJSON(result)
	case "yaml":
		return of.printCloneResultYAML(result)
	case "table":
		return of.printCloneResultTable(result)
	default:
		return fmt.Errorf("unsupported format: %s", of.format)
	}
}

// PrintListResult prints the result of a list repositories operation.
func (of *OutputFormatter) PrintListResult(result *ListResult) error {
	switch of.format {
	case "json":
		return of.printListResultJSON(result)
	case "yaml":
		return of.printListResultYAML(result)
	case "table":
		return of.printListResultTable(result)
	default:
		return fmt.Errorf("unsupported format: %s", of.format)
	}
}

// PrintProgress prints progress information during operation.
func (of *OutputFormatter) PrintProgress(session *CloneSession) error {
	completed, failed, pending, percent := session.GetProgress()

	switch of.format {
	case "json":
		return of.printProgressJSON(session, completed, failed, pending, percent)
	case "yaml":
		return of.printProgressYAML(session, completed, failed, pending, percent)
	case "table", "text":
		return of.printProgressTable(session, completed, failed, pending, percent)
	default:
		return fmt.Errorf("unsupported format: %s", of.format)
	}
}

// PrintSessionList prints a list of sessions.
func (of *OutputFormatter) PrintSessionList(sessions []SessionInfo) error {
	switch of.format {
	case "json":
		return of.printSessionListJSON(sessions)
	case "yaml":
		return of.printSessionListYAML(sessions)
	case "table":
		return of.printSessionListTable(sessions)
	default:
		return fmt.Errorf("unsupported format: %s", of.format)
	}
}

// PrintSessionDetails prints detailed information about a session.
func (of *OutputFormatter) PrintSessionDetails(session *SessionInfo) error {
	switch of.format {
	case "json":
		return of.printSessionDetailsJSON(session)
	case "yaml":
		return of.printSessionDetailsYAML(session)
	case "text":
		return of.printSessionDetailsText(session)
	default:
		return fmt.Errorf("unsupported format: %s", of.format)
	}
}

// JSON formatters

func (of *OutputFormatter) printCloneResultJSON(result *CloneResult) error {
	encoder := json.NewEncoder(of.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func (of *OutputFormatter) printListResultJSON(result *ListResult) error {
	encoder := json.NewEncoder(of.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func (of *OutputFormatter) printProgressJSON(session *CloneSession, completed, failed, pending int, percent float64) error {
	progress := map[string]interface{}{
		"sessionId":    session.SessionID,
		"status":       session.State.Status,
		"completed":    completed,
		"failed":       failed,
		"pending":      pending,
		"percent":      percent,
		"startTime":    session.StartTime,
		"lastActivity": session.LastActivity,
	}

	encoder := json.NewEncoder(of.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(progress)
}

func (of *OutputFormatter) printSessionListJSON(sessions []SessionInfo) error {
	encoder := json.NewEncoder(of.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(sessions)
}

func (of *OutputFormatter) printSessionDetailsJSON(session *SessionInfo) error {
	encoder := json.NewEncoder(of.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(session)
}

// YAML formatters

func (of *OutputFormatter) printCloneResultYAML(result *CloneResult) error {
	encoder := yaml.NewEncoder(of.writer)
	defer encoder.Close()
	return encoder.Encode(result)
}

func (of *OutputFormatter) printListResultYAML(result *ListResult) error {
	encoder := yaml.NewEncoder(of.writer)
	defer encoder.Close()
	return encoder.Encode(result)
}

func (of *OutputFormatter) printProgressYAML(session *CloneSession, completed, failed, pending int, percent float64) error {
	progress := map[string]interface{}{
		"sessionId":    session.SessionID,
		"status":       session.State.Status,
		"completed":    completed,
		"failed":       failed,
		"pending":      pending,
		"percent":      percent,
		"startTime":    session.StartTime,
		"lastActivity": session.LastActivity,
	}

	encoder := yaml.NewEncoder(of.writer)
	defer encoder.Close()
	return encoder.Encode(progress)
}

func (of *OutputFormatter) printSessionListYAML(sessions []SessionInfo) error {
	encoder := yaml.NewEncoder(of.writer)
	defer encoder.Close()
	return encoder.Encode(sessions)
}

func (of *OutputFormatter) printSessionDetailsYAML(session *SessionInfo) error {
	encoder := yaml.NewEncoder(of.writer)
	defer encoder.Close()
	return encoder.Encode(session)
}

// Table formatters

func (of *OutputFormatter) printCloneResultTable(result *CloneResult) error {
	fmt.Fprintf(of.writer, "\n📊 Clone Operation Summary\n")
	fmt.Fprintf(of.writer, "==========================\n\n")

	fmt.Fprintf(of.writer, "Total Repositories: %d\n", result.TotalRepositories)
	fmt.Fprintf(of.writer, "✅ Successful:      %d\n", result.ClonesSuccessful)
	fmt.Fprintf(of.writer, "❌ Failed:          %d\n", result.ClonesFailed)
	fmt.Fprintf(of.writer, "⏭️  Skipped:         %d\n", result.ClonesSkipped)

	if result.ClonesFailed > 0 {
		fmt.Fprintf(of.writer, "\n❌ Failed Repositories:\n")
		fmt.Fprintf(of.writer, "------------------------\n")
		for _, repo := range result.Repositories {
			if !repo.Success && !repo.Skipped {
				fmt.Fprintf(of.writer, "  • %s: %s\n", repo.Name, repo.Error)
			}
		}
	}

	if result.ClonesSkipped > 0 {
		fmt.Fprintf(of.writer, "\n⏭️ Skipped Repositories:\n")
		fmt.Fprintf(of.writer, "------------------------\n")
		for _, repo := range result.Repositories {
			if repo.Skipped {
				fmt.Fprintf(of.writer, "  • %s: %s\n", repo.Name, repo.SkipReason)
			}
		}
	}

	fmt.Fprintf(of.writer, "\n")
	return nil
}

func (of *OutputFormatter) printListResultTable(result *ListResult) error {
	if result.TotalRepositories == 0 {
		fmt.Fprintf(of.writer, "No repositories found.\n")
		return nil
	}

	fmt.Fprintf(of.writer, "\n📋 Repository List (%d repositories)\n", result.TotalRepositories)
	fmt.Fprintf(of.writer, "=====================================\n\n")

	// Sort repositories by name for consistent output
	repos := make([]RepositoryInfo, len(result.Repositories))
	copy(repos, result.Repositories)
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	// Print table header
	fmt.Fprintf(of.writer, "%-30s %-15s %-10s %-8s %s\n", "NAME", "LANGUAGE", "VISIBILITY", "STARS", "DESCRIPTION")
	fmt.Fprintf(of.writer, "%s\n", strings.Repeat("-", 100))

	for _, repo := range repos {
		visibility := "public"
		if repo.Private {
			visibility = "private"
		}

		description := repo.Description
		if len(description) > 40 {
			description = description[:37] + "..."
		}

		language := repo.Language
		if language == "" {
			language = "Unknown"
		}

		fmt.Fprintf(of.writer, "%-30s %-15s %-10s %-8d %s\n",
			truncateString(repo.Name, 30),
			truncateString(language, 15),
			visibility,
			repo.Stars,
			description,
		)
	}

	fmt.Fprintf(of.writer, "\n")
	return nil
}

func (of *OutputFormatter) printProgressTable(session *CloneSession, completed, failed, pending int, percent float64) error {
	fmt.Fprintf(of.writer, "\n🔄 Clone Progress\n")
	fmt.Fprintf(of.writer, "================\n\n")

	fmt.Fprintf(of.writer, "Session ID:   %s\n", session.SessionID)
	fmt.Fprintf(of.writer, "Status:       %s\n", session.State.Status)
	fmt.Fprintf(of.writer, "Progress:     %.1f%% (%d/%d)\n", percent, completed+failed, session.State.TotalRepositories)
	fmt.Fprintf(of.writer, "✅ Completed: %d\n", completed)
	fmt.Fprintf(of.writer, "❌ Failed:    %d\n", failed)
	fmt.Fprintf(of.writer, "⏳ Pending:   %d\n", pending)
	fmt.Fprintf(of.writer, "Started:      %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(of.writer, "Last Update:  %s\n", session.LastActivity.Format("2006-01-02 15:04:05"))

	// Draw progress bar
	of.drawProgressBar(percent, 50)

	fmt.Fprintf(of.writer, "\n")
	return nil
}

func (of *OutputFormatter) printSessionListTable(sessions []SessionInfo) error {
	if len(sessions) == 0 {
		fmt.Fprintf(of.writer, "No sessions found.\n")
		return nil
	}

	fmt.Fprintf(of.writer, "\n📋 Sessions (%d sessions)\n", len(sessions))
	fmt.Fprintf(of.writer, "==========================\n\n")

	// Sort sessions by last updated (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastUpdated.After(sessions[j].LastUpdated)
	})

	fmt.Fprintf(of.writer, "%-20s %-12s %-15s %-12s %-8s %s\n",
		"SESSION ID", "PROVIDER", "ORGANIZATION", "STATUS", "PROGRESS", "LAST UPDATED")
	fmt.Fprintf(of.writer, "%s\n", strings.Repeat("-", 90))

	for _, session := range sessions {
		progressStr := fmt.Sprintf("%.1f%%", session.ProgressPercent)
		lastUpdated := session.LastUpdated.Format("2006-01-02 15:04")

		fmt.Fprintf(of.writer, "%-20s %-12s %-15s %-12s %-8s %s\n",
			truncateString(session.SessionID, 20),
			session.Provider,
			truncateString(session.Organization, 15),
			session.Status,
			progressStr,
			lastUpdated,
		)
	}

	fmt.Fprintf(of.writer, "\n")
	return nil
}

func (of *OutputFormatter) printSessionDetailsText(session *SessionInfo) error {
	fmt.Fprintf(of.writer, "\n📊 Session Details\n")
	fmt.Fprintf(of.writer, "==================\n\n")

	fmt.Fprintf(of.writer, "Session ID:       %s\n", session.SessionID)
	fmt.Fprintf(of.writer, "Provider:         %s\n", session.Provider)
	fmt.Fprintf(of.writer, "Organization:     %s\n", session.Organization)
	fmt.Fprintf(of.writer, "Target Path:      %s\n", session.TargetPath)
	fmt.Fprintf(of.writer, "Strategy:         %s\n", session.Strategy)
	fmt.Fprintf(of.writer, "Status:           %s\n", session.Status)
	fmt.Fprintf(of.writer, "\nProgress:\n")
	fmt.Fprintf(of.writer, "  Total:          %d repositories\n", session.TotalRepositories)
	fmt.Fprintf(of.writer, "  ✅ Completed:   %d\n", session.CompletedRepos)
	fmt.Fprintf(of.writer, "  ❌ Failed:      %d\n", session.FailedRepos)
	fmt.Fprintf(of.writer, "  ⏳ Pending:     %d\n", session.PendingRepos)
	fmt.Fprintf(of.writer, "  Progress:       %.1f%%\n", session.ProgressPercent)

	fmt.Fprintf(of.writer, "\nTimestamps:\n")
	fmt.Fprintf(of.writer, "  Started:        %s\n", session.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(of.writer, "  Last Updated:   %s\n", session.LastUpdated.Format("2006-01-02 15:04:05"))

	duration := session.LastUpdated.Sub(session.StartTime)
	fmt.Fprintf(of.writer, "  Duration:       %s\n", duration.Truncate(time.Second))

	// Draw progress bar
	of.drawProgressBar(session.ProgressPercent, 50)

	fmt.Fprintf(of.writer, "\n")
	return nil
}

// Helper functions

func (of *OutputFormatter) drawProgressBar(percent float64, width int) {
	fmt.Fprintf(of.writer, "\nProgress: [")

	filled := int(percent * float64(width) / 100)
	for i := 0; i < width; i++ {
		if i < filled {
			fmt.Fprintf(of.writer, "█")
		} else {
			fmt.Fprintf(of.writer, "░")
		}
	}

	fmt.Fprintf(of.writer, "] %.1f%%\n", percent)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// PrintError prints an error message in the appropriate format.
func (of *OutputFormatter) PrintError(err error) error {
	switch of.format {
	case "json":
		errorObj := map[string]interface{}{
			"error": err.Error(),
			"time":  time.Now(),
		}
		encoder := json.NewEncoder(of.writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(errorObj)
	case "yaml":
		errorObj := map[string]interface{}{
			"error": err.Error(),
			"time":  time.Now(),
		}
		encoder := yaml.NewEncoder(of.writer)
		defer encoder.Close()
		return encoder.Encode(errorObj)
	default:
		fmt.Fprintf(of.writer, "❌ Error: %s\n", err.Error())
		return nil
	}
}

// PrintSuccess prints a success message in the appropriate format.
func (of *OutputFormatter) PrintSuccess(message string) error {
	switch of.format {
	case "json":
		successObj := map[string]interface{}{
			"message": message,
			"time":    time.Now(),
			"status":  "success",
		}
		encoder := json.NewEncoder(of.writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(successObj)
	case "yaml":
		successObj := map[string]interface{}{
			"message": message,
			"time":    time.Now(),
			"status":  "success",
		}
		encoder := yaml.NewEncoder(of.writer)
		defer encoder.Close()
		return encoder.Encode(successObj)
	default:
		fmt.Fprintf(of.writer, "✅ %s\n", message)
		return nil
	}
}
