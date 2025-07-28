// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package clone

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ProgressReporter handles progress reporting for clone operations.
type ProgressReporter struct {
	format    OutputFormat
	quiet     bool
	verbose   bool
	total     int
	completed int
	failed    int
	skipped   int
	startTime time.Time
	session   *Session
}

// NewProgressReporter creates a new progress reporter.
func NewProgressReporter(format string, quiet, verbose bool) *ProgressReporter {
	return &ProgressReporter{
		format:  OutputFormat(format),
		quiet:   quiet,
		verbose: verbose,
	}
}

// Start initializes the progress tracking.
func (p *ProgressReporter) Start(total int) {
	p.total = total
	p.completed = 0
	p.failed = 0
	p.skipped = 0
	p.startTime = time.Now()

	if !p.quiet {
		switch p.format {
		case FormatProgress:
			p.Info("Starting clone operation for %d repositories...", total)
		case FormatJSON:
			p.printJSONEvent("start", map[string]interface{}{
				"total":      total,
				"started_at": p.startTime.Format(time.RFC3339),
			})
		}
	}
}

// Success reports a successful clone operation.
func (p *ProgressReporter) Success(repoName string) {
	p.completed++

	if !p.quiet {
		switch p.format {
		case FormatProgress:
			p.printProgress(repoName, "✅", "completed")
		case FormatJSON:
			p.printJSONEvent("success", map[string]interface{}{
				"repository": repoName,
				"completed":  p.completed,
				"total":      p.total,
			})
		case FormatTable:
			fmt.Printf("%-50s %s\n", repoName, "SUCCESS")
		}
	}
}

// Fail reports a failed clone operation.
func (p *ProgressReporter) Fail(repoName string, err error) {
	p.failed++

	if !p.quiet {
		switch p.format {
		case FormatProgress:
			if p.verbose {
				p.printProgress(repoName, "❌", fmt.Sprintf("failed: %v", err))
			} else {
				p.printProgress(repoName, "❌", "failed")
			}
		case FormatJSON:
			p.printJSONEvent("fail", map[string]interface{}{
				"repository": repoName,
				"error":      err.Error(),
				"failed":     p.failed,
				"total":      p.total,
			})
		case FormatTable:
			if p.verbose {
				fmt.Printf("%-50s %s: %v\n", repoName, "FAILED", err)
			} else {
				fmt.Printf("%-50s %s\n", repoName, "FAILED")
			}
		}
	}
}

// Skip reports a skipped repository.
func (p *ProgressReporter) Skip(repoName, reason string) {
	p.skipped++

	if !p.quiet && p.verbose {
		switch p.format {
		case FormatProgress:
			p.printProgress(repoName, "⏭️", fmt.Sprintf("skipped: %s", reason))
		case FormatJSON:
			p.printJSONEvent("skip", map[string]interface{}{
				"repository": repoName,
				"reason":     reason,
				"skipped":    p.skipped,
				"total":      p.total,
			})
		case FormatTable:
			fmt.Printf("%-50s %s: %s\n", repoName, "SKIPPED", reason)
		}
	}
}

// Retry reports a retry attempt.
func (p *ProgressReporter) Retry(repoName string, attempt int, err error) {
	if !p.quiet && p.verbose {
		switch p.format {
		case FormatProgress:
			p.printProgress(repoName, "🔄", fmt.Sprintf("retry %d: %v", attempt, err))
		case FormatJSON:
			p.printJSONEvent("retry", map[string]interface{}{
				"repository": repoName,
				"attempt":    attempt,
				"error":      err.Error(),
			})
		case FormatTable:
			fmt.Printf("%-50s %s %d: %v\n", repoName, "RETRY", attempt, err)
		}
	}
}

// Info prints an informational message.
func (p *ProgressReporter) Info(format string, args ...interface{}) {
	if !p.quiet {
		switch p.format {
		case FormatProgress, FormatTable:
			fmt.Printf(format+"\n", args...)
		case FormatJSON:
			p.printJSONEvent("info", map[string]interface{}{
				"message": fmt.Sprintf(format, args...),
			})
		}
	}
}

// Warning prints a warning message.
func (p *ProgressReporter) Warning(format string, args ...interface{}) {
	if !p.quiet {
		switch p.format {
		case FormatProgress, FormatTable:
			fmt.Printf("⚠️  "+format+"\n", args...)
		case FormatJSON:
			p.printJSONEvent("warning", map[string]interface{}{
				"message": fmt.Sprintf(format, args...),
			})
		}
	}
}

