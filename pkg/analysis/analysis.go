// Package analysis extracts semantic impact from git diffs.
package analysis

import (
	"diffscope/pkg/diff"
	"regexp"
	"strings"
)

// Impact indicates the severity of a change.
type Impact int

const (
	Critical Impact = iota
	High
	Medium
	Low
	Info
)

func (i Impact) String() string {
	switch i {
	case Critical:
		return "critical"
	case High:
		return "high"
	case Medium:
		return "medium"
	case Low:
		return "low"
	case Info:
		return "info"
	default:
		return "unknown"
	}
}

// Finding represents a single analysis finding.
type Finding struct {
	File     string   `json:"file"`
	Function string   `json:"function,omitempty"`
	Line     int      `json:"line"`
	Type     string   `json:"type"`
	Impact   Impact   `json:"impact"`
	Message  string   `json:"message"`
	Severity string   `json:"severity"`
	Rules    []string `json:"rules,omitempty"`
}

// ImpactResult holds the complete analysis of a diff.
type ImpactResult struct {
	Files    []*FileImpact `json:"files"`
	Findings []Finding     `json:"findings"`
	Summary  Summary       `json:"summary"`
}

// Summary provides a high-level overview.
type Summary struct {
	TotalFiles     int  `json:"total_files"`
	TotalAdditions int  `json:"total_additions"`
	TotalDeletions int  `json:"total_deletions"`
	TotalChanges   int  `json:"total_changes"`
	CriticalCount  int  `json:"critical_count"`
	HighCount      int  `json:"high_count"`
	MediumCount    int  `json:"medium_count"`
	LowCount       int  `json:"low_count"`
	InfoCount      int  `json:"info_count"`
	HasBreaking    bool `json:"has_breaking_changes"`
	HasSecurity    bool `json:"has_security_issues"`
	HasTestFiles   bool `json:"has_test_files"`
}

// FileImpact represents the impact analysis for a single file.
type FileImpact struct {
	Path        string     `json:"path"`
	ChangeType  string     `json:"change_type"`
	Additions   int        `json:"additions"`
	Deletions   int        `json:"deletions"`
	Functions   []FuncInfo `json:"functions"`
	TestFile    bool       `json:"test_file"`
	HasBreaking bool       `json:"has_breaking_change"`
	HasSecurity bool       `json:"has_security_issue"`
	HasImports  bool       `json:"has_import_changes"`
	HasConfig   bool       `json:"has_config_change"`
	HasDocs     bool       `json:"has_docs_change"`
}

// FuncInfo represents information about a changed function.
type FuncInfo struct {
	Name      string `json:"name"`
	Line      int    `json:"line"`
	Added     int    `json:"added"`
	Deleted   int    `json:"deleted"`
	IsNew     bool   `json:"is_new"`
	IsRemoved bool   `json:"is_removed"`
}

// Analyzer analyzes diffs for impact.
type Analyzer struct {
	// Language-specific analyzers
	goAnalyzer *GoAnalyzer
	pyAnalyzer *PythonAnalyzer
	tsAnalyzer *TypeScriptAnalyzer
}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		goAnalyzer: &GoAnalyzer{},
		pyAnalyzer: &PythonAnalyzer{},
		tsAnalyzer: &TypeScriptAnalyzer{},
	}
}

// Analyze performs impact analysis on a diff.
func (a *Analyzer) Analyze(d *diff.Diff) *ImpactResult {
	result := &ImpactResult{
		Findings: make([]Finding, 0),
		Files:    make([]*FileImpact, 0),
	}

	for _, fd := range d.Files {
		fi := a.analyzeFile(fd)
		result.Files = append(result.Files, fi)

		// Run security checks
		securityFindings := a.checkSecurity(fd)
		result.Findings = append(result.Findings, securityFindings...)

		// Run breaking change detection
		breakingFindings := a.checkBreakingChanges(fd)
		result.Findings = append(result.Findings, breakingFindings...)
	}

	// Build summary
	result.Summary = a.BuildSummary(result)

	return result
}

