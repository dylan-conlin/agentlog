package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentlog/agentlog/internal/self"
	"github.com/spf13/cobra"
)

const (
	// MaxFileSize is the maximum recommended size for errors.jsonl (10MB)
	MaxFileSize = 10 * 1024 * 1024
	// WarnFileSize is the threshold for warning about file size (80% of max)
	WarnFileSize = 8 * 1024 * 1024
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "ok", "warning", "error"
	Message string `json:"message"`
}

// HealthResult is the overall health check result
type HealthResult struct {
	Status  string        `json:"status"` // "healthy", "unhealthy", "warning"
	Checks  []HealthCheck `json:"checks"`
	Summary string        `json:"summary"`
}

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check agentlog configuration and health",
	Long: `Check agentlog configuration and health.

Verifies:
  - .agentlog/ directory exists
  - errors.jsonl is valid JSONL format
  - File size is within limits
  - No obvious configuration issues

Examples:
  agentlog doctor         # Human-readable health check
  agentlog doctor --json  # JSON output for programmatic use`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		self.LogError(".", "GETWD_ERROR", err.Error())
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	result := checkHealth(cwd)

	if IsJSONOutput() {
		fmt.Fprint(cmd.OutOrStdout(), formatHealthJSON(result))
	} else {
		fmt.Fprint(cmd.OutOrStdout(), formatHealthHuman(result))
	}

	return nil
}

// checkHealth performs all health checks and returns the result
func checkHealth(baseDir string) HealthResult {
	result := HealthResult{
		Status: "healthy",
		Checks: []HealthCheck{},
	}

	agentlogDir := filepath.Join(baseDir, ".agentlog")
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")

	// Check 1: .agentlog directory exists
	dirCheck := checkDirectory(agentlogDir)
	result.Checks = append(result.Checks, dirCheck)

	if dirCheck.Status == "error" {
		result.Status = "unhealthy"
		result.Summary = "agentlog is not initialized. Run 'agentlog init' to set up."
		return result
	}

	// Check 2: errors.jsonl exists and is accessible
	fileCheck := checkFile(errorsFile)
	result.Checks = append(result.Checks, fileCheck)

	if fileCheck.Status == "error" {
		// File doesn't exist yet - this is OK for a fresh setup
		if fileCheck.Message == "File does not exist" {
			fileCheck.Status = "ok"
			fileCheck.Message = "errors.jsonl not yet created (will be created on first error)"
		}
	}

	// Check 3: JSONL validity (only if file exists)
	if fileExists(errorsFile) {
		jsonlCheck := checkJSONL(errorsFile)
		result.Checks = append(result.Checks, jsonlCheck)

		if jsonlCheck.Status == "error" {
			result.Status = "unhealthy"
		} else if jsonlCheck.Status == "warning" && result.Status == "healthy" {
			result.Status = "warning"
		}
	}

	// Check file size
	if fileExists(errorsFile) {
		sizeCheck := checkFileSize(errorsFile)
		result.Checks = append(result.Checks, sizeCheck)

		if sizeCheck.Status == "warning" && result.Status == "healthy" {
			result.Status = "warning"
		}
	}

	// Generate summary
	result.Summary = generateSummary(result)

	return result
}

// checkDirectory verifies the .agentlog directory exists
func checkDirectory(dirPath string) HealthCheck {
	check := HealthCheck{
		Name: "Directory",
	}

	info, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		check.Status = "error"
		check.Message = ".agentlog directory NOT FOUND. Run 'agentlog init' to create it."
		return check
	}
	if err != nil {
		check.Status = "error"
		check.Message = fmt.Sprintf("Cannot access directory: %v", err)
		return check
	}
	if !info.IsDir() {
		check.Status = "error"
		check.Message = ".agentlog exists but is not a directory"
		return check
	}

	check.Status = "ok"
	check.Message = ".agentlog directory exists"
	return check
}

// checkFile verifies errors.jsonl exists and is readable
func checkFile(filePath string) HealthCheck {
	check := HealthCheck{
		Name: "Errors file",
	}

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		check.Status = "error"
		check.Message = "File does not exist"
		return check
	}
	if err != nil {
		check.Status = "error"
		check.Message = fmt.Sprintf("Cannot access file: %v", err)
		return check
	}
	if info.IsDir() {
		check.Status = "error"
		check.Message = "errors.jsonl is a directory (expected file)"
		return check
	}

	check.Status = "ok"
	check.Message = "errors.jsonl exists and is readable"
	return check
}