// Error prints an error message.
func (p *ProgressReporter) Error(format string, args ...interface{}) {
	switch p.format {
	case FormatProgress, FormatTable:
		fmt.Fprintf(os.Stderr, "❌ "+format+"\n", args...)
	case FormatJSON:
		p.printJSONEvent("error", map[string]interface{}{
			"message": fmt.Sprintf(format, args...),
		})
	default:
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// Finish completes the progress reporting.
func (p *ProgressReporter) Finish() {
	if !p.quiet {
		duration := time.Since(p.startTime)

		switch p.format {
		case FormatProgress:
			p.Info("\nOperation completed in %v", duration)
			p.Info("Total: %d, Completed: %d, Failed: %d, Skipped: %d",
				p.total, p.completed, p.failed, p.skipped)
		case FormatJSON:
			p.printJSONEvent("finish", map[string]interface{}{
				"total":     p.total,
				"completed": p.completed,
				"failed":    p.failed,
				"skipped":   p.skipped,
				"duration":  duration.String(),
				"ended_at":  time.Now().Format(time.RFC3339),
			})
		case FormatTable:
			fmt.Println("\n" + p.getSummaryTable())
		}
	}
}

// ResumeSession configures the reporter for a resumed session.
func (p *ProgressReporter) ResumeSession(session *Session) {
	p.session = session
	if !p.quiet {
		switch p.format {
		case FormatProgress:
			progress := session.GetProgress()
			p.Info("Resuming session %s (%.1f%% complete)", session.ID, progress)
		case FormatJSON:
			p.printJSONEvent("resume", map[string]interface{}{
				"session_id": session.ID,
				"progress":   session.GetProgress(),
				"started_at": session.StartedAt.Format(time.RFC3339),
			})
		}
	}
}

// printProgress prints a progress line for a repository.
func (p *ProgressReporter) printProgress(repoName, icon, status string) {
	progress := float64(p.completed+p.failed+p.skipped) / float64(p.total) * 100
	fmt.Printf("[%5.1f%%] %s %-50s %s\n", progress, icon, repoName, status)
}

// printJSONEvent prints a JSON event.
func (p *ProgressReporter) printJSONEvent(eventType string, data map[string]interface{}) {
	event := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"type":      eventType,
		"data":      data,
	}

	if jsonData, err := json.Marshal(event); err == nil {
		fmt.Println(string(jsonData))
	}
}

// getSummaryTable returns a formatted summary table.
func (p *ProgressReporter) getSummaryTable() string {
	duration := time.Since(p.startTime)

	return fmt.Sprintf(`Summary:
┌─────────────┬───────┐
│ Status      │ Count │
├─────────────┼───────┤
│ Total       │ %5d │
│ Completed   │ %5d │
│ Failed      │ %5d │
│ Skipped     │ %5d │
├─────────────┼───────┤
│ Duration    │ %5s │
└─────────────┴───────┘`,
		p.total, p.completed, p.failed, p.skipped,
		formatDuration(duration))
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}

// ProgressStats contains current progress statistics.
type ProgressStats struct {
	Total     int           `json:"total"`
	Completed int           `json:"completed"`
	Failed    int           `json:"failed"`
	Skipped   int           `json:"skipped"`
	Duration  time.Duration `json:"duration"`
	Progress  float64       `json:"progress"`
}

// GetStats returns current progress statistics.
func (p *ProgressReporter) GetStats() ProgressStats {
	var progress float64
	if p.total > 0 {
		progress = float64(p.completed+p.failed+p.skipped) / float64(p.total) * 100
	}

	return ProgressStats{
		Total:     p.total,
		Completed: p.completed,
		Failed:    p.failed,
		Skipped:   p.skipped,
		Duration:  time.Since(p.startTime),
		Progress:  progress,
	}
}

// SetVerbose sets the verbose flag.
func (p *ProgressReporter) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// SetQuiet sets the quiet flag.
func (p *ProgressReporter) SetQuiet(quiet bool) {
	p.quiet = quiet
}

// IsQuiet returns whether quiet mode is enabled.
func (p *ProgressReporter) IsQuiet() bool {
	return p.quiet
}

// IsVerbose returns whether verbose mode is enabled.
func (p *ProgressReporter) IsVerbose() bool {
	return p.verbose
}
