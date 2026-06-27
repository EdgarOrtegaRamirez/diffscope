// Package main implements the DiffScope CLI.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"diffscope/pkg/analysis"
	"diffscope/pkg/config"
	"diffscope/pkg/diff"
	"diffscope/pkg/output"
	"diffscope/pkg/scoring"
)

// Version is set at build time.
var Version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	// Parse CLI flags
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	rulesCmd := flag.NewFlagSet("rules", flag.ExitOnError)
	sampleCmd := flag.NewFlagSet("sample-config", flag.ExitOnError)
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	// Run flags
	runDiff := runCmd.String("diff", "", "Path to diff file (reads stdin if empty)")
	runFormat := runCmd.String("format", "text", "Output format: text, json, markdown")
	runConfig := runCmd.String("config", "", "Path to config file")
	runScore := runCmd.Bool("score", true, "Include impact score")

	// Handle subcommands
	if len(os.Args) < 2 {
		printUsage()
		return 1
	}

	switch os.Args[1] {
	case "run":
		runCmd.Parse(os.Args[2:])
		return cmdRun(*runDiff, *runFormat, *runConfig, *runScore)
	case "rules":
		rulesCmd.Parse(os.Args[2:])
		return cmdRules()
	case "sample-config":
		sampleCmd.Parse(os.Args[2:])
		return cmdSampleConfig()
	case "version", "--version", "-v":
		versionCmd.Parse(os.Args[2:])
		return cmdVersion()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		return 1
	}
}

func cmdRun(diffPath, format, configPath string, includeScore bool) int {
	// Load config
	cfg := config.DefaultConfig()
	if configPath != "" {
		loaded, err := config.LoadFromFile(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			return 1
		}
		cfg.Merge(loaded)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
		return 1
	}

	// Read diff
	var rawDiff string
	if diffPath != "" {
		data, err := os.ReadFile(diffPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading diff: %v\n", err)
			return 1
		}
		rawDiff = string(data)
	} else {
		data, err := readAllStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			return 1
		}
		rawDiff = string(data)
	}

	if rawDiff == "" {
		fmt.Fprintln(os.Stderr, "No diff provided (use --diff flag or pipe to stdin)")
		return 1
	}

	// Parse diff
	d, err := diff.ParseDiff(rawDiff)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing diff: %v\n", err)
		return 1
	}
	d.Analyze()

	// Analyze impact
	analyzer := analysis.NewAnalyzer()
	result := analyzer.Analyze(d)

	// Compute score
	var score *scoring.Score
	if includeScore {
		scorer := scoring.NewScorer(cfg)
		score = scorer.ScoreResult(result)
	}

	// Render output
	renderer := output.NewRenderer(output.Format(format), os.Stdout)
	if err := renderer.Render(result, score); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering output: %v\n", err)
		return 1
	}

	// Return non-zero if score exceeds threshold
	if includeScore && score != nil {
		scorer := scoring.NewScorer(cfg)
		if !scorer.MeetsThreshold(score) {
			return 2
		}
	}

	return 0
}

func cmdRules() int {
	fmt.Println("DiffScope Analysis Rules")
	fmt.Println("========================")
	fmt.Println()
	fmt.Println("Security Rules:")
	fmt.Println("  hardcoded-secret    - Detects hardcoded passwords, API keys, tokens")
	fmt.Println("  unsafe-eval         - Detects use of eval/exec")
	fmt.Println("  sql-injection-risk  - Detects potential SQL injection patterns")
	fmt.Println()
	fmt.Println("Breaking Change Rules:")
	fmt.Println("  removed-export      - Detects removal of exported symbols")
	fmt.Println("  interface-change    - Detects interface definition changes")
	fmt.Println()
	fmt.Println("Impact Rules:")
	fmt.Println("  function-added      - Detects new functions")
	fmt.Println("  function-removed    - Detects removed functions")
	fmt.Println("  import-change       - Detects import modifications")
	fmt.Println("  config-change       - Detects configuration file modifications")
	fmt.Println("  test-file           - Detects test file modifications")
	fmt.Println("  doc-change          - Detects documentation changes")
	return 0
}

func cmdSampleConfig() int {
	cfg := config.DefaultConfig()
	path := filepath.Join(".", "diffscope.toml")
	if err := cfg.SaveTOML(path); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		return 1
	}
	fmt.Printf("Sample config written to %s\n", path)
	fmt.Println("Edit it to customize DiffScope behavior, then use --config flag.")
	return 0
}

func cmdVersion() int {
	fmt.Printf("DiffScope %s\n", Version)
	return 0
}

func printUsage() {
	fmt.Println("DiffScope — Analyze git diff impact and scope")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  diffscope run [flags]          Analyze a diff")
	fmt.Println("  diffscope rules                List analysis rules")
	fmt.Println("  diffscope sample-config        Generate sample config")
	fmt.Println("  diffscope version              Show version")
	fmt.Println()
	fmt.Println("Run flags:")
	fmt.Println("  --diff string    Path to diff file (reads stdin if empty)")
	fmt.Println("  --format string  Output format: text, json, markdown (default \"text\")")
	fmt.Println("  --config string  Path to config file")
	fmt.Println("  --score          Include impact score (default true)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  diffscope run --diff diff.patch")
	fmt.Println("  git diff HEAD~1 | diffscope run")
	fmt.Println("  diffscope run --format json --diff staged.patch")
}

func readAllStdin() ([]byte, error) {
	buf := make([]byte, 0, 1024*1024)
	for {
		b := make([]byte, 64*1024)
		n, err := os.Stdin.Read(b)
		if n > 0 {
			buf = append(buf, b[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" || err.Error() == "io: read/write on closed pipe" {
				break
			}
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return buf, err
		}
	}
	return buf, nil
}
