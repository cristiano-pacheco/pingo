package cmd

import (
	"fmt"

	"github.com/cristiano-pacheco/pingo/pkg/tracegen"
	"github.com/spf13/cobra"
)

var (
	tracegenPaths      []string
	tracegenPattern    string
	tracegenImportPath string
	tracegenDryRun     bool
	tracegenVerbose    bool
	tracegenRemove     bool
)

// tracegenCmd represents the tracegen command
var tracegenCmd = &cobra.Command{
	Use:   "tracegen",
	Short: "Automatically inject or remove OpenTelemetry tracing code",
	Long: `Tracegen automatically injects OpenTelemetry tracing code into functions.

It scans Go files and adds tracing instrumentation to functions that:
- Have context.Context as the first parameter
- Match the specified pattern (if provided)

Example usage:
  # Inject traces into all usecase files
  pingo tracegen --path ./internal/modules/identity/usecase

  # Inject traces with a specific function pattern
  pingo tracegen --path ./internal --pattern Execute

  # Dry run to see what would be changed
  pingo tracegen --path ./internal --dry-run --verbose

  # Remove existing traces
  pingo tracegen --path ./internal --remove
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(tracegenPaths) == 0 {
			return fmt.Errorf("at least one path must be specified with --path")
		}

		if tracegenRemove {
			return runTraceRemoval()
		}

		return runTraceGeneration()
	},
}

func init() {
	rootCmd.AddCommand(tracegenCmd)

	tracegenCmd.Flags().StringSliceVarP(&tracegenPaths, "path", "p", []string{}, "Path(s) to scan for Go files (required)")
	tracegenCmd.Flags().StringVar(&tracegenPattern, "pattern", "", "Function name pattern to match (e.g., 'Execute')")
	tracegenCmd.Flags().StringVar(&tracegenImportPath, "import", "github.com/cristiano-pacheco/go-otel/trace", "Import path for trace package")
	tracegenCmd.Flags().BoolVar(&tracegenDryRun, "dry-run", false, "Show what would be changed without modifying files")
	tracegenCmd.Flags().BoolVarP(&tracegenVerbose, "verbose", "v", false, "Enable verbose output")
	tracegenCmd.Flags().BoolVar(&tracegenRemove, "remove", false, "Remove existing trace instrumentation")

	tracegenCmd.MarkFlagRequired("path")
}

func runTraceGeneration() error {
	config := tracegen.Config{
		Paths:           tracegenPaths,
		FunctionPattern: tracegenPattern,
		TraceImportPath: tracegenImportPath,
		DryRun:          tracegenDryRun,
		Verbose:         tracegenVerbose,
	}

	generator := tracegen.NewGenerator(config)

	if tracegenVerbose {
		fmt.Println("Starting trace generation...")
		fmt.Printf("Paths: %v\n", tracegenPaths)
		if tracegenPattern != "" {
			fmt.Printf("Pattern: %s\n", tracegenPattern)
		}
		if tracegenDryRun {
			fmt.Println("DRY RUN MODE - No files will be modified")
		}
		fmt.Println()
	}

	if err := generator.Generate(); err != nil {
		return fmt.Errorf("trace generation failed: %w", err)
	}

	if !tracegenDryRun {
		fmt.Println("✓ Trace generation completed successfully")
	}

	return nil
}

func runTraceRemoval() error {
	config := tracegen.RemovalConfig{
		Paths:   tracegenPaths,
		DryRun:  tracegenDryRun,
		Verbose: tracegenVerbose,
	}

	remover := tracegen.NewRemover(config)

	if tracegenVerbose {
		fmt.Println("Starting trace removal...")
		fmt.Printf("Paths: %v\n", tracegenPaths)
		if tracegenDryRun {
			fmt.Println("DRY RUN MODE - No files will be modified")
		}
		fmt.Println()
	}

	if err := remover.Remove(); err != nil {
		return fmt.Errorf("trace removal failed: %w", err)
	}

	if !tracegenDryRun {
		fmt.Println("✓ Trace removal completed successfully")
	}

	return nil
}
