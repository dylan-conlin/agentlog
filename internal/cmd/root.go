package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	jsonOutput bool
	aiHelp     bool
)

// CommandMetadata provides machine-readable command information for AI agents
type CommandMetadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Commands    []CommandInfo     `json:"commands"`
	GlobalFlags map[string]string `json:"global_flags"`
}

// CommandInfo describes a single command
type CommandInfo struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Usage       string            `json:"usage"`
	Flags       map[string]string `json:"flags,omitempty"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "agentlog",
	Short: "AI-native development observability CLI",
	Long: `agentlog - Error visibility for AI agents in any development environment

agentlog provides structured error aggregation for AI agents working on
any development stack. It reads errors from .agentlog/errors.jsonl and
presents them in formats optimized for both humans and AI agents.

Key features:
  - Local-first: No cloud, no account required
  - AI-optimized: Structured output for agent consumption
  - Zero-config: Auto-detect stack, works out of the box
  - Universal: Any language via file convention

Quick start:
  agentlog init       Initialize agentlog in your project
  agentlog errors     View recent errors
  agentlog tail       Watch errors in real-time
  agentlog prime      Output context summary for AI agents`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Handle --ai-help before running any command
		if aiHelp {
			printAIHelp()
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior: show help
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format for programmatic use")
	rootCmd.PersistentFlags().BoolVar(&aiHelp, "ai-help", false, "Output machine-readable command metadata")
}

// IsJSONOutput returns whether JSON output is enabled
func IsJSONOutput() bool {
	return jsonOutput
}

// IsTTY returns whether stdout is a terminal
func IsTTY() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// printAIHelp outputs machine-readable metadata for AI agents
func printAIHelp() {
	metadata := CommandMetadata{
		Name:        "agentlog",
		Version:     "0.1.0",
		Description: "AI-native development observability CLI - error visibility for agents in any stack",
		GlobalFlags: map[string]string{
			"--json":    "Output in JSON format for programmatic use",
			"--ai-help": "Output this machine-readable command metadata",
		},
		Commands: []CommandInfo{
			{
				Name:        "init",
				Description: "Initialize agentlog in your project, detect stack, create config",
				Usage:       "agentlog init [flags]",
			},
			{
				Name:        "errors",
				Description: "Query and display errors from .agentlog/errors.jsonl",
				Usage:       "agentlog errors [flags]",
				Flags: map[string]string{
					"--limit":  "Maximum number of errors to show (default: 10)",
					"--source": "Filter by source (frontend, backend, cli, worker, test)",
					"--type":   "Filter by error type",
					"--since":  "Show errors since time (e.g., '1h', '30m', '2024-01-01')",
				},
			},
			{
				Name:        "tail",
				Description: "Watch .agentlog/errors.jsonl for new errors in real-time",
				Usage:       "agentlog tail [flags]",
			},
			{
				Name:        "doctor",
				Description: "Check agentlog configuration and health",
				Usage:       "agentlog doctor",
			},
			{
				Name:        "prime",
				Description: "Output context summary for AI agent injection",
				Usage:       "agentlog prime",
			},
		},
	}

	output, _ := json.MarshalIndent(metadata, "", "  ")
	fmt.Println(string(output))
}
