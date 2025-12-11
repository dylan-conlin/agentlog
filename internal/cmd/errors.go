package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// ErrorEntry represents a single error from errors.jsonl
type ErrorEntry struct {
	Timestamp string                 `json:"timestamp"`
	Source    string                 `json:"source"`
	ErrorType string                 `json:"error_type"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

var (
	errorsLimit  int
	errorsSource string
	errorsType   string
	errorsSince  string
)

// errorsCmd represents the errors command
var errorsCmd = &cobra.Command{
	Use:   "errors",
	Short: "Query and display errors from .agentlog/errors.jsonl",
	Long: `Query and display errors from the local .agentlog/errors.jsonl file.

Supports filtering by source, type, and time. Output is human-readable by
default, or JSON with the --json flag.

Examples:
  agentlog errors                    # Show last 10 errors
  agentlog errors --limit 50         # Show last 50 errors
  agentlog errors --source frontend  # Show only frontend errors
  agentlog errors --type DATABASE_ERROR  # Show only database errors
  agentlog errors --since 1h         # Show errors from last hour
  agentlog errors --json             # Output as JSON array`,
	RunE: runErrors,
}

func init() {
	rootCmd.AddCommand(errorsCmd)

	errorsCmd.Flags().IntVar(&errorsLimit, "limit", 10, "Maximum number of errors to show")
	errorsCmd.Flags().StringVar(&errorsSource, "source", "", "Filter by source (frontend, backend, cli, worker, test)")
	errorsCmd.Flags().StringVar(&errorsType, "type", "", "Filter by error type")
	errorsCmd.Flags().StringVar(&errorsSince, "since", "", "Show errors since time (e.g., '1h', '30m', '2024-01-01')")
}

func runErrors(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Read errors
	entries, err := readErrors(cwd)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(cmd.OutOrStdout(), "No errors file found. Run 'agentlog init' to set up.")
			return nil
		}
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No errors recorded yet.")
		return nil
	}

	// Parse --since if provided
	var sinceTime time.Time
	if errorsSince != "" {
		sinceTime, err = parseSince(errorsSince)
		if err != nil {
			return fmt.Errorf("invalid --since value: %w", err)
		}
	}

	// Apply filters
	filtered := filterErrors(entries, errorsSource, errorsType, sinceTime)

	// Apply limit (from the end - most recent)
	if errorsLimit > 0 && len(filtered) > errorsLimit {
		filtered = filtered[len(filtered)-errorsLimit:]
	}

	// Output
	if IsJSONOutput() {
		fmt.Fprintln(cmd.OutOrStdout(), formatJSON(filtered))
	} else {
		fmt.Fprint(cmd.OutOrStdout(), formatHuman(filtered, len(entries)))
	}

	return nil
}

// readErrors reads all error entries from .agentlog/errors.jsonl
func readErrors(baseDir string) ([]ErrorEntry, error) {
	filePath := filepath.Join(baseDir, ".agentlog", "errors.jsonl")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []ErrorEntry
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry ErrorEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip malformed lines with warning to stderr
			fmt.Fprintf(os.Stderr, "Warning: skipping malformed line %d: %v\n", lineNum, err)
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return entries, fmt.Errorf("error reading file: %w", err)
	}

	return entries, nil
}

// parseSince parses a --since value into a time.Time
// Supports duration format (1h, 30m) and date format (2024-01-01)
func parseSince(since string) (time.Time, error) {
	if since == "" {
		return time.Time{}, fmt.Errorf("empty since value")
	}

	// Try duration first (relative)
	if dur, err := time.ParseDuration(since); err == nil {
		return time.Now().Add(-dur), nil
	}

	// Try date formats
	formats := []string{
		"2006-01-02",                // YYYY-MM-DD
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 UTC
	}

	for _, f := range formats {
		if t, err := time.Parse(f, since); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s (use '1h', '30m', or 'YYYY-MM-DD')", since)
}

// filterErrors applies source, type, and since filters
func filterErrors(entries []ErrorEntry, source, errType string, since time.Time) []ErrorEntry {
	if source == "" && errType == "" && since.IsZero() {
		return entries
	}

	var filtered []ErrorEntry
	for _, e := range entries {
		// Filter by source
		if source != "" && e.Source != source {
			continue
		}

		// Filter by type
		if errType != "" && e.ErrorType != errType {
			continue
		}

		// Filter by since
		if !since.IsZero() {
			entryTime, err := time.Parse(time.RFC3339, e.Timestamp)
			if err != nil {
				// Try with milliseconds
				entryTime, err = time.Parse("2006-01-02T15:04:05.000Z", e.Timestamp)
			}
			if err != nil {
				continue // Skip entries with unparseable timestamps
			}
			if entryTime.Before(since) {
				continue
			}
		}

		filtered = append(filtered, e)
	}

	return filtered
}

// formatHuman formats errors for human-readable output
func formatHuman(entries []ErrorEntry, totalCount int) string {
	if len(entries) == 0 {
		return "No errors match the filter criteria.\n"
	}

	var sb strings.Builder

	for i, e := range entries {
		if i > 0 {
			sb.WriteString("\n")
		}

		sb.WriteString(fmt.Sprintf("Error: %s\n", e.Message))
		sb.WriteString(fmt.Sprintf("  Source: %s | Type: %s\n", e.Source, e.ErrorType))
		sb.WriteString(fmt.Sprintf("  Time: %s\n", e.Timestamp))
	}

	if len(entries) < totalCount {
		sb.WriteString(fmt.Sprintf("\nShowing %d of %d errors (use --limit to see more)\n", len(entries), totalCount))
	}

	return sb.String()
}

// formatJSON formats errors as JSON array
func formatJSON(entries []ErrorEntry) string {
	if entries == nil {
		entries = []ErrorEntry{}
	}

	output, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return "[]"
	}

	return string(output)
}
