package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Djanghao/tree-pro/internal"
)

var (
    maxFiles int
    maxDirs  int
    maxLevel int
)

var rootCmd = &cobra.Command{
	Use:   "tree-pro [path]",
	Short: "Print a concise, colored directory tree",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if maxFiles < 0 {
			return fmt.Errorf("--files must be >= 0")
		}
		if maxDirs < 0 {
			return fmt.Errorf("--dirs must be >= 0")
		}
		if maxLevel < 0 {
			return fmt.Errorf("--level must be >= 0")
		}

		target := "."
		if len(args) > 0 {
			target = args[0]
		}
		cleaned := filepath.Clean(target)

		walkerOpts := internal.Options{
			MaxFiles: maxFiles,
			MaxLevel: maxLevel,
		}

		dir, err := internal.Walk(cleaned, walkerOpts)
		if err != nil {
			return err
		}

		label := formatRootLabel(target)
        printerOpts := internal.PrinterOptions{
            Writer:   cmd.OutOrStdout(),
            MaxDirs:  maxDirs,
            UseColor: true,
        }
        return internal.PrintTree(label, dir, printerOpts)
    },
}

// Execute runs the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
    rootCmd.Flags().IntVarP(&maxFiles, "files", "f", 5, "maximum files to display per directory (0 for unlimited)")
    rootCmd.Flags().IntVarP(&maxDirs, "dirs", "d", 1, "maximum identical directories to expand per group (0 for unlimited)")
    rootCmd.Flags().IntVarP(&maxLevel, "level", "L", 0, "maximum recursion depth (0 for unlimited)")
}

func formatRootLabel(input string) string {
	if input == "" {
		input = "."
	}
	cleaned := filepath.Clean(input)
	if cleaned == "." {
		return cleaned
	}

	sep := string(os.PathSeparator)
	if strings.HasSuffix(input, sep) {
		return input
	}
	return cleaned + sep
}
