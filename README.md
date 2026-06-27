# DiffScope

> Analyze git diff impact and scope before you commit.

DiffScope is a lightweight CLI tool that analyzes git diffs to determine the impact and scope of code changes. It identifies affected functions, detects security-sensitive changes, flags breaking changes, and produces an impact score — all before you commit.

Designed for developers and AI coding agents who want to understand the blast radius of their changes.

## Features

- **Diff Parsing** — Full unified diff parser supporting additions, modifications, deletions, and renames
- **Function-Level Analysis** — Detects added, removed, and modified functions in Go, Python, and TypeScript
- **Security Detection** — Identifies hardcoded secrets, eval/exec usage, and SQL injection patterns in diffs
- **Breaking Change Detection** — Flags removed exported symbols and interface changes
- **Impact Scoring** — Weighted scoring algorithm (0–100) based on size, breaking changes, security, test coverage, and imports
- **Multiple Output Formats** — Human-readable text, JSON, and Markdown
- **Configurable** — TOML config with thresholds, timeouts, and language settings
- **Piped Input** — Pipe `git diff` directly into DiffScope via stdin

## Installation

### From Source

```bash
go install github.com/EdgarOrtegaRamirez/diffscope@latest
```

### Build from Source

```bash
git clone https://github.com/EdgarOrtegaRamirez/diffscope.git
cd diffscope
go build -o diffscope ./cmd/
```

## Quick Start

### Analyze a Diff File

```bash
diffscope run --diff diff.patch
```

### Pipe from Git

```bash
git diff HEAD~1 | diffscope run
git diff --cached | diffscope run
```

### JSON Output

```bash
diffscope run --diff diff.patch --format json
```

### Markdown Report

```bash
diffscope run --diff diff.patch --format markdown
```

### List Analysis Rules

```bash
diffscope rules
```

### Generate Sample Config

```bash
diffscope sample-config
# Creates diffscope.toml in current directory
diffscope run --config diffscope.toml --diff diff.patch
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0    | Success, score within threshold |
| 1    | Error (invalid input, file not found, etc.) |
| 2    | Score exceeds threshold |

## Configuration

Generate a sample config:

```bash
diffscope sample-config
```

This creates a `diffscope.toml`:

```toml
[defaults]
  max_diff_lines = 10000
  timeout = 30
  languages = ["go", "python", "typescript"]
  scoring_threshold = 50
  scan_dependents = true
  exclude_patterns = []
```

### Config Options

| Option | Default | Description |
|--------|---------|-------------|
| `max_diff_lines` | 10000 | Maximum diff lines to process |
| `timeout` | 30 | Analysis timeout in seconds |
| `languages` | ["go", "python", "typescript"] | Languages to analyze |
| `scoring_threshold` | 50 | Score threshold for exit code 2 |
| `scan_dependents` | true | Scan dependent files |
| `exclude_patterns` | [] | File patterns to exclude |

## Impact Scoring

DiffScope uses a weighted scoring algorithm:

| Factor | Weight | Description |
|--------|--------|-------------|
| Size | 25% | Lines changed (logarithmic scale, max 40) |
| Breaking | 30% | Removed exports, interface changes |
| Security | 25% | Hardcoded secrets, eval/exec, SQL injection |
| Test Impact | 10% | Function changes without test coverage |
| Imports | 10% | Import modifications |

### Severity Levels

| Score | Severity |
|-------|----------|
| 0–14 | Safe |
| 15–34 | Minor |
| 35–54 | Moderate |
| 55–74 | Significant |
| 75–100 | Critical |

## Architecture

```
cmd/
  main.go          — CLI entry point with subcommands
pkg/
  config/          — Configuration loading and validation
  diff/            — Unified diff parser
  analysis/        — Semantic analysis (functions, security, breaking changes)
  scoring/         — Impact scoring algorithm
  output/          — Text, JSON, and Markdown renderers
```

### Language Analyzers

- **Go** — Regex-based detection of `func` definitions
- **Python** — Regex-based detection of `def` definitions
- **TypeScript/JavaScript** — Detection of `function` declarations and arrow functions

## License

MIT — See [LICENSE](LICENSE) for details.