// analyzeFile analyzes a single file diff.
func (a *Analyzer) analyzeFile(fd *diff.FileDiff) *FileImpact {
	fi := &FileImpact{
		Path:       fd.NewPath,
		ChangeType: fd.ChangeType.String(),
		Additions:  fd.Added,
		Deletions:  fd.Deleted,
		Functions:  make([]FuncInfo, 0),
	}

	if fd.OldPath != "" {
		fi.Path = fd.OldPath
	}

	// Detect file type
	fi.TestFile = isTestFile(fd.NewPath)
	fi.HasImports = hasImportChanges(fd)
	fi.HasConfig = isConfigFile(fd.NewPath)
	fi.HasDocs = isDocFile(fd.NewPath)

	// Analyze functions based on language
	lang := detectLanguage(fd.NewPath)
	switch lang {
	case "go":
		fi.Functions = a.goAnalyzer.AnalyzeFile(fd)
	case "python":
		fi.Functions = a.pyAnalyzer.AnalyzeFile(fd)
	case "typescript", "javascript":
		fi.Functions = a.tsAnalyzer.AnalyzeFile(fd)
	}

	// Detect breaking changes
	for _, f := range fi.Functions {
		if f.IsRemoved {
			fi.HasBreaking = true
		}
		if f.IsNew && (fi.ChangeType == "modified") {
			// New function in existing file could be breaking if it changes signatures
			fi.HasBreaking = false
		}
	}

	// Check for security issues in the file
	fi.HasSecurity = a.hasSecurityIssue(fd)

	return fi
}

// checkSecurity checks for security-sensitive changes.
func (a *Analyzer) checkSecurity(fd *diff.FileDiff) []Finding {
	var findings []Finding

	for _, l := range fd.Lines {
		if l.Prefix != "+" {
			continue
		}

		content := l.Content

		// Check for hardcoded secrets
		if isHardcodedSecret(content) {
			findings = append(findings, Finding{
				File:     fd.NewPath,
				Line:     l.LineNumber,
				Type:     "security",
				Impact:   High,
				Severity: "high",
				Message:  "Potential hardcoded secret detected",
				Rules:    []string{"hardcoded-secret"},
			})
		}

		// Check for eval/exec patterns
		if containsPattern(content, `eval\(`) || containsPattern(content, `exec\(`) {
			findings = append(findings, Finding{
				File:     fd.NewPath,
				Line:     l.LineNumber,
				Type:     "security",
				Impact:   Critical,
				Severity: "critical",
				Message:  "Use of eval/exec detected",
				Rules:    []string{"unsafe-eval"},
			})
		}

		// Check for SQL injection patterns
		if containsPattern(content, `SQL\s*\(`) || containsPattern(content, `query\s*\(`) {
			findings = append(findings, Finding{
				File:     fd.NewPath,
				Line:     l.LineNumber,
				Type:     "security",
				Impact:   High,
				Severity: "high",
				Message:  "Potential SQL query — ensure parameterized queries are used",
				Rules:    []string{"sql-injection-risk"},
			})
		}
	}

	return findings
}

// checkBreakingChanges checks for breaking changes.
func (a *Analyzer) checkBreakingChanges(fd *diff.FileDiff) []Finding {
	var findings []Finding

	for _, l := range fd.Lines {
		if l.Prefix != "-" {
			continue
		}

		content := l.Content

		// Check for removed exported functions/types
		if isExportedSymbol(content) {
			findings = append(findings, Finding{
				File:     fd.NewPath,
				Line:     l.LineNumber,
				Type:     "breaking",
				Impact:   High,
				Severity: "high",
				Message:  "Removed exported symbol — potential breaking change",
				Rules:    []string{"removed-export"},
			})
		}

		// Check for removed interface methods
		if containsPattern(content, `interface\s*\{`) || containsPattern(content, `interface\s*\{`) {
			findings = append(findings, Finding{
				File:     fd.NewPath,
				Line:     l.LineNumber,
				Type:     "breaking",
				Impact:   Critical,
				Severity: "critical",
				Message:  "Interface definition changed — may break implementations",
				Rules:    []string{"interface-change"},
			})
		}
	}

	return findings
}

