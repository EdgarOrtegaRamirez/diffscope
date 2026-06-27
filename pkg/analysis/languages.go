// Package analysis provides language-specific analyzers.
package analysis

import (
	"regexp"
	"strings"

	"diffscope/pkg/diff"
)

// GoAnalyzer analyzes Go code diffs.
type GoAnalyzer struct{}

// AnalyzeFile extracts function-level impact from a Go file diff.
func (ga *GoAnalyzer) AnalyzeFile(fd *diff.FileDiff) []FuncInfo {
	var funcs []FuncInfo

	// Track function context
	var funcAdded, funcDeleted int
	var inFunc bool

	for _, l := range fd.Lines {
		if l.IsHunk || l.IsHeader || l.IsNewFile || l.IsOldFile {
			continue
		}

		switch l.Prefix {
		case "+":
			if inFunc {
				funcAdded++
			}
			// Detect new function definition
			if match := goFuncRe.FindStringSubmatch(l.Content); match != nil {
				funcs = append(funcs, FuncInfo{
					Name:  match[1],
					Line:  l.LineNumber,
					Added: 1,
					IsNew: true,
				})
				inFunc = true
				funcAdded = 1
			}
		case "-":
			if inFunc {
				funcDeleted++
			}
			// Detect removed function
			if match := goFuncRe.FindStringSubmatch(l.Content); match != nil {
				funcs = append(funcs, FuncInfo{
					Name:      match[1],
					Line:      l.LineNumber,
					Deleted:   1,
					IsRemoved: true,
				})
			}
		case " ":
			// Context line — check if we're inside a function
			if !inFunc {
				if match := goFuncRe.FindStringSubmatch(l.Content); match != nil {
					inFunc = true
				}
			}
		}
	}

	// Update first function with final counts
	if len(funcs) > 0 {
		funcs[0].Added = funcAdded
		funcs[0].Deleted = funcDeleted
	}

	return funcs
}

var goFuncRe = regexp.MustCompile(`^func\s+\*?(\w+)\s*\(`)

// PythonAnalyzer analyzes Python code diffs.
type PythonAnalyzer struct{}

// AnalyzeFile extracts function-level impact from a Python file diff.
func (pa *PythonAnalyzer) AnalyzeFile(fd *diff.FileDiff) []FuncInfo {
	var funcs []FuncInfo

	var inFunc bool
	var funcAdded, funcDeleted int

	for _, l := range fd.Lines {
		if l.IsHunk || l.IsHeader || l.IsNewFile || l.IsOldFile {
			continue
		}

		switch l.Prefix {
		case "+":
			if inFunc {
				funcAdded++
			}
			if match := pyFuncRe.FindStringSubmatch(l.Content); match != nil {
				funcs = append(funcs, FuncInfo{
					Name:  match[1],
					Line:  l.LineNumber,
					Added: 1,
					IsNew: true,
				})
				inFunc = true
				funcAdded = 1
			}
		case "-":
			if inFunc {
				funcDeleted++
			}
			if match := pyFuncRe.FindStringSubmatch(l.Content); match != nil {
				funcs = append(funcs, FuncInfo{
					Name:      match[1],
					Line:      l.LineNumber,
					Deleted:   1,
					IsRemoved: true,
				})
			}
		case " ":
			if !inFunc {
				if match := pyFuncRe.FindStringSubmatch(l.Content); match != nil {
					inFunc = true
				}
			}
		}
	}

	if len(funcs) > 0 {
		funcs[0].Added = funcAdded
		funcs[0].Deleted = funcDeleted
	}

	return funcs
}

var pyFuncRe = regexp.MustCompile(`^def\s+(\w+)\s*\(`)

// TypeScriptAnalyzer analyzes TypeScript/JavaScript code diffs.
type TypeScriptAnalyzer struct{}

// AnalyzeFile extracts function-level impact from a TS/JS file diff.
func (ta *TypeScriptAnalyzer) AnalyzeFile(fd *diff.FileDiff) []FuncInfo {
	var funcs []FuncInfo

	var inFunc bool
	var funcAdded, funcDeleted int

	for _, l := range fd.Lines {
		if l.IsHunk || l.IsHeader || l.IsNewFile || l.IsOldFile {
			continue
		}

		switch l.Prefix {
		case "+":
			if inFunc {
				funcAdded++
			}
			if match := tsFuncRe.FindStringSubmatch(l.Content); match != nil {
				funcs = append(funcs, FuncInfo{
					Name:  match[1],
					Line:  l.LineNumber,
					Added: 1,
					IsNew: true,
				})
				inFunc = true
				funcAdded = 1
			} else if match := tsArrowRe.FindStringSubmatch(l.Content); match != nil {
				funcs = append(funcs, FuncInfo{
					Name:  match[1],
					Line:  l.LineNumber,
					Added: 1,
					IsNew: true,
				})
				inFunc = true
				funcAdded = 1
			}
		case "-":
			if inFunc {
				funcDeleted++
			}
			if match := tsFuncRe.FindStringSubmatch(l.Content); match != nil {
				funcs = append(funcs, FuncInfo{
					Name:      match[1],
					Line:      l.LineNumber,
					Deleted:   1,
					IsRemoved: true,
				})
			} else if match := tsArrowRe.FindStringSubmatch(l.Content); match != nil {
				funcs = append(funcs, FuncInfo{
					Name:      match[1],
					Line:      l.LineNumber,
					Deleted:   1,
					IsRemoved: true,
				})
			}
		case " ":
			if !inFunc {
				if match := tsFuncRe.FindStringSubmatch(l.Content); match != nil {
					inFunc = true
				} else if match := tsArrowRe.FindStringSubmatch(l.Content); match != nil {
					inFunc = true
				}
			}
		}
	}

	if len(funcs) > 0 {
		funcs[0].Added = funcAdded
		funcs[0].Deleted = funcDeleted
	}

	return funcs
}

var tsFuncRe = regexp.MustCompile(`(?:function|export\s+function)\s+(\w+)`)
var tsArrowRe = regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s*)?\(`)

// hasImportChanges checks if a diff contains import changes.
func hasImportChanges(fd *diff.FileDiff) bool {
	for _, l := range fd.Lines {
		if l.Prefix != "+" && l.Prefix != "-" {
			continue
		}
		content := strings.TrimSpace(l.Content)
		if strings.HasPrefix(content, "import ") ||
			strings.HasPrefix(content, "from ") ||
			strings.HasPrefix(content, "require(") {
			return true
		}
	}
	return false
}
