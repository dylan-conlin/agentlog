package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// tailCmd represents the tail command
var tailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Watch .agentlog/errors.jsonl for new errors in real-time",
	Long: `Watch the .agentlog/errors.jsonl file for new errors as they appear.

Outputs errors in real-time as they are logged. Use Ctrl+C to stop watching.

Examples:
  agentlog tail          # Watch errors in human-readable format
  agentlog tail --json   # Watch errors in JSON format (one object per line)`,
	RunE: runTail,
}

func init() {
	rootCmd.AddCommand(tailCmd)
}

func runTail(cmd *cobra.Command, args []string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Run tail
	err = tailFile(ctx, cwd, cmd.OutOrStdout(), IsJSONOutput())
	if err != nil && err != context.Canceled {
		if os.IsNotExist(err) {
			fmt.Fprintln(cmd.OutOrStdout(), "No errors file found. Run 'agentlog init' to set up.")
			return nil
		}
		return err
	}

	return nil
}

// formatTailEntry formats a single error entry for tail output
func formatTailEntry(entry ErrorEntry, jsonMode bool) string {
	if jsonMode {
		output, err := json.Marshal(entry)
		if err != nil {
			return "{}"
		}
		return string(output)
	}

	// Human-readable format
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s\n", entry.Timestamp, entry.Message))
	sb.WriteString(fmt.Sprintf("  Source: %s | Type: %s\n", entry.Source, entry.ErrorType))
	return sb.String()
}

// tailFile watches the errors file and outputs new entries
func tailFile(ctx context.Context, baseDir string, w io.Writer, jsonMode bool) error {
	filePath := filepath.Join(baseDir, ".agentlog", "errors.jsonl")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}

	// Open file and seek to end initially (to show existing entries first)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read and output all existing entries first
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry ErrorEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip malformed lines
		}

		fmt.Fprintln(w, formatTailEntry(entry, jsonMode))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Get current position (after reading existing entries)
	offset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return fmt.Errorf("error getting file position: %w", err)
	}

	// Poll for new entries
	pollInterval := 500 * time.Millisecond
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check for new content
			newOffset, err := readNewEntries(filePath, offset, w, jsonMode)
			if err != nil {
				// File might have been truncated or rotated
				if os.IsNotExist(err) {
					return err
				}
				// Try to recover by re-checking file
				continue
			}
			offset = newOffset
		}
	}
}

// readNewEntries reads any new entries after the given offset
func readNewEntries(filePath string, offset int64, w io.Writer, jsonMode bool) (int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return offset, err
	}
	defer file.Close()

	// Seek to our last position
	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return offset, err
	}

	// Read any new lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry ErrorEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip malformed lines
		}

		fmt.Fprintln(w, formatTailEntry(entry, jsonMode))
	}

	// Get new offset
	newOffset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return offset, err
	}

	return newOffset, scanner.Err()
}