// checkJSONL validates that the file contains valid JSONL
func checkJSONL(filePath string) HealthCheck {
	check := HealthCheck{
		Name: "JSONL format",
	}

	file, err := os.Open(filePath)
	if err != nil {
		check.Status = "error"
		check.Message = fmt.Sprintf("Cannot open file: %v", err)
		return check
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	validLines := 0
	malformedLines := 0
	var malformedLineNums []int

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var js json.RawMessage
		if err := json.Unmarshal([]byte(line), &js); err != nil {
			malformedLines++
			if len(malformedLineNums) < 5 { // Only track first 5 malformed lines
				malformedLineNums = append(malformedLineNums, lineNum)
			}
		} else {
			validLines++
		}
	}

	if err := scanner.Err(); err != nil {
		check.Status = "error"
		check.Message = fmt.Sprintf("Error reading file: %v", err)
		return check
	}

	if malformedLines > 0 {
		check.Status = "warning"
		lineNumStr := formatLineNumbers(malformedLineNums)
		check.Message = fmt.Sprintf("%d malformed/invalid JSON lines (lines: %s). %d valid entries.", malformedLines, lineNumStr, validLines)
		return check
	}

	check.Status = "ok"
	check.Message = fmt.Sprintf("All %d entries are valid JSON", validLines)
	return check
}

// checkFileSize checks if file size is within limits
func checkFileSize(filePath string) HealthCheck {
	check := HealthCheck{
		Name: "File size",
	}

	info, err := os.Stat(filePath)
	if err != nil {
		check.Status = "error"
		check.Message = fmt.Sprintf("Cannot stat file: %v", err)
		return check
	}

	size := info.Size()
	sizeMB := float64(size) / (1024 * 1024)

	if size > MaxFileSize {
		check.Status = "error"
		check.Message = fmt.Sprintf("File size (%.1fMB) exceeds 10MB limit. Rotation needed.", sizeMB)
		return check
	}

	if size > WarnFileSize {
		check.Status = "warning"
		check.Message = fmt.Sprintf("File is large (%.1fMB). Approaching 10MB limit. Consider rotation.", sizeMB)
		return check
	}

	check.Status = "ok"
	check.Message = fmt.Sprintf("File size OK (%.2fMB)", sizeMB)
	return check
}

// fileExists checks if a file exists
func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// formatLineNumbers formats line numbers for display
func formatLineNumbers(nums []int) string {
	if len(nums) == 0 {
		return ""
	}
	var strs []string
	for _, n := range nums {
		strs = append(strs, fmt.Sprintf("%d", n))
	}
	result := strings.Join(strs, ", ")
	if len(nums) == 5 {
		result += "..."
	}
	return result
}

// generateSummary creates a summary message based on checks
func generateSummary(result HealthResult) string {
	okCount := 0
	warnCount := 0
	errorCount := 0

	for _, check := range result.Checks {
		switch check.Status {
		case "ok":
			okCount++
		case "warning":
			warnCount++
		case "error":
			errorCount++
		}
	}

	if errorCount > 0 {
		return fmt.Sprintf("%d issues found. See details above.", errorCount)
	}
	if warnCount > 0 {
		return fmt.Sprintf("All checks passed with %d warning(s).", warnCount)
	}
	return "All checks passed. agentlog is healthy."
}

// formatHealthJSON formats health result as JSON
func formatHealthJSON(result HealthResult) string {
	output, _ := json.MarshalIndent(result, "", "  ")
	return string(output) + "\n"
}

// formatHealthHuman formats health result for human reading
func formatHealthHuman(result HealthResult) string {
	var sb strings.Builder

	sb.WriteString("agentlog doctor\n")
	sb.WriteString("===============\n\n")

	for _, check := range result.Checks {
		icon := getStatusIcon(check.Status)
		sb.WriteString(fmt.Sprintf("%s %s: %s\n", icon, check.Name, check.Message))
	}

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Status: %s\n", strings.ToUpper(result.Status)))
	sb.WriteString(result.Summary + "\n")

	return sb.String()
}

// getStatusIcon returns an icon for the status
func getStatusIcon(status string) string {
	switch status {
	case "ok":
		return "[OK]"
	case "warning":
		return "[WARNING]"
	case "error":
		return "[ERROR]"
	default:
		return "[?]"
	}
}
