# DiffScope — AI Agent Notes

## Project Overview
DiffScope analyzes git diffs to determine the impact and scope of code changes. It identifies affected functions, suggests downstream impact, recommends test files, and produces an impact score.

## Architecture
- **Go 1.23** project, single binary
- Modular pkg layout: `diff/`, `analysis/`, `scoring/`, `output/`, `config/`
- CLI built with standard `flag` package (no heavy framework)
- Diff parsing uses `go-diff` library (unified diff format)
- AST analysis uses `go/ast` and `go/parser` for Go files

## Key Design Decisions
- **No external linter dependencies** — impact analysis is custom-built using AST
- **Language-agnostic diff parsing** — unified diff is standard across all VCS
- **AST-based language analysis** — supports Go, Python, TypeScript (pluggable)
- **Configurable impact scoring** — weighted by file type, change severity, function criticality

## Extension Points
- Add new language analyzers by implementing `analysis.LanguageAnalyzer` interface
- Add new output formats by implementing `output.Format` interface
- Add new scoring factors to `scoring/Algorithm`

## Common Patterns
- Error handling: wrap errors with context using `fmt.Errorf("action: %w", err)`
- Config loading: try default, then env var, then file path
- CLI flags: override config file values when explicitly set

## Testing
- Unit tests: table-driven tests for each pkg
- Integration tests: run `go test ./...` in CI
- Diff fixtures: stored in `tests/fixtures/` as real git diffs

## Security Notes
- Reads diffs from git — validated file paths only
- No network calls by default
- Config file is user-writable, validated before use
