package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bonyuta0204/pr-analyzer/internal/analyzer"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "pr-analyzer <owner/repo>",
		Short: "Analyze GitHub PR data and export for analysis",
		Long: `pr-analyzer fetches GitHub pull request data, caches it locally, and exports it in analysis-ready formats.

Examples:
  pr-analyzer microsoft/vscode                    # Export recent 100 PRs to JSONL
  pr-analyzer microsoft/vscode --limit 50        # Export recent 50 PRs  
  pr-analyzer microsoft/vscode --format csv      # Export to CSV format
  pr-analyzer microsoft/vscode --all             # Export all PRs
  pr-analyzer microsoft/vscode --refetch         # Force refresh cache`,
		Args: cobra.ExactArgs(1),
		RunE: runAnalyze,
	}

	// Add flags
	rootCmd.Flags().String("format", "jsonl", "Export format: jsonl, csv")
	rootCmd.Flags().Int("limit", 100, "Fetch recent N PRs (use --all for unlimited)")
	rootCmd.Flags().Bool("all", false, "Fetch all PRs (overrides --limit)")
	rootCmd.Flags().Bool("include-diffs", false, "Include file diffs in export")
	rootCmd.Flags().Bool("refetch", false, "Force refetch all data (ignore cache)")
	rootCmd.Flags().String("since", "", "Fetch PRs updated since date (YYYY-MM-DD)")
	rootCmd.Flags().Int("pr", 0, "Fetch specific PR number only")
	rootCmd.Flags().String("output", "", "Custom output filename")

	// Add version as subcommand
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	// Parse repository argument
	repo := args[0]

	// Get flags
	format, _ := cmd.Flags().GetString("format")
	limit, _ := cmd.Flags().GetInt("limit")
	all, _ := cmd.Flags().GetBool("all")
	includeDiffs, _ := cmd.Flags().GetBool("include-diffs")
	refetch, _ := cmd.Flags().GetBool("refetch")
	since, _ := cmd.Flags().GetString("since")
	prNumber, _ := cmd.Flags().GetInt("pr")
	output, _ := cmd.Flags().GetString("output")

	// Create analyzer service
	service, err := analyzer.NewService()
	if err != nil {
		return fmt.Errorf("initializing analyzer: %w", err)
	}
	defer service.Close()

	// Create analyze options
	opts := analyzer.AnalyzeOptions{
		Repo:         repo,
		Format:       format,
		Limit:        limit,
		All:          all,
		IncludeDiffs: includeDiffs,
		Refetch:      refetch,
		Since:        since,
		PRNumber:     prNumber,
		Output:       output,
	}

	// Run analysis
	ctx := context.Background()
	return service.Analyze(ctx, opts)
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("pr-analyzer version %s (commit: %s, built: %s)\n", version, commit, date)
		},
	}
}
