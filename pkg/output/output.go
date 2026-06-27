// Package output handles rendering analysis results.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"diffscope/pkg/analysis"
	"diffscope/pkg/scoring"
)

// Format represents the output format.
type Format string

const (
	FormatText  Format = "text"
	FormatJSON  Format = "json"
	FormatMarkdown Format = "markdown"
)

// Renderer renders analysis results in a specific format.
type Renderer struct {
	format Format
	writer io.Writer
}

// NewRenderer creates a new Renderer.
func NewRenderer(format Format, w io.Writer) *Renderer {
	return &Renderer{format: format, writer: w}
}

// Render outputs the analysis result.
func (r *Renderer) Render(result *analysis.ImpactResult, score *scoring.Score) error {
	switch r.format {
	case FormatJSON:
		return r.renderJSON(result, score)
	case FormatMarkdown:
		return r.renderMarkdown(result, score)
	default:
		return r.renderText(result, score)
	}
}

// renderText renders the result as human-readable text.
func (r *Renderer) renderText(result *analysis.ImpactResult, score *scoring.Score) error {
	w := r.writer

	fmt.Fprintln(w, "=== DiffScope Impact Analysis ===")
	fmt.Fprintln(w)

	// Summary
	fmt.Fprintf(w, "Files changed: %d\n", result.Summary.TotalFiles)
	fmt.Fprintf(w, "Lines added:   %d\n", result.Summary.TotalAdditions)
	fmt.Fprintf(w, "Lines deleted: %d\n", result.Summary.TotalDeletions)
	fmt.Fprintf(w, "Total changes: %d\n", result.Summary.TotalChanges)
	fmt.Fprintln(w)

	// Score
	if score != nil {
		fmt.Fprintf(w, "Impact Score: %d/100 (%s)\n", score.Overall, score.Severity)
		fmt.Fprintln(w)
	}

	// Findings
	if len(result.Findings) > 0 {
		fmt.Fprintln(w, "--- Findings ---")
		for _, f := range result.Findings {
			fmt.Fprintf(w, "  [%s] %s:%d - %s\n",
				strings.ToUpper(f.Severity), f.File, f.Line, f.Message)
		}
		fmt.Fprintln(w)
	}

	// Per-file details
	fmt.Fprintln(w, "--- Files ---")
	for _, fi := range result.Files {
		fmt.Fprintf(w, "  %s (%s, +%d -%d)\n",
			fi.Path, fi.ChangeType, fi.Additions, fi.Deletions)

		if fi.TestFile {
			fmt.Fprint(w, "    [test]\n")
		}
		if fi.HasBreaking {
			fmt.Fprint(w, "    [breaking]\n")
		}
		if fi.HasSecurity {
			fmt.Fprint(w, "    [security]\n")
		}
		if fi.HasImports {
			fmt.Fprint(w, "    [imports]\n")
		}

		for _, fn := range fi.Functions {
			if fn.IsNew {
				fmt.Fprintf(w, "    + func %s (line %d)\n", fn.Name, fn.Line)
			} else if fn.IsRemoved {
				fmt.Fprintf(w, "    - func %s (line %d)\n", fn.Name, fn.Line)
			} else {
				fmt.Fprintf(w, "    ~ func %s (line %d, +%d -%d)\n", fn.Name, fn.Line, fn.Added, fn.Deleted)
			}
		}
	}

	return nil
}

// renderJSON renders the result as JSON.
func (r *Renderer) renderJSON(result *analysis.ImpactResult, score *scoring.Score) error {
	output := struct {
		Summary analysis.Summary `json:"summary"`
		Files   []*analysis.FileImpact `json:"files"`
		Findings []analysis.Finding `json:"findings"`
		Score   *scoring.Score `json:"score,omitempty"`
	}{
		Summary:  result.Summary,
		Files:    result.Files,
		Findings: result.Findings,
		Score:    score,
	}

	encoder := json.NewEncoder(r.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// renderMarkdown renders the result as Markdown.
func (r *Renderer) renderMarkdown(result *analysis.ImpactResult, score *scoring.Score) error {
	w := r.writer

	fmt.Fprintln(w, "# DiffScope Impact Analysis")
	fmt.Fprintln(w)

	// Summary table
	fmt.Fprintln(w, "## Summary")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| Metric | Value |")
	fmt.Fprintln(w, "|--------|-------|")
	fmt.Fprintf(w, "| Files changed | %d |\n", result.Summary.TotalFiles)
	fmt.Fprintf(w, "| Lines added | %d |\n", result.Summary.TotalAdditions)
	fmt.Fprintf(w, "| Lines deleted | %d |\n", result.Summary.TotalDeletions)
	fmt.Fprintf(w, "| Total changes | %d |\n", result.Summary.TotalChanges)
	fmt.Fprintln(w)

	// Score
	if score != nil {
		fmt.Fprintf(w, "**Impact Score: %d/100 (%s)**\n\n", score.Overall, score.Severity)
	}

	// Findings
	if len(result.Findings) > 0 {
		fmt.Fprintln(w, "## Findings")
		fmt.Fprintln(w)
		for _, f := range result.Findings {
			fmt.Fprintf(w, "- **[%s]** %s:%d — %s\n",
				strings.ToUpper(f.Severity), f.File, f.Line, f.Message)
		}
		fmt.Fprintln(w)
	}

	// Files
	fmt.Fprintln(w, "## Files")
	fmt.Fprintln(w)
	for _, fi := range result.Files {
		fmt.Fprintf(w, "### %s\n\n", fi.Path)
		fmt.Fprintf(w, "- **Type:** %s\n", fi.ChangeType)
		fmt.Fprintf(w, "- **Added:** %d | **Deleted:** %d\n", fi.Additions, fi.Deletions)

		if fi.TestFile {
			fmt.Fprint(w, "- **Test file**\n")
		}
		if fi.HasBreaking {
			fmt.Fprint(w, "- **Breaking change**\n")
		}
		if fi.HasSecurity {
			fmt.Fprint(w, "- **Security issue**\n")
		}
		if fi.HasImports {
			fmt.Fprint(w, "- **Import changes**\n")
		}
		fmt.Fprintln(w)

		if len(fi.Functions) > 0 {
			fmt.Fprintln(w, "| Function | Line | Added | Deleted |")
			fmt.Fprintln(w, "|----------|------|-------|---------|")
			for _, fn := range fi.Functions {
				status := "~"
				if fn.IsNew {
					status = "+"
				} else if fn.IsRemoved {
					status = "-"
				}
				fmt.Fprintf(w, "| %s %s | %d | %d | %d |\n",
					status, fn.Name, fn.Line, fn.Added, fn.Deleted)
			}
			fmt.Fprintln(w)
		}
	}

	return nil
}
