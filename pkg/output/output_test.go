// Package output_test tests rendering.
package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"diffscope/pkg/analysis"
	"diffscope/pkg/output"
	"diffscope/pkg/scoring"
)

func TestRenderText(t *testing.T) {
	result := &analysis.ImpactResult{
		Summary: analysis.Summary{
			TotalFiles:     1,
			TotalAdditions: 5,
			TotalDeletions: 2,
			TotalChanges:   7,
		},
		Files: []*analysis.FileImpact{
			{Path: "main.go", ChangeType: "modified", Additions: 5, Deletions: 2},
		},
	}
	score := &scoring.Score{Overall: 30, Severity: "minor"}

	var buf bytes.Buffer
	r := output.NewRenderer(output.FormatText, &buf)
	if err := r.Render(result, score); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	text := buf.String()
	if !strings.Contains(text, "DiffScope") {
		t.Error("expected output to contain DiffScope")
	}
	if !strings.Contains(text, "Files changed: 1") {
		t.Error("expected output to contain file count")
	}
	if !strings.Contains(text, "Impact Score: 30/100") {
		t.Error("expected output to contain score")
	}
}

func TestRenderJSON(t *testing.T) {
	result := &analysis.ImpactResult{
		Summary: analysis.Summary{
			TotalFiles: 2,
		},
		Files: []*analysis.FileImpact{
			{Path: "main.go"},
			{Path: "test.go", TestFile: true},
		},
	}
	score := &scoring.Score{Overall: 50}

	var buf bytes.Buffer
	r := output.NewRenderer(output.FormatJSON, &buf)
	if err := r.Render(result, score); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify it's valid JSON
	var data map[string]interface{}
	if err := json.NewDecoder(&buf).Decode(&data); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if data["summary"] == nil {
		t.Error("expected summary in JSON output")
	}
	if data["files"] == nil {
		t.Error("expected files in JSON output")
	}
	if data["score"] == nil {
		t.Error("expected score in JSON output")
	}
}

func TestRenderMarkdown(t *testing.T) {
	result := &analysis.ImpactResult{
		Summary: analysis.Summary{
			TotalFiles: 1,
		},
		Files: []*analysis.FileImpact{
			{Path: "main.go", ChangeType: "modified"},
		},
		Findings: []analysis.Finding{
			{File: "main.go", Line: 10, Severity: "high", Message: "test finding"},
		},
	}
	score := &scoring.Score{Overall: 70, Severity: "significant"}

	var buf bytes.Buffer
	r := output.NewRenderer(output.FormatMarkdown, &buf)
	if err := r.Render(result, score); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	text := buf.String()
	if !strings.Contains(text, "# DiffScope") {
		t.Error("expected markdown heading")
	}
	if !strings.Contains(text, "## Summary") {
		t.Error("expected summary section")
	}
	if !strings.Contains(text, "## Findings") {
		t.Error("expected findings section")
	}
	if !strings.Contains(text, "Impact Score: 70/100") {
		t.Error("expected score in markdown")
	}
}

func TestRenderNoScore(t *testing.T) {
	result := &analysis.ImpactResult{
		Summary: analysis.Summary{
			TotalFiles: 0,
		},
	}

	var buf bytes.Buffer
	r := output.NewRenderer(output.FormatText, &buf)
	if err := r.Render(result, nil); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	text := buf.String()
	if !strings.Contains(text, "DiffScope") {
		t.Error("expected output even without score")
	}
}

func TestRenderEmpty(t *testing.T) {
	result := &analysis.ImpactResult{}

	var buf bytes.Buffer
	r := output.NewRenderer(output.FormatJSON, &buf)
	if err := r.Render(result, nil); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(&buf).Decode(&data); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

func TestRenderJSONNoScore(t *testing.T) {
	result := &analysis.ImpactResult{}

	var buf bytes.Buffer
	r := output.NewRenderer(output.FormatJSON, &buf)
	if err := r.Render(result, nil); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(&buf).Decode(&data); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if data["score"] != nil {
		t.Error("score should be omitted when nil")
	}
}
