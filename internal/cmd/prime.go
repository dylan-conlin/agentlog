package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// PrimeSummary is the output structure for prime command
type PrimeSummary struct {
	TotalErrors    int              `json:"total_errors"`
	Last24hErrors  int              `json:"last_24h_errors"`
	LastHourErrors int              `json:"last_hour_errors"`
	TopErrorTypes  []ErrorTypeCount `json:"top_error_types"`
	TopSources     []SourceCount    `json:"top_sources"`
	ActionableTip  string           `json:"actionable_tip"`
	GeneratedAt    string           `json:"generated_at"`
	NoLogFile      bool             `json:"no_log_file,omitempty"`
}

// ErrorTypeCount aggregates error counts by type
type ErrorTypeCount struct {
	ErrorType string `json:"error_type"`
	Count     int    `json:"count"`
}

// SourceCount aggregates error counts by source
type SourceCount struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// primeCmd represents the prime command
var primeCmd = &cobra.Command{
	Use:   "prime",
	Short: "Output context summary for AI agent injection",
	Long: `Output a concise summary of recent errors for AI agent context injection.

This command is designed to be used by orchestration hooks to inject
error context into agent prompts. Output includes:
  - Recent error count (last hour, last 24h)
  - Top error types by frequency
  - Top sources by frequency
  - Actionable tip for the agent

Examples:
  agentlog prime          # Human-readable summary
  agentlog prime --json   # JSON for programmatic use`,
	Run: runPrimeCommand,
}

func init() {
	rootCmd.AddCommand(primeCmd)
}

func runPrimeCommand(cmd *cobra.Command, args []string) {
	summary, err := generatePrimeSummary()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error generating summary: %v\n", err)
		return
	}

	var output string
	if IsJSONOutput() {
		output = formatPrimeSummaryJSON(summary)
	} else {
		output = formatPrimeSummaryHuman(summary)
	}

	fmt.Fprint(cmd.OutOrStdout(), output)
}

// generatePrimeSummary reads errors and generates aggregate summary
func generatePrimeSummary() (PrimeSummary, error) {
	summary := PrimeSummary{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return summary, err
	}

	// Read errors using existing function
	entries, err := readErrors(cwd)
	if err != nil {
		if os.IsNotExist(err) {
			summary.NoLogFile = true
			return summary, nil
		}
		return summary, err
	}

	if len(entries) == 0 {
		return summary, nil
	}

	// Calculate time boundaries
	now := time.Now().UTC()
	oneHourAgo := now.Add(-1 * time.Hour)
	twentyFourHoursAgo := now.Add(-24 * time.Hour)

	// Aggregate counts
	errorTypeCounts := make(map[string]int)
	sourceCounts := make(map[string]int)
	var lastHour, last24h int

	for _, entry := range entries {
		// Parse timestamp
		ts, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
		if err != nil {
			// Try without nano
			ts, err = time.Parse(time.RFC3339, entry.Timestamp)
		}
		if err != nil {
			continue
		}

		// Count by time window
		if ts.After(oneHourAgo) {
			lastHour++
		}
		if ts.After(twentyFourHoursAgo) {
			last24h++
		}

		// Aggregate by type and source
		errorTypeCounts[entry.ErrorType]++
		sourceCounts[entry.Source]++
	}

	summary.TotalErrors = len(entries)
	summary.LastHourErrors = lastHour
	summary.Last24hErrors = last24h
	summary.TopErrorTypes = topN(errorTypeCounts, 3)
	summary.TopSources = topNSources(sourceCounts, 3)
	summary.ActionableTip = generateTip(summary)

	return summary, nil
}

// topN returns top N error types sorted by count
func topN(counts map[string]int, n int) []ErrorTypeCount {
	var result []ErrorTypeCount
	for errType, count := range counts {
		result = append(result, ErrorTypeCount{ErrorType: errType, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	if len(result) > n {
		result = result[:n]
	}
	return result
}

// topNSources returns top N sources sorted by count
func topNSources(counts map[string]int, n int) []SourceCount {
	var result []SourceCount
	for source, count := range counts {
		result = append(result, SourceCount{Source: source, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})
	if len(result) > n {
		result = result[:n]
	}
	return result
}

// generateTip creates actionable advice based on error patterns
func generateTip(summary PrimeSummary) string {
	if summary.TotalErrors == 0 {
		return ""
	}

	if len(summary.TopErrorTypes) == 0 || len(summary.TopSources) == 0 {
		return ""
	}

	topType := summary.TopErrorTypes[0]
	topSource := summary.TopSources[0]
	percentage := (topType.Count * 100) / summary.TotalErrors

	return fmt.Sprintf("Focus on %s in %s - %d%% of errors", topType.ErrorType, topSource.Source, percentage)
}

// formatPrimeSummaryJSON returns JSON formatted output
func formatPrimeSummaryJSON(summary PrimeSummary) string {
	output, _ := json.MarshalIndent(summary, "", "  ")
	return string(output) + "\n"
}

// formatPrimeSummaryHuman returns human-readable output
func formatPrimeSummaryHuman(summary PrimeSummary) string {
	var sb strings.Builder

	if summary.NoLogFile {
		sb.WriteString("agentlog: No error log found (.agentlog/errors.jsonl)\n")
		sb.WriteString("  Run 'agentlog init' to set up error tracking\n")
		return sb.String()
	}

	if summary.TotalErrors == 0 {
		sb.WriteString("agentlog: No errors logged\n")
		return sb.String()
	}

	// Error counts
	errWord := "errors"
	if summary.TotalErrors == 1 {
		errWord = "error"
	}
	sb.WriteString(fmt.Sprintf("agentlog: %d %s", summary.TotalErrors, errWord))
	if summary.LastHourErrors > 0 {
		sb.WriteString(fmt.Sprintf(" (%d in last hour)", summary.LastHourErrors))
	}
	sb.WriteString("\n")

	// Top error types
	if len(summary.TopErrorTypes) > 0 {
		sb.WriteString("  Top types: ")
		var types []string
		for _, t := range summary.TopErrorTypes {
			types = append(types, fmt.Sprintf("%s (%d)", t.ErrorType, t.Count))
		}
		sb.WriteString(strings.Join(types, ", "))
		sb.WriteString("\n")
	}

	// Top sources
	if len(summary.TopSources) > 0 {
		sb.WriteString("  Sources: ")
		var sources []string
		for _, s := range summary.TopSources {
			sources = append(sources, fmt.Sprintf("%s (%d)", s.Source, s.Count))
		}
		sb.WriteString(strings.Join(sources, ", "))
		sb.WriteString("\n")
	}

	// Actionable tip
	if summary.ActionableTip != "" {
		sb.WriteString("  Tip: ")
		sb.WriteString(summary.ActionableTip)
		sb.WriteString("\n")
	}

	return sb.String()
}