// BuildSummary creates a summary from the analysis result.
func (a *Analyzer) BuildSummary(result *ImpactResult) Summary {
	s := Summary{
		TotalFiles:     len(result.Files),
		TotalAdditions: 0,
		TotalDeletions: 0,
	}

	for _, fi := range result.Files {
		s.TotalAdditions += fi.Additions
		s.TotalDeletions += fi.Deletions
		if fi.HasBreaking {
			s.HasBreaking = true
		}
		if fi.HasSecurity {
			s.HasSecurity = true
		}
		if fi.TestFile {
			s.HasTestFiles = true
		}
	}
	s.TotalChanges = s.TotalAdditions + s.TotalDeletions

	for _, f := range result.Findings {
		switch f.Impact {
		case Critical:
			s.CriticalCount++
		case High:
			s.HighCount++
		case Medium:
			s.MediumCount++
		case Low:
			s.LowCount++
		case Info:
			s.InfoCount++
		}
	}

	return s
}

// isTestFile checks if a file path looks like a test file.
func isTestFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, "_test.go") ||
		strings.HasSuffix(lower, "_test.py") ||
		strings.HasSuffix(lower, ".test.") ||
		strings.HasSuffix(lower, ".spec.") ||
		strings.Contains(lower, "/test/") ||
		strings.Contains(lower, "/tests/") ||
		strings.Contains(lower, "/__tests__/")
}

// isConfigFile checks if a file is a configuration file.
func isConfigFile(path string) bool {
	lower := strings.ToLower(path)
	configExts := []string{".yaml", ".yml", ".toml", ".json", ".ini", ".cfg", ".conf", ".env"}
	for _, ext := range configExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// isDocFile checks if a file is documentation.
func isDocFile(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".md") ||
		strings.HasSuffix(lower, ".rst") ||
		strings.HasSuffix(lower, ".txt") ||
		strings.Contains(lower, "/docs/") ||
		strings.Contains(lower, "/doc/")
}

// detectLanguage detects the programming language from a file path.
func detectLanguage(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".go"):
		return "go"
	case strings.HasSuffix(lower, ".py"):
		return "python"
	case strings.HasSuffix(lower, ".ts") || strings.HasSuffix(lower, ".tsx"):
		return "typescript"
	case strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".jsx"):
		return "javascript"
	case strings.HasSuffix(lower, ".rs"):
		return "rust"
	default:
		return "unknown"
	}
}

// containsPattern checks if a string matches a simple pattern.
func containsPattern(s, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// isHardcodedSecret checks for potential hardcoded secrets.
func isHardcodedSecret(line string) bool {
	secretPatterns := []string{
		`password\s*=\s*["'][^"']+["']`,
		`secret\s*=\s*["'][^"']+["']`,
		`api_key\s*=\s*["'][^"']+["']`,
		`token\s*=\s*["'][^"']+["']`,
		`api_secret\s*=\s*["'][^"']+["']`,
	}
	for _, p := range secretPatterns {
		if containsPattern(line, p) {
			return true
		}
	}
	return false
}

// isExportedSymbol checks if a line references an exported symbol (Go convention).
func isExportedSymbol(line string) bool {
	// Check for exported function, type, or variable removals
	return containsPattern(line, `\bfunc\s+[A-Z]`) ||
		containsPattern(line, `\btype\s+[A-Z]`) ||
		containsPattern(line, `\bvar\s+[A-Z]`) ||
		containsPattern(line, `\bconst\s+[A-Z]`)
}

// hasSecurityIssue checks if a file has security-related changes.
func (a *Analyzer) hasSecurityIssue(fd *diff.FileDiff) bool {
	for _, l := range fd.Lines {
		if l.Prefix != "+" {
			continue
		}
		if isHardcodedSecret(l.Content) {
			return true
		}
		if containsPattern(l.Content, `eval\(`) || containsPattern(l.Content, `exec\(`) {
			return true
		}
	}
	return false
}
