// Package analysis_test tests impact analysis.
package analysis_test

import (
	"testing"

	"diffscope/pkg/analysis"
	"diffscope/pkg/diff"
)

func TestAnalyzer(t *testing.T) {
	a := analysis.NewAnalyzer()

	d, err := diff.ParseDiff(sampleDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	d.Analyze()

	result := a.Analyze(d)
	if len(result.Files) == 0 {
		t.Error("expected at least one file in result")
	}
	if result.Summary.TotalFiles == 0 {
		t.Error("expected total files > 0")
	}
}

func TestBuildSummary(t *testing.T) {
	a := analysis.NewAnalyzer()
	result := &analysis.ImpactResult{
		Files: []*analysis.FileImpact{
			{Path: "main.go", Additions: 10, Deletions: 5, HasBreaking: true},
			{Path: "test.go", Additions: 3, Deletions: 0, TestFile: true},
		},
		Findings: []analysis.Finding{
			{Impact: analysis.Critical},
			{Impact: analysis.High},
			{Impact: analysis.Medium},
			{Impact: analysis.Medium},
			{Impact: analysis.Low},
			{Impact: analysis.Info},
		},
	}

	summary := a.BuildSummary(result)
	if summary.TotalFiles != 2 {
		t.Errorf("TotalFiles: got %d, want 2", summary.TotalFiles)
	}
	if summary.TotalAdditions != 13 {
		t.Errorf("TotalAdditions: got %d, want 13", summary.TotalAdditions)
	}
	if summary.TotalDeletions != 5 {
		t.Errorf("TotalDeletions: got %d, want 5", summary.TotalDeletions)
	}
	if !summary.HasBreaking {
		t.Error("expected HasBreaking=true")
	}
	if !summary.HasTestFiles {
		t.Error("expected HasTestFiles=true")
	}
	if summary.CriticalCount != 1 {
		t.Errorf("CriticalCount: got %d, want 1", summary.CriticalCount)
	}
	if summary.MediumCount != 2 {
		t.Errorf("MediumCount: got %d, want 2", summary.MediumCount)
	}
}

func TestImpactString(t *testing.T) {
	tests := []struct {
		impact   analysis.Impact
		expected string
	}{
		{analysis.Critical, "critical"},
		{analysis.High, "high"},
		{analysis.Medium, "medium"},
		{analysis.Low, "low"},
		{analysis.Info, "info"},
		{analysis.Impact(99), "unknown"},
	}

	for _, tt := range tests {
		if tt.impact.String() != tt.expected {
			t.Errorf("Impact(%d).String() = %q, want %q", tt.impact, tt.impact.String(), tt.expected)
		}
	}
}

func TestAnalysisNewFile(t *testing.T) {
	a := analysis.NewAnalyzer()
	d, err := diff.ParseDiff(newFileDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	d.Analyze()

	result := a.Analyze(d)
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}

	fi := result.Files[0]
	if fi.ChangeType != "added" {
		t.Errorf("changeType = %q, want %q", fi.ChangeType, "added")
	}
	if fi.Additions != 3 {
		t.Errorf("Additions = %d, want 3", fi.Additions)
	}
}

func TestAnalysisDeletedFile(t *testing.T) {
	a := analysis.NewAnalyzer()
	d, err := diff.ParseDiff(deletedFileDiff)
	if err != nil {
		t.Fatalf("ParseDiff failed: %v", err)
	}
	d.Analyze()

	result := a.Analyze(d)
	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}

	fi := result.Files[0]
	if fi.ChangeType != "deleted" {
		t.Errorf("changeType = %q, want %q", fi.ChangeType, "deleted")
	}
	if fi.Deletions != 2 {
		t.Errorf("Deletions = %d, want 2", fi.Deletions)
	}
}

var sampleDiff = `diff --git a/pkg/main.go b/pkg/main.go
index abc123..def456 100644
--- a/pkg/main.go
+++ b/pkg/main.go
@@ -10,3 +10,5 @@ package main
 	"fmt"
 	"os"
 )
+
+func NewApp() *App { return &App{} }
`

var newFileDiff = `diff --git a/pkg/new.go b/pkg/new.go
new file mode 100644
index 000000..123456
--- /dev/null
+++ b/pkg/new.go
@@ -0,0 +1,3 @@
+package pkg
+
+func NewThing() {}
`

var deletedFileDiff = `diff --git a/pkg/old.go b/pkg/old.go
deleted file mode 100644
index abcdef..000000
--- a/pkg/old.go
+++ /dev/null
@@ -1,2 +0,0 @@
-package pkg
-func Old() {}
`
